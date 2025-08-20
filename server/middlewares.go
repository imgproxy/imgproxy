package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
)

const (
	categoryTimeout = "timeout"
)

// WithMetrics wraps RouteHandler with metrics handling.
func (r *Router) WithMetrics(h RouteHandler) RouteHandler {
	if !metrics.Enabled() {
		return h
	}

	return func(reqID string, rw http.ResponseWriter, req *http.Request) error {
		ctx, metricsCancel, rw := metrics.StartRequest(req.Context(), rw, req)
		defer metricsCancel()

		return h(reqID, rw, req.WithContext(ctx))
	}
}

// WithCORS wraps RouteHandler with CORS handling
func (r *Router) WithCORS(h RouteHandler) RouteHandler {
	if len(r.config.CORSAllowOrigin) == 0 {
		return h
	}

	return func(reqID string, rw http.ResponseWriter, req *http.Request) error {
		rw.Header().Set(httpheaders.AccessControlAllowOrigin, r.config.CORSAllowOrigin)
		rw.Header().Set(httpheaders.AccessControlAllowMethods, "GET, OPTIONS")

		return h(reqID, rw, req)
	}
}

// WithSecret wraps RouteHandler with secret handling
func (r *Router) WithSecret(h RouteHandler) RouteHandler {
	if len(r.config.Secret) == 0 {
		return h
	}

	authHeader := fmt.Appendf(nil, "Bearer %s", r.config.Secret)

	return func(reqID string, rw http.ResponseWriter, req *http.Request) error {
		if subtle.ConstantTimeCompare([]byte(req.Header.Get(httpheaders.Authorization)), authHeader) == 1 {
			return h(reqID, rw, req)
		} else {
			return newInvalidSecretError()
		}
	}
}

// WithPanic recovers panic and converts it to normal error
func (r *Router) WithPanic(h RouteHandler) RouteHandler {
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

			retErr = err
		}()

		return h(reqID, rw, r)
	}
}

// WithReportError handles error reporting.
// It should be placed after `WithMetrics`, but before `WithPanic`.
func (r *Router) WithReportError(h RouteHandler) RouteHandler {
	return func(reqID string, rw http.ResponseWriter, req *http.Request) error {
		// Open the error context
		ctx := errorreport.StartRequest(req)
		req = req.WithContext(ctx)
		errorreport.SetMetadata(req, "Request ID", reqID)

		// Call the underlying handler passing the context downwards
		err := h(reqID, rw, req)
		if err == nil {
			return nil
		}

		// Wrap a resulting error into ierrors.Error
		ierr := ierrors.Wrap(err, 0)

		// Get the error category
		errCat := ierr.Category()

		// Exception: any context.DeadlineExceeded error is timeout
		if errors.Is(ierr, context.DeadlineExceeded) {
			errCat = categoryTimeout
		}

		// We do not need to send any canceled context
		if !errors.Is(ierr, context.Canceled) {
			metrics.SendError(ctx, errCat, err)
		}

		// Report error to error collectors
		if ierr.ShouldReport() {
			errorreport.Report(ierr, req)
		}

		// Log response and format the error output
		LogResponse(reqID, req, ierr.StatusCode(), ierr)

		// Error message: either is public message or full development error
		rw.Header().Set(httpheaders.ContentType, "text/plain")
		rw.WriteHeader(ierr.StatusCode())

		if r.config.DevelopmentErrorsMode {
			rw.Write([]byte(ierr.Error()))
		} else {
			rw.Write([]byte(ierr.PublicMessage()))
		}

		return nil
	}
}
