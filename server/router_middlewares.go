package server

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
)

// WithMetrics wraps RouteHandler with metrics handling.
func (ro *Router) WithMetrics(h RouteHandler) RouteHandler {
	if !metrics.Enabled() {
		return h
	}

	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		ctx, metricsCancel, rw := metrics.StartRequest(r.Context(), rw, r)
		defer metricsCancel()

		return h(reqID, rw, r.WithContext(ctx))
	}
}

// WithCORS wraps RouteHandler with CORS handling
func (ro *Router) WithCORS(h RouteHandler) RouteHandler {
	if len(ro.config.CORSAllowOrigin) == 0 {
		return h
	}

	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		rw.Header().Set(httpheaders.AccessControlAllowOrigin, ro.config.CORSAllowOrigin)
		rw.Header().Set(httpheaders.AccessControlAllowMethods, "GET, OPTIONS")

		return h(reqID, rw, r)
	}
}

// WithSecret wraps RouteHandler with secret handling
func (ro *Router) WithSecret(h RouteHandler) RouteHandler {
	if len(ro.config.Secret) == 0 {
		return h
	}

	authHeader := fmt.Appendf(nil, "Bearer %s", ro.config.Secret)

	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get(httpheaders.Authorization)), authHeader) == 1 {
			return h(reqID, rw, r)
		} else {
			return newInvalidSecretError()
		}
	}
}

// WithReportError handles error reporting.
// It should be placed after `WithMetrics`, but before `WithPanic`.
func (ro *Router) WithReportError(h RouteHandler) RouteHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		// Open the error context
		ctx := errorreport.StartRequest(r)
		r = r.WithContext(ctx)
		errorreport.SetMetadata(r, "Request ID", reqID)

		// Call the underlying handler passing the context downwards
		err := h(reqID, rw, r)

		if err == nil {
			return nil
		}

		// Wrap a resulting error into ierrors.Error
		ierr := ierrors.Wrap(err, 0)

		// Get the error category
		errCat := ierr.Category()

		// Report error to metrics (if metrics are disabled, it will be a no-op)
		if ierr.StatusCode() != 499 {
			metrics.SendError(ctx, errCat, err)
		}

		// Report error to error collectors
		if ierr.ShouldReport() {
			errorreport.Report(ierr, r)
		}

		// Log response and format the error output
		LogResponse(reqID, r, ierr.StatusCode(), ierr)

		// Error message: either is public message or full development error
		rw.Header().Set(httpheaders.ContentType, "text/plain")
		rw.WriteHeader(ierr.StatusCode())

		if ro.config.DevelopmentErrorsMode {
			rw.Write([]byte(ierr.Error()))
		} else {
			rw.Write([]byte(ierr.PublicMessage()))
		}

		return nil
	}
}

// WithPanic recovers panic and converts it to normal error
func (ro *Router) WithPanic(h RouteHandler) RouteHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) (retErr error) {
		defer func() {
			// try to recover from panic
			rerr := recover()
			if rerr == nil {
				return
			}

			// abort handler is an exception of net/http, we should simply repanic it.
			// it will supress the stack trace
			if rerr == http.ErrAbortHandler {
				panic(rerr)
			}

			// let's recover error value from panic if it has panicked with error
			err, ok := rerr.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", err)
			}

			// wrap ierror unless already done
			ierr := ierrors.Wrap(err, 0)
			if ierr.ShouldReport() {
				errorreport.Report(ierr, r)
			}

			retErr = ierr
		}()

		return h(reqID, rw, r)
	}
}
