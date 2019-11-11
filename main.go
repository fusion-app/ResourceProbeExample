package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/savaki/jq"
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
				break
			}
			//log.Printf("Probe result: %+v", result)

			op, err := jq.Parse(".data.wendu")
			if err != nil {
				log.Fatalf("Invalid jq error: %+v", err.Error())
			}
			value, err := op.Apply(result.ProbeResult)
			if err != nil {
				log.Fatalf("Not found value in json error: %+v", err.Error())
			}
			log.Printf("Parse value in json by '%s': %s", ".data.wendu", string(value))

			patchStr := fmt.Sprintf(`{"op": "replace", "path": "/temperature", "value": "%s"}`, string(value))
			log.Printf("JSON-Patch string: %s", patchStr)

			msg := mqhub.MessageSpec{
				Target:      TargetCRDOption,
				UpdatePatch: []byte(patchStr),
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
