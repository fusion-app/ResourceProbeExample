package main

import (
	"context"
	"flag"
	"github.com/fusion-app/ResourceProbeExample/pkg/parser"
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

	PatchCreator    parser.PatchCreatorSpec
)

func init() {
	flag.StringVar(&MQAddress, "mq-address", "localhost:8082", "Kafka MQ listen address")
	flag.StringVar(&MQTopic, "mq-topic", "event", "Kafka topic to publish event message")

	flag.StringVar(&TargetCRDOption.Kind, "crd-kind", "ServiceResource", "")
	flag.StringVar(&TargetCRDOption.Name, "crd-name", "weather", "")
	flag.StringVar(&TargetCRDOption.Namespace, "crd-namespace", "default", "")
	flag.StringVar(&TargetCRDOption.UID, "crd-uid", "", "")

	flag.DurationVar(&ProbeOption.Interval, "probe-interval", 15 * time.Second, "")
	flag.DurationVar(&ProbeOption.Timeout, "probe-timeout", 3 * time.Second, "")

	flag.StringVar(&EndpointOption.URL, "http-url", "http", "")
	flag.StringVar(&EndpointOption.Method, "http-method", "GET", "")
	flag.BoolVar(&EndpointOption.EnableTLSValidate, "http-tls-validation", true, "")
	flag.DurationVar(&EndpointOption.RetryInterval, "http-retry-interval", time.Second, "")
	flag.Var(&EndpointOption.Headers, "http-headers", "example: 'Accept: */*; Host: localhost:8080'")

	flag.Var(&PatchCreator, "patch-creator", "example: '.weight;/weight;float'")
}

func main() {
	flag.Parse()
	prober := &httpprobe.HTTPProbe{}
	if err := prober.Init("http-probe(weather)", &ProbeOption, &EndpointOption); err != nil {
		log.Fatalf("Probe init error: %+v", err)
	}

	probeResult := make(chan *probe.Result)
	go func() {
		for {
			result, ok := <-probeResult
			if !ok {
				log.Fatalf("Probe result channel has been closed")
			}
			apiParseResult, err := parser.PKUAPIParse(result.ProbeResult)
			if err != nil {
				log.Printf("Parse PKU API error: %+v", err.Error())
				continue
			}
			log.Printf("Parse PKU API result: %s", string(apiParseResult))

			patches := PatchCreator.CreatePatches(apiParseResult)

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
