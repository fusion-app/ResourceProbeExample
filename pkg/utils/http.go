package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fusion-app/prober/pkg/base"
)

type HTTPTargetOption struct {
	URL               string
	Method            string
	Headers           HTTPHeaders
	EnableTLSValidate bool
	RetryInterval     time.Duration
}

type HTTPHeaders struct {
	Data map[string]string
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
		headerVal := strings.TrimSpace(headerStr[len(pairs[0])+1:])
		headers.Data[headerKey] = headerVal
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

type HTTPReqHandler struct {
	httpOpt *HTTPTargetOption
	client  *http.Client
}

func NewHTTPReqHandler(timeout time.Duration, httpOpt *HTTPTargetOption) *HTTPReqHandler {
	h := &HTTPReqHandler{
		httpOpt: httpOpt,
	}

	// Create a transport for our use. This is mostly based on
	// http.DefaultTransport with some timeouts changed.
	// TODO(manugarg): Considering cloning DefaultTransport once
	// https://github.com/golang/go/issues/26013 is fixed.
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second, // TCP keep-alive
		DualStack: true,
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         dialer.DialContext,
		MaxIdleConns:        256, // http.DefaultTransport.MaxIdleConns: 100.
		TLSHandshakeTimeout: timeout,
		DisableKeepAlives:   true,
		IdleConnTimeout:     2 * timeout,
	}

	if !httpOpt.EnableTLSValidate {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Clients are safe for concurrent use by multiple goroutines.
	h.client = &http.Client{
		Transport: transport,
	}
	return h
}

func (h *HTTPReqHandler) DoHTTPRequest(req *http.Request) (*base.Result, error) {
	start := time.Now()

	resp, err := h.client.Do(req)
	//defer resp.Body.Close()
	latency := time.Since(start)

	if err != nil {
		if isClientTimeout(err) {
			return nil, fmt.Errorf("timeout error: %+v", err.Error())
		}
		return nil, fmt.Errorf("unknown error: %+v", err.Error())
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode error: %+v", err.Error())
	}

	return &base.Result{
		StartTime:   start,
		Latency:     latency,
		ProbeResult: resBody,
	}, nil
}

func (h *HTTPReqHandler) DoHTTPRequestWithRetry(req *http.Request, retryTimes int) (res *base.Result, err error) {
	for {
		retryTimes--
		res, err = h.DoHTTPRequest(req)
		if err != nil {
			err = fmt.Errorf("URL: %s, http.doHTTPRequest error: %+v", req.URL, err)
		} else {
			return res, err
		}
		if retryTimes < 0 {
			return res, err
		}
		time.Sleep(h.httpOpt.RetryInterval)
	}
}

func (h *HTTPReqHandler) MakeHTTPRequest(query map[string]string, body map[string]interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request body error: %+v", err)
	}

	fullUrl, err := url.Parse(h.httpOpt.URL)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request URL error: %+v", err)
	}
	for k, v := range query {
		fullUrl.Query().Set(k, v)
	}
	fullUrlStr := fullUrl.String()

	req, err := http.NewRequest(h.httpOpt.Method, fullUrlStr, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("target: %s, error creating HTTP request: %+v", fullUrlStr, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	for headerKey, headerVal := range h.httpOpt.Headers.Data {
		req.Header.Set(headerKey, headerVal)
	}

	return req, nil
}
