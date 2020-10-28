package http_probe

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/fusion-app/prober/pkg/base"
	"github.com/fusion-app/prober/pkg/config"
	"github.com/fusion-app/prober/pkg/utils"
)

type HTTPProbe struct {
	name   string
	opt    *base.ProbeOption
	logger *log.Logger
	handler *utils.HTTPReqHandler
}

func NewHTTPProbe(name string, opt *base.ProbeOption, reqCfg *config.HTTPActionSpec) *HTTPProbe {
	h := utils.NewHTTPReqHandler(opt.Timeout, &utils.HTTPTargetOption{
		URL:               reqCfg.URL,
		Method:            reqCfg.Action,
		Headers:           utils.HTTPHeaders{ Data: reqCfg.Headers },
		EnableTLSValidate: reqCfg.ValidateTLS,
		RetryInterval:     opt.Interval,
	})
	return &HTTPProbe{
		name:    name,
		opt:     opt,
		handler: h,
		logger:  log.New(os.Stdout, "[http-probe] ", log.LstdFlags),
	}
}

func (p *HTTPProbe) Start(ctx context.Context, resultChan chan<- *base.Result) {
	ticker := time.NewTicker(p.opt.Interval)
	defer ticker.Stop()

	go func() {
		<-ctx.Done()
		close(resultChan)
	}()

	var runCount int64 = 0
	for range ticker.C {
		select {
		case <-ctx.Done():
			break
		default:
		}
		runCount++
		p.logger.Printf("Start %d-th probe", runCount)

		reqCtx, _ := context.WithTimeout(ctx, p.opt.Timeout)
		req, err := p.handler.MakeHTTPRequest(nil, nil)
		if err != nil {
			p.logger.Printf("build HTTP request err: %v", err)
		}
		result, err := p.handler.DoHTTPRequestWithRetry(req.WithContext(reqCtx), 3)
		if err != nil {
			p.logger.Printf("do HTTP request err: %v", err)
		}
		if result != nil {
			resultChan <- result
		}
	}
}
