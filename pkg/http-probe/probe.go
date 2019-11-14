package httpprobe

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fusion-app/ResourceProbeExample/pkg/probe"
)

type HTTPProbe struct {
	name    string
	opt     *probe.Option
	logger  *log.Logger
	httpOpt *HTTPTargetOption
	stats   *HTTPRequestStats

	client  *http.Client
}

type HTTPRequestStats struct {
	RunCount     int64
	SuccessCount int64
	TimeoutCount int64
}

type HTTPHeaders struct {
	Data map[string]string
}

type HTTPTargetOption struct {
	URL               string
	Method            string
	Headers           HTTPHeaders
	EnableTLSValidate bool
}

func (headers *HTTPHeaders) String() string {
	return fmt.Sprintf("http-headers: %d", len(headers.Data))
}

func (headers *HTTPHeaders) Set(value string) error {
	if len(headers.Data) > 0 {
		return fmt.Errorf("The headers is already set ")
	}
	headers.Data = make(map[string]string)

	headersArr := strings.Split(value, ";")
	for _, headerStr := range headersArr {
		pairs := strings.Split(headerStr, ":")
		if len(pairs) < 2 {
			continue
		}
		headerKey := strings.TrimSpace(pairs[0])
		headerVal := strings.TrimSpace(headerStr[len(pairs[0]) + 1:])
		headers.Data[headerKey] = headerVal
	}
	return nil
}

func (p *HTTPProbe) Init(name string, option *probe.Option, targetOption *HTTPTargetOption) error {
	p.name = name
	p.opt = option
	p.httpOpt = targetOption
	p.stats = &HTTPRequestStats{}
	p.logger = log.New(os.Stdout, "http-probe: ", log.LstdFlags)

	// Create a transport for our use. This is mostly based on
	// http.DefaultTransport with some timeouts changed.
	// TODO(manugarg): Considering cloning DefaultTransport once
	// https://github.com/golang/go/issues/26013 is fixed.
	dialer := &net.Dialer{
		Timeout:   p.opt.Timeout,
		KeepAlive: 30 * time.Second, // TCP keep-alive
		DualStack: true,
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         dialer.DialContext,
		MaxIdleConns:        256, // http.DefaultTransport.MaxIdleConns: 100.
		TLSHandshakeTimeout: p.opt.Timeout,
		DisableKeepAlives:   true,
		IdleConnTimeout:     2 * p.opt.Interval,
	}

	if !p.httpOpt.EnableTLSValidate {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Clients are safe for concurrent use by multiple goroutines.
	p.client = &http.Client{
		Transport: transport,
	}

	return nil
}

func isClientTimeout(err error) bool {
	if uerr, ok := err.(*url.Error); ok {
		if nerr, ok := uerr.Err.(net.Error); ok && nerr.Timeout() {
			return true
		}
	}
	return false
}

type RespBody struct {
	Time string   `json:"time"`
	Data DataSpec `json:"data"`
}

type DataSpec struct {
	Temp string  `json:"wendu"`
	PM10 float32 `json:"pm10"`
}

func (p *HTTPProbe) doHTTPRequest(req *http.Request) (*probe.Result, error) {
	start := time.Now()

	resp, err := p.client.Do(req)
	//defer resp.Body.Close()
	latency := time.Since(start)

	if err != nil {
		if isClientTimeout(err) {
			p.stats.TimeoutCount++
			return nil, fmt.Errorf("timeout error: %+v", err.Error())
		}
		return nil, fmt.Errorf("unknown error: %+v", err.Error())
	} else {
		p.stats.SuccessCount++
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode error: %+v", err.Error())
	}

	return &probe.Result{
		StartTime: start,
		Latency: latency,
		ProbeResult: resBody,
	}, nil
}

func (p *HTTPProbe) doHTTPRequestWithRetry(req *http.Request, retryTimes int) *probe.Result {
	for {
		retryTimes--
		res, err := p.doHTTPRequest(req)
		if err != nil {
			p.logger.Printf("URL: %s, http.doHTTPRequest error: %+v", req.URL, err)
		} else {
			return res
		}
		if retryTimes < 0 {
			return res
		}
	}
}

func (p *HTTPProbe) makeHTTPRequest() *http.Request {
	body, err := json.Marshal(map[string]string{})
	if err != nil {
		p.logger.Printf("Create HTTP request body error: %+v", err)
	}

	req, err := http.NewRequest(p.httpOpt.Method, p.httpOpt.URL, bytes.NewBuffer(body))
	if err != nil {
		p.logger.Printf("target: %s, error creating HTTP request: %+v", p.httpOpt.URL, err.Error())
		return nil
	}
	req.Header.Set("Content-type", "application/json")
	for headerKey, headerVal := range p.httpOpt.Headers.Data {
		req.Header.Set(headerKey, headerVal)
		p.logger.Printf("Client set header %s = %s", headerKey, headerVal)
	}

	return req
}

func (p *HTTPProbe) Start(ctx context.Context, resultChan chan<- *probe.Result) {
	ticker := time.NewTicker(p.opt.Interval)
	defer ticker.Stop()

	go func() {
		<-ctx.Done()
		close(resultChan)
	}()

	for range ticker.C {
		select {
		case <-ctx.Done():
			break
		default:
		}
		p.stats.RunCount++
		p.logger.Printf("Start %d-th probe", p.stats.RunCount)

		reqCtx, _ := context.WithTimeout(ctx, p.opt.Timeout)
		go func() {
			req := p.makeHTTPRequest()
			result := p.doHTTPRequestWithRetry(req.WithContext(reqCtx), 3)
			if result != nil {
				resultChan <- result
			}
		}()
	}
}
