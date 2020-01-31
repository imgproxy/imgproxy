package main

import (
	"context"
	"fmt"
	"time"
)

var timerSinceCtxKey = ctxKey("timerSince")

func setTimerSince(ctx context.Context) context.Context {
	return context.WithValue(ctx, timerSinceCtxKey, time.Now())
}

func getTimerSince(ctx context.Context) time.Duration {
	return time.Since(ctx.Value(timerSinceCtxKey).(time.Time))
}

func checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		d := getTimerSince(ctx)

		if ctx.Err() != context.DeadlineExceeded {
			panic(newError(499, fmt.Sprintf("Request was cancelled after %v", d), "Cancelled"))
		}

		if newRelicEnabled {
			sendTimeoutToNewRelic(ctx, d)
		}

		if prometheusEnabled {
			incrementPrometheusErrorsTotal("timeout")
		}

		panic(newError(503, fmt.Sprintf("Timeout after %v", d), "Timeout"))
	default:
		// Go ahead
	}
}
