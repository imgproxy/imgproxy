package main

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	datadogEnabled = false
)

func initDatadog() (bool, context.CancelFunc) {
	name := conf.DatadogServiceName
	if len(name) == 0 {
		return false, nil
	}

	tracer.Start(tracer.WithServiceName(name))

	datadogEnabled = true

	cancel := func() { tracer.Stop() }
	return true, cancel
}

func startDatadogTransaction(ctx context.Context, r *http.Request) (context.Context, context.CancelFunc) {
	po := getProcessingOptions(ctx)
	resourceName := fmt.Sprintf("%vx%v (%v)", po.Width, po.Height, po.ResizingType)
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeWeb),
		tracer.ResourceName(resourceName),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, r.URL.Path),
	}
	return startDatadogSpan(ctx, "http.request", opts...)
}

func startDatadogSpan(ctx context.Context, name string, opts ...ddtrace.StartSpanOption) (context.Context, context.CancelFunc) {
	span, ctx := tracer.StartSpanFromContext(ctx, name, opts...)
	cancel := func() { span.Finish() }
	return ctx, cancel
}
