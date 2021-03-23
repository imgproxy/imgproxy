package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	dataDogSpanCtxKey = ctxKey("dataDogSpan")
)

func initDataDog() {
	if !conf.DataDogEnable {
		return
	}

	name := os.Getenv("DD_SERVICE")
	if len(name) == 0 {
		name = "imgproxy"
	}

	tracer.Start(
		tracer.WithService(name),
		tracer.WithServiceVersion(version),
		tracer.WithLogger(dataDogLogger{}),
	)
}

func stopDataDog() {
	tracer.Stop()
}

func startDataDogRootSpan(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	span := tracer.StartSpan(
		"request",
		tracer.Measured(),
		tracer.SpanType("web"),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, r.RequestURI),
	)
	cancel := func() { span.Finish() }
	newRw := dataDogResponseWriter{rw, span}

	return context.WithValue(ctx, dataDogSpanCtxKey, span), cancel, newRw
}

func startDataDogSpan(ctx context.Context, name string) context.CancelFunc {
	rootSpan, _ := ctx.Value(dataDogSpanCtxKey).(tracer.Span)
	span := tracer.StartSpan(name, tracer.Measured(), tracer.ChildOf(rootSpan.Context()))
	return func() { span.Finish() }
}

func sendErrorToDataDog(ctx context.Context, err error) {
	rootSpan, _ := ctx.Value(dataDogSpanCtxKey).(tracer.Span)
	rootSpan.Finish(tracer.WithError(err))
}

func sendTimeoutToDataDog(ctx context.Context, d time.Duration) {
	rootSpan, _ := ctx.Value(dataDogSpanCtxKey).(tracer.Span)
	rootSpan.SetTag("timeout_duration", d)
	rootSpan.Finish(tracer.WithError(errors.New("Timeout")))
}

type dataDogLogger struct {
}

func (l dataDogLogger) Log(msg string) {
	logNotice(msg)
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
