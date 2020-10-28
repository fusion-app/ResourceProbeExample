package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fusion-app/prober/pkg/base"
	"github.com/fusion-app/prober/pkg/config"
	"github.com/fusion-app/prober/pkg/http-probe"
	"github.com/fusion-app/prober/pkg/http-setter"
	"github.com/fusion-app/prober/pkg/mq-hub"
)

var (
	MQAddress string
	MQTopic   string

	TargetCRDOption mqhub.TargetCRDSpec
	ProbeOption     base.ProbeOption

	PatcherConfigPath string
)

func init() {
	flag.StringVar(&MQAddress, "mq-address", "localhost:8082", "Kafka MQ listen address")
	flag.StringVar(&MQTopic, "mq-topic", "event", "Kafka topic to publish event message")

	flag.StringVar(&TargetCRDOption.Kind, "crd-kind", "ServiceResource", "")
	flag.StringVar(&TargetCRDOption.Name, "crd-name", "weather", "")
	flag.StringVar(&TargetCRDOption.Namespace, "crd-namespace", "default", "")
	flag.StringVar(&TargetCRDOption.UID, "crd-uid", "", "")

	flag.DurationVar(&ProbeOption.Interval, "probe-interval", 15*time.Second, "")
	flag.DurationVar(&ProbeOption.Timeout, "probe-timeout", 3*time.Second, "")

	flag.StringVar(&PatcherConfigPath, "patcher-cfg-path", "", "patcher config file path")
}

func main() {
	flag.Parse()

	patcherCfg, err := config.ParsePatcherConfig(PatcherConfigPath)
	if err != nil {
		log.Fatalf("Setter config (%s) init error: %v", PatcherConfigPath, err)
	}

	for idx, patcher := range patcherCfg.Patchers {
		probe := http_probe.NewHTTPProbe(fmt.Sprintf("probe-%d", idx), &ProbeOption, &patcher.Source)

		var setters []base.Setter
		for _, setterCfg := range patcher.Setters {
			setters = append(setters, http_setter.NewHTTPSetter(
				fmt.Sprintf("setter-%d", idx),
				ProbeOption.Timeout, ProbeOption.Interval,
				&setterCfg))
		}

		worker := base.NewWorker(fmt.Sprintf("worker-%d", idx), probe, setters)
		go worker.Start(context.TODO())
	}
	log.Printf("All patchers has been started")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Printf("\r- Ctrl+C pressed in Terminal")
	os.Exit(0)
}
