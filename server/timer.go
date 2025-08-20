// timer.go contains methods for storing, retrieving and checking
// timer in a request context.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

// timerSinceCtxKey represents a context key for start time.
type timerSinceCtxKey struct{}

// startRequestTimer starts a new request timer.
func startRequestTimer(r *http.Request) (*http.Request, context.CancelFunc) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, timerSinceCtxKey{}, time.Now())
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
	return r.WithContext(ctx), cancel
}

// requestStartedAt returns the duration since the timer started in the context.
func requestStartedAt(ctx context.Context) time.Duration {
	if t, ok := ctx.Value(timerSinceCtxKey{}).(time.Time); ok {
		return time.Since(t)
	}
	return 0
}

// CheckTimeout checks if the request context has timed out or cancelled and returns
// wrapped error.
func CheckTimeout(ctx context.Context) error {
	select {
	case <-ctx.Done():
		d := requestStartedAt(ctx)

		err := ctx.Err()
		switch err {
		case context.Canceled:
			return newRequestCancelledError(d)
		case context.DeadlineExceeded:
			return newRequestTimeoutError(d)
		default:
			return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryTimeout))
		}
	default:
		return nil
	}
}
