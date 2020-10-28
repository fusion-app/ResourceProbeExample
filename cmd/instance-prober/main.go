package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/fusion-app/prober/pkg/base"
	"github.com/fusion-app/prober/pkg/config"
	"github.com/fusion-app/prober/pkg/http-probe"
	"github.com/fusion-app/prober/pkg/mq-hub"
	"github.com/fusion-app/prober/pkg/parser"
	"github.com/fusion-app/prober/pkg/utils"
)

var (
	MQAddress string
	MQTopic   string

	TargetCRDOption mqhub.TargetCRDSpec
	ProbeOption     base.ProbeOption
	EndpointOption  utils.HTTPTargetOption
)

func init() {
	flag.StringVar(&MQAddress, "mq-address", "localhost:8082", "Kafka MQ listen address")
	flag.StringVar(&MQTopic, "mq-topic", "event", "Kafka topic to publish event message")

	flag.StringVar(&TargetCRDOption.Kind, "crd-kind", "ServiceResource", "")
	flag.StringVar(&TargetCRDOption.Name, "crd-name", "weather", "")
	flag.StringVar(&TargetCRDOption.Namespace, "crd-namespace", "default", "")
	flag.StringVar(&TargetCRDOption.UID, "crd-uid", "", "")

	flag.DurationVar(&ProbeOption.Interval, "probe-interval", 6*time.Second, "")
	flag.DurationVar(&ProbeOption.Timeout, "probe-timeout", 3*time.Second, "")

	flag.StringVar(&EndpointOption.URL, "http-url", "http", "")
	flag.DurationVar(&EndpointOption.RetryInterval, "http-retry-interval", time.Second, "")
	flag.Var(&EndpointOption.Headers, "http-headers", "example: 'Accept: */*; Host: localhost:8080'")
}

func main() {
	flag.Parse()

	EndpointOption.URL = fmt.Sprintf("%s?uid=%s/%s", EndpointOption.URL, TargetCRDOption.Namespace, TargetCRDOption.Name)
	EndpointOption.Method = "GET"
	EndpointOption.EnableTLSValidate = false

	httpProbe := http_probe.NewHTTPProbe("http-probe(weather)", &ProbeOption, &config.HTTPActionSpec{
		Action:  "POST",
		URL:     EndpointOption.URL,
		Headers: EndpointOption.Headers.Data,
	})

	probeResult := make(chan *base.Result)
	var preProbeResult []byte
	go func() {
		for {
			result, ok := <-probeResult
			if !ok {
				log.Fatalf("Probe result channel has been closed")
			}
			apiParseResult, err := parser.Parse(parser.Normal, result.ProbeResult)
			if err != nil {
				log.Printf("Parse probe result error: %+v", err.Error())
				continue
			} else if bytes.Equal(preProbeResult, apiParseResult) {
				log.Printf("Probe result not changed, ignore")
				continue
			}
			preProbeResult = apiParseResult
			log.Printf("Parse probe result: %s", string(apiParseResult))

			patches, err := parser.CreateAppInstanceStatusPatches(apiParseResult)

			if err != nil {
				log.Printf("Create patches has error: %+v", err)
			} else if len(patches) == 0 {
				log.Printf("Patches is empty, not Pub Msg")
			} else {
				msg := mqhub.MessageSpec{
					Target:      TargetCRDOption,
					StatusPatch: patches,
					ProbeTime:   result.StartTime,
				}
				err = mqhub.Pub(MQAddress, MQTopic, msg)
				if err != nil {
					log.Printf("Pub msg error: %+v", err.Error())
				}
			}
		}
	}()
	httpProbe.Start(context.Background(), probeResult)
}
