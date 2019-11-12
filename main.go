package main

import (
	"context"
	"flag"
	"github.com/savaki/jq"
	"log"
	"strconv"
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
	ValueType  ValueTypeName
}

type ValueTypeName string

const (
	String ValueTypeName = "string"
	Float ValueTypeName = "float"
	Bool ValueTypeName = "Bool"
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

	selectors := []JSONSelector{
		{ JQSelector: ".data.shidu", PatchPath: "/temperature", ValueType: String },
	}

	probeResult := make(chan *probe.Result)
	go func() {
		for {
			result, ok := <-probeResult
			if !ok {
				break
			}

			var patches []mqhub.PatchItem

			for _, item := range selectors {
				op, err := jq.Parse(item.JQSelector)
				if err != nil {
					log.Fatalf("Invalid jq error: %+v", err.Error())
				}
				valueBytes, err := op.Apply(result.ProbeResult)
				if err != nil {
					log.Fatalf("Not found value in json error: %+v", err.Error())
				}
				var value interface{}
				switch item.ValueType {
				case String:
					if val, err := strconv.Unquote(string(valueBytes)); err == nil {
						value = val
					} else {
						log.Fatalf("Parse string value error: %+v", err.Error())
					}
				case Float:
					if val, err := strconv.ParseFloat(string(valueBytes), 64); err == nil {
						value = val
					} else {
						log.Fatalf("Parse float value error: %+v", err.Error())
					}
				case Bool:
					if val, err := strconv.ParseBool(string(valueBytes)); err == nil {
						value = val
					} else {
						log.Fatalf("Parse boolean value error: %+v", err.Error())
					}
				default:
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
			err := mqhub.Pub(MQAddress, MQTopic, msg)
			if err != nil {
				log.Fatalf("Pub msg error: %+v", err.Error())
			}
		}
	}()
	prober.Start(context.Background(), probeResult)
}
