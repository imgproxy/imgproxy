package main

import (
	"fmt"
	"time"
)

type timer struct {
	StartTime time.Time
	Timer     <-chan time.Time
}

func startTimer(dt time.Duration) *timer {
	return &timer{time.Now(), time.After(dt)}
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
	return newError(503, fmt.Sprintf("Timeout after %v", t.Since()), "Timeout")
}

func (t *timer) Since() time.Duration {
	return time.Since(t.StartTime)
}
