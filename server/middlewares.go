package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// WithMonitoring wraps RouteHandler with monitoring handling.
func (r *Router) WithMonitoring(h RouteHandler) RouteHandler {
	if !r.monitoring.Enabled() {
		return h
	}

	return func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		ctx, cancel, newRw := r.monitoring.StartRequest(req.Context(), rw.HTTPResponseWriter(), req)
		defer cancel()

		// Replace rw.ResponseWriter with new one returned from monitoring
		rw.SetHTTPResponseWriter(newRw)

		return h(reqID, rw, req.WithContext(ctx))
	}
}

// WithCORS wraps RouteHandler with CORS handling
func (r *Router) WithCORS(h RouteHandler) RouteHandler {
	if len(r.config.CORSAllowOrigin) == 0 {
		return h
	}

	return func(reqID string, rw ResponseWriter, req *http.Request) *Error {
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

	return func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		if subtle.ConstantTimeCompare([]byte(req.Header.Get(httpheaders.Authorization)), authHeader) == 1 {
			return h(reqID, rw, req)
		} else {
			return NewError(newInvalidSecretError(), errCategorySecurity)
		}
	}
}

// WithPanic recovers panic and converts it to normal error
func (r *Router) WithPanic(h RouteHandler) RouteHandler {
	return func(reqID string, rw ResponseWriter, r *http.Request) (retErr *Error) {
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

			retErr = NewError(errctx.Wrap(err, 1), errCategoryUnexpected)
		}()

		return h(reqID, rw, r)
	}
}

// WithReportError handles error reporting.
// It should be placed after `WithMonitoring`, but before `WithPanic`.
func (r *Router) WithReportError(h RouteHandler) RouteHandler {
	return func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		// Open the error context
		ctx := errorreport.StartRequest(req)
		req = req.WithContext(ctx)
		errorreport.SetMetadata(req, "Request ID", reqID)

		// Call the underlying handler passing the context downwards
		err := h(reqID, rw, req)
		if err == nil {
			return nil
		}

		// We do not need to send any canceled context
		if !errors.Is(err.Err, context.Canceled) {
			r.monitoring.SendError(ctx, err.Category, err.Err)
		}

		// Report error to error collectors
		if err.Err.ShouldReport() {
			r.errorReporter.Report(err.Err, req)
		}

		// Log response and format the error output
		LogResponse(reqID, req, err.Err.StatusCode(), err.Err)

		// Error message: either is public message or full development error
		rw.Header().Set(httpheaders.ContentType, "text/plain")
		rw.WriteHeader(err.Err.StatusCode())

		if r.config.DevelopmentErrorsMode {
			rw.Write([]byte(err.Err.Error()))
		} else {
			rw.Write([]byte(err.Err.PublicMessage()))
		}

		return nil
	}
}
