package metrics

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/metrics/cloudwatch"
	"github.com/imgproxy/imgproxy/v3/metrics/datadog"
	"github.com/imgproxy/imgproxy/v3/metrics/newrelic"
	"github.com/imgproxy/imgproxy/v3/metrics/otel"
	"github.com/imgproxy/imgproxy/v3/metrics/prometheus"
)

func Init() error {
	prometheus.Init()

	if err := newrelic.Init(); err != nil {
		return nil
	}

	datadog.Init()

	if err := otel.Init(); err != nil {
		return err
	}

	if err := cloudwatch.Init(); err != nil {
		return err
	}

	return nil
}

func Stop() {
	newrelic.Stop()
	datadog.Stop()
	otel.Stop()
	cloudwatch.Stop()
}

func Enabled() bool {
	return prometheus.Enabled() ||
		newrelic.Enabled() ||
		datadog.Enabled() ||
		otel.Enabled() ||
		cloudwatch.Enabled()
}

func StartRequest(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	promCancel, rw := prometheus.StartRequest(rw)
	ctx, nrCancel, rw := newrelic.StartTransaction(ctx, rw, r)
	ctx, ddCancel, rw := datadog.StartRootSpan(ctx, rw, r)
	ctx, otelCancel, rw := otel.StartRootSpan(ctx, rw, r)

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return ctx, cancel, rw
}

func StartQueueSegment(ctx context.Context) context.CancelFunc {
	promCancel := prometheus.StartQueueSegment()
	nrCancel := newrelic.StartSegment(ctx, "Queue")
	ddCancel := datadog.StartSpan(ctx, "queue")
	otelCancel := otel.StartSpan(ctx, "queue")

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return cancel
}

func StartDownloadingSegment(ctx context.Context) context.CancelFunc {
	promCancel := prometheus.StartDownloadingSegment()
	nrCancel := newrelic.StartSegment(ctx, "Downloading image")
	ddCancel := datadog.StartSpan(ctx, "downloading_image")
	otelCancel := otel.StartSpan(ctx, "downloading_image")

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return cancel
}

func StartProcessingSegment(ctx context.Context) context.CancelFunc {
	promCancel := prometheus.StartProcessingSegment()
	nrCancel := newrelic.StartSegment(ctx, "Processing image")
	ddCancel := datadog.StartSpan(ctx, "processing_image")
	otelCancel := otel.StartSpan(ctx, "processing_image")

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return cancel
}

func StartStreamingSegment(ctx context.Context) context.CancelFunc {
	promCancel := prometheus.StartStreamingSegment()
	nrCancel := newrelic.StartSegment(ctx, "Streaming image")
	ddCancel := datadog.StartSpan(ctx, "streaming_image")
	otelCancel := otel.StartSpan(ctx, "streaming_image")

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return cancel
}

func SendError(ctx context.Context, errType string, err error) {
	prometheus.IncrementErrorsTotal(errType)
	newrelic.SendError(ctx, errType, err)
	datadog.SendError(ctx, errType, err)
	otel.SendError(ctx, errType, err)
}

func ObserveBufferSize(t string, size int) {
	prometheus.ObserveBufferSize(t, size)
	newrelic.ObserveBufferSize(t, size)
	datadog.ObserveBufferSize(t, size)
	otel.ObserveBufferSize(t, size)
	cloudwatch.ObserveBufferSize(t, size)
}

func SetBufferDefaultSize(t string, size int) {
	prometheus.SetBufferDefaultSize(t, size)
	newrelic.SetBufferDefaultSize(t, size)
	datadog.SetBufferDefaultSize(t, size)
	otel.SetBufferDefaultSize(t, size)
	cloudwatch.SetBufferDefaultSize(t, size)
}

func SetBufferMaxSize(t string, size int) {
	prometheus.SetBufferMaxSize(t, size)
	newrelic.SetBufferMaxSize(t, size)
	datadog.SetBufferMaxSize(t, size)
	otel.SetBufferMaxSize(t, size)
	cloudwatch.SetBufferMaxSize(t, size)
}
