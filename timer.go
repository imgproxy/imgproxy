package main

import (
	"fmt"
	"time"
)

type timer struct {
	StartTime time.Time
	Timer     <-chan time.Time
	Info      string
}

func startTimer(dt time.Duration, info string) *timer {
	return &timer{time.Now(), time.After(dt), info}
}

func (t *timer) Check() {
	select {
	case <-t.Timer:
		panic(t.TimeoutErr())
	default:
		// Go ahead
	}
}

func (t *timer) TimeoutErr() imgproxyError {
	return newError(503, fmt.Sprintf("Timeout after %v (%s)", t.Since(), t.Info), "Timeout")
}

func (t *timer) Since() time.Duration {
	return time.Since(t.StartTime)
}
