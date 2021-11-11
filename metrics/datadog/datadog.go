package datadog

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/version"
)

type spanCtxKey struct{}

var enabled bool

func Init() {
	if !config.DataDogEnable {
		return
	}

	name := os.Getenv("DD_SERVICE")
	if len(name) == 0 {
		name = "imgproxy"
	}

	tracer.Start(
		tracer.WithService(name),
		tracer.WithServiceVersion(version.Version()),
		tracer.WithLogger(dataDogLogger{}),
	)
}

func Stop() {
	if enabled {
		tracer.Stop()
	}
}

func StartRootSpan(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !enabled {
		return ctx, func() {}, rw
	}

	span := tracer.StartSpan(
		"request",
		tracer.Measured(),
		tracer.SpanType("web"),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, r.RequestURI),
	)
	cancel := func() { span.Finish() }
	newRw := dataDogResponseWriter{rw, span}

	return context.WithValue(ctx, spanCtxKey{}, span), cancel, newRw
}

func StartSpan(ctx context.Context, name string) context.CancelFunc {
	if !enabled {
		return func() {}
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(tracer.Span); ok {
		span := tracer.StartSpan(name, tracer.Measured(), tracer.ChildOf(rootSpan.Context()))
		return func() { span.Finish() }
	}

	return func() {}
}

func SendError(ctx context.Context, err error) {
	if !enabled {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(tracer.Span); ok {
		rootSpan.Finish(tracer.WithError(err))
	}
}

func SendTimeout(ctx context.Context, d time.Duration) {
	if !enabled {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(tracer.Span); ok {
		rootSpan.SetTag("timeout_duration", d)
		rootSpan.Finish(tracer.WithError(errors.New("Timeout")))
	}
}

type dataDogLogger struct {
}

func (l dataDogLogger) Log(msg string) {
	log.Info(msg)
}

type dataDogResponseWriter struct {
	rw   http.ResponseWriter
	span tracer.Span
}

func (ddrw dataDogResponseWriter) Header() http.Header {
	return ddrw.rw.Header()
}
func (ddrw dataDogResponseWriter) Write(data []byte) (int, error) {
	return ddrw.rw.Write(data)
}
func (ddrw dataDogResponseWriter) WriteHeader(statusCode int) {
	ddrw.span.SetTag(ext.HTTPCode, statusCode)
	ddrw.rw.WriteHeader(statusCode)
}
