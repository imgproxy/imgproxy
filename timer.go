package main

import (
	"context"
	"fmt"
	"time"
)

var timerSinceCtxKey = ctxKey("timerSince")

type timer struct {
	StartTime time.Time
	Timer     <-chan time.Time
}

func startTimer(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		context.WithValue(context.Background(), timerSinceCtxKey, time.Now()),
		d,
	)
}

func getTimerSince(ctx context.Context) time.Duration {
	return time.Since(ctx.Value(timerSinceCtxKey).(time.Time))
}

func checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		panic(newError(503, fmt.Sprintf("Timeout after %v", getTimerSince(ctx)), "Timeout"))
	default:
		// Go ahead
	}
}
