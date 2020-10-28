package base

import (
	"context"
	"time"
)

type Probe interface {
	Start(context.Context, chan<- *Result)
}

type Result struct {
	StartTime time.Time
	Latency   time.Duration

	ProbeResult []byte
}

type ProbeOption struct {
	Interval time.Duration
	Timeout time.Duration
}