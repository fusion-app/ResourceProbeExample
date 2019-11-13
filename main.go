package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fusion-app/ResourceProbeExample/pkg/http-probe"
	"github.com/fusion-app/ResourceProbeExample/pkg/mq-hub"
	"github.com/fusion-app/ResourceProbeExample/pkg/probe"
)

var (
	MQAddress string
	MQTopic   string

	TargetCRDOption mqhub.TargetCRDSpec
	ProbeOption     probe.Option
	EndpointOption  httpprobe.HTTPTargetOption

	PatchCreator    JSONSelector
)

type JSONSelector struct {
	JQSelector string
	PatchPath  string
	PrevValue  interface{}
	ValueType  httpprobe.ValueTypeName
}

func (selector *JSONSelector) String() string {
	return fmt.Sprintf("selector: %s; path: %s; type: %s", selector.JQSelector, selector.PatchPath, selector.ValueType)
}

func (selector *JSONSelector) Set(value string) error {
	items := strings.Split(value, ";")
	if len(items) < 3 {
		return fmt.Errorf("JSONSelector is invalid")
	}
	selector.JQSelector = strings.TrimSpace(items[0])
	selector.PatchPath = strings.TrimSpace(items[1])
	selector.ValueType = httpprobe.ValueTypeName(strings.TrimSpace(items[2]))
	return nil
}

func init() {
	flag.StringVar(&MQAddress, "mq-address", "localhost:8082", "Kafka MQ listen address")
	flag.StringVar(&MQTopic, "mq-topic", "event", "Kafka topic to publish event message")

	flag.StringVar(&TargetCRDOption.Kind, "crd-kind", "ServiceResource", "")
	flag.StringVar(&TargetCRDOption.Name, "crd-name", "weather", "")
	flag.StringVar(&TargetCRDOption.Namespace, "crd-namespace", "default", "")
	flag.StringVar(&TargetCRDOption.UID, "crd-uid", "", "")

	flag.DurationVar(&ProbeOption.Interval, "probe-interval", 15 * time.Second, "")
	flag.DurationVar(&ProbeOption.Timeout, "probe-timeout", 2 * time.Second, "")

	flag.StringVar(&EndpointOption.URL, "http-url", "http", "")
	flag.StringVar(&EndpointOption.Method, "http-method", "GET", "")
	flag.BoolVar(&EndpointOption.EnableTLSValidate, "http-tls-validation", true, "")
	flag.Var(&EndpointOption.Headers, "http-headers", "example: 'Accept: */*; Host: localhost:8080'")

	flag.Var(&PatchCreator, "patch-creator", "example: '.weight;/weight;float'")
}

func main() {
	flag.Parse()
	prober := &httpprobe.HTTPProbe{}
	if err := prober.Init("http-probe(weather)", &ProbeOption, &EndpointOption); err != nil {
		log.Fatalf("Probe init error: %+v", err)
	}

	selectors := []JSONSelector{
		PatchCreator,
	}

	var isFirstProbe = true
	probeResult := make(chan *probe.Result)
	go func() {
		for {
			result, ok := <-probeResult
			if !ok {
				log.Fatalf("Probe result channel has been closed")
			}
			apiParseResult, err := httpprobe.PKUAPIParse(result.ProbeResult)
			if err != nil {
				log.Printf("Parse PKU API error: %+v", err.Error())
				continue
			}
			log.Printf("Parse PKU API result: %s", string(apiParseResult))

			var patches []mqhub.PatchItem
			for idx, item := range selectors {
				value, err := httpprobe.JQParse(apiParseResult, item.JQSelector, item.ValueType)
				if err != nil || value == nil {
					log.Printf("Not found value in json error by jq '%s': %+v", item.JQSelector, err.Error())
					continue
				}
				log.Printf("Parse value in json by '%s': %v, initValue: %v", item.JQSelector, value, item.PrevValue)
				if isFirstProbe {
					selectors[idx].PrevValue = value
				} else if item.PrevValue != value {
					selectors[idx].PrevValue = value
					patches = append(patches, mqhub.PatchItem{
						Op: mqhub.Replace,
						Path: item.PatchPath,
						Value: value,
					})
				} else {
					log.Printf("Value('%s') not changed", item.JQSelector)
				}
			}
			isFirstProbe = false

			if len(patches) == 0 {
				log.Printf("Patches is empty, not Pub Msg")
			} else {
				msg := mqhub.MessageSpec{
					Target:      TargetCRDOption,
					UpdatePatch: patches,
					ProbeTime:   result.StartTime,
				}
				err = mqhub.Pub(MQAddress, MQTopic, msg)
				if err != nil {
					log.Printf("Pub msg error: %+v", err.Error())
				}
			}
		}
	}()
	prober.Start(context.Background(), probeResult)
}
