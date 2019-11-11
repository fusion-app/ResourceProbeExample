package probe

import "time"

type Probe interface {
	Init(name string, options map[string]interface{}) error
	Start(resultChan chan<- interface{}) error
}

type Result struct {
	StartTime time.Time
	Latency   time.Duration

	ProbeResult []byte
}

type Option struct {
	Interval time.Duration
	Timeout time.Duration
}