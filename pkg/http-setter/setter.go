package http_setter

import (
	"bytes"
	"log"
	"math"
	"os"
	"reflect"
	"time"

	"github.com/fusion-app/prober/pkg/config"
	"github.com/fusion-app/prober/pkg/parser"
	"github.com/fusion-app/prober/pkg/utils"
)

const (
	float64EqualityThreshold = 1e-9
	valuePlaceholder         = "__VALUE__"
)

type HTTPSetter struct {
	name          string
	handler       *utils.HTTPReqHandler
	logger        *log.Logger
	parseSelector string
	prevValue     interface{}
	valueType     parser.ValueTypeName
}

func NewHTTPSetter(name string, timeout, retryInterval time.Duration, cfg *config.FieldSetter) *HTTPSetter {
	h := utils.NewHTTPReqHandler(timeout, &utils.HTTPTargetOption{
		URL:               cfg.Target.URL,
		Method:            cfg.Target.Action,
		Headers:           utils.HTTPHeaders{Data: cfg.Target.Headers},
		EnableTLSValidate: cfg.Target.ValidateTLS,
		RetryInterval:     retryInterval,
	})
	return &HTTPSetter{
		name:          name,
		handler:       h,
		parseSelector: cfg.Parser,
		prevValue:     nil,
		valueType:     parser.ValueTypeName(cfg.Type),
		logger:        log.New(os.Stdout, "[http-setter] ", log.LstdFlags),
	}
}

func (p *HTTPSetter) Parse(src []byte) (interface{}, error) {
	return parser.JQParse(src, p.parseSelector, p.valueType)
}

func (p *HTTPSetter) SetValue(val interface{}) error {
	if !IsValueChange(p.prevValue, val, p.valueType) {
		p.logger.Printf("value(%s) not change, skip setter", p.prevValue)
		return nil
	}

	req, err := p.handler.MakeHTTPRequest(nil, map[string]interface{}{
		"value": val,
	})
	if err != nil {
		p.logger.Printf("build HTTP request err: %v", err)
	}

	p.logger.Printf("SetValue(%v) to %s", val, req.URL.String())
	_, err = p.handler.DoHTTPRequestWithRetry(req, 3)
	if err == nil {
		p.prevValue = val
	}
	return err
}

func IsValueChange(oldVal, newVal interface{}, valType parser.ValueTypeName) bool {
	if oldVal == nil {
		return true
	} else if newVal == nil {
		return false
	}
	switch valType {
	case parser.String:
		return oldVal.(string) != newVal.(string)
	case parser.Float:
		return math.Abs(oldVal.(float64)-newVal.(float64)) >= float64EqualityThreshold
	case parser.Bool:
		return oldVal.(bool) != newVal.(bool)
	case parser.Int:
		return oldVal.(int) != newVal.(int)
	case parser.Any:
		if reflect.TypeOf(oldVal).Kind() == reflect.Slice && reflect.TypeOf(newVal).Kind() == reflect.Slice {
			return bytes.Compare(oldVal.([]byte), newVal.([]byte)) != 0
		}
		return true
	}
	return true
}
