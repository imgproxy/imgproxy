package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
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

func CheckTimeout(ctx context.Context) error {
	select {
	case <-ctx.Done():
		d := ctxTime(ctx)

		err := ctx.Err()
		switch err {
		case context.Canceled:
			return ierrors.New(499, fmt.Sprintf("Request was cancelled after %v", d), "Cancelled")
		case context.DeadlineExceeded:
			return ierrors.New(http.StatusServiceUnavailable, fmt.Sprintf("Request was timed out after %v", d), "Timeout")
		default:
			return err
		}
	default:
		return nil
	}
}
