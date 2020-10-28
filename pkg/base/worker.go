package base

import (
	"context"
	"log"
	"os"
)

type Worker struct {
	name    string
	probe   Probe
	setters []Setter
	logger  *log.Logger
}

func NewWorker(name string, probe Probe, setters []Setter) *Worker {
	return &Worker{
		name:    name,
		probe:   probe,
		setters: setters,
		logger:  log.New(os.Stdout, "[worker] ", log.LstdFlags),
	}
}

const probeResultMaxCount = 5

func (w *Worker) Start(ctx context.Context) {
	resultChan := make(chan *Result, probeResultMaxCount)
	go func() {
		for {
			result, ok := <-resultChan
			if !ok {
				w.logger.Fatalf("Probe[%s] result channel has been closed", w.name)
			}

			for _, setter := range w.setters {
				parseResult, err := setter.Parse(result.ProbeResult)
				if err != nil {
					w.logger.Printf("Probe[%s] parse error: %+v", w.name, err.Error())
					continue
				}
				if err := setter.SetValue(parseResult); err != nil {
					w.logger.Printf("Probe[%s] patch error: %+v", w.name, err.Error())
				}
			}

			w.logger.Printf("Probe[%s] done", w.name)
		}
	}()
	go w.probe.Start(ctx, resultChan)
}
