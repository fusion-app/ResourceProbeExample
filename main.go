package main

import (
	"context"
	"flag"
	"log"
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
)

type JSONSelector struct {
	JQSelector string
	PatchPath  string
	ValueType  httpprobe.ValueTypeName
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
}

func main() {
	flag.Parse()
	prober := &httpprobe.HTTPProbe{}
	if err := prober.Init("http-probe(weather)", &ProbeOption, &EndpointOption); err != nil {
		log.Fatalf("Probe init error: %+v", err)
	}

	selectors := []JSONSelector{
		{ JQSelector: ".weight", PatchPath: "/weight", ValueType: httpprobe.Float },
	}

	probeResult := make(chan *probe.Result)
	go func() {
		for {
			result, ok := <-probeResult
			if !ok {
				break
			}
			apiParseResult, err := httpprobe.PKUAPIParse(result.ProbeResult)
			if err != nil {
				log.Fatalf("Parse PKU API error: %+v", err.Error())
			}
			log.Printf("Parse PKU API result: %s", string(apiParseResult))

			var patches []mqhub.PatchItem

			for _, item := range selectors {
				value, err := httpprobe.JQParse(apiParseResult, item.JQSelector, item.ValueType)
				if err != nil {
					log.Fatalf("Not found value in json error by js '%s': %+v", item.JQSelector, err.Error())
				}
				log.Printf("Parse value in json by '%s': %v", item.JQSelector, value)
				patches = append(patches, mqhub.PatchItem{
					Op: mqhub.Replace,
					Path: item.PatchPath,
					Value: value,
				})
			}

			msg := mqhub.MessageSpec{
				Target:      TargetCRDOption,
				UpdatePatch: patches,
				ProbeTime:   result.StartTime,
			}
			err = mqhub.Pub(MQAddress, MQTopic, msg)
			if err != nil {
				log.Fatalf("Pub msg error: %+v", err.Error())
			}
		}
	}()
	prober.Start(context.Background(), probeResult)
}
