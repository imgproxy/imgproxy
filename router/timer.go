package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
)

type timerSinceCtxKey = struct{}

func setRequestTime(r *http.Request) *http.Request {
	return r.WithContext(
		context.WithValue(r.Context(), timerSinceCtxKey{}, time.Now()),
	)
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
