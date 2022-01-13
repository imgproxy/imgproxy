package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
)

type timerSinceCtxKey = struct{}

func startRequestTimer(r *http.Request) (*http.Request, context.CancelFunc) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, timerSinceCtxKey{}, time.Now())
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.WriteTimeout)*time.Second)
	return r.WithContext(ctx), cancel
}

func ctxTime(ctx context.Context) time.Duration {
	if t, ok := ctx.Value(timerSinceCtxKey{}).(time.Time); ok {
		return time.Since(t)
	}
	return 0
}

func CheckTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		d := ctxTime(ctx)

		if ctx.Err() != context.DeadlineExceeded {
			panic(ierrors.New(499, fmt.Sprintf("Request was cancelled after %v", d), "Cancelled"))
		}

		metrics.SendTimeout(ctx, d)

		panic(ierrors.New(503, fmt.Sprintf("Timeout after %v", d), "Timeout"))
	default:
		// Go ahead
	}
}
