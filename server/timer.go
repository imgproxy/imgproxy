package server

import (
	"context"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type timerSinceCtxKey struct{}

func startRequestTimer(r *http.Request) (*http.Request, context.CancelFunc) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, timerSinceCtxKey{}, time.Now())
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
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
			return newRequestCancelledError(d)
		case context.DeadlineExceeded:
			return newRequestTimeoutError(d)
		default:
			return ierrors.Wrap(err, 0)
		}
	default:
		return nil
	}
}
