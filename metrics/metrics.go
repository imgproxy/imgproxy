package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/metrics/datadog"
	"github.com/imgproxy/imgproxy/v3/metrics/newrelic"
	"github.com/imgproxy/imgproxy/v3/metrics/prometheus"
)

func Init() error {
	prometheus.Init()

	if err := newrelic.Init(); err != nil {
		return nil
	}

	datadog.Init()

	return nil
}

func Stop() {
	datadog.Stop()
}

func Enabled() bool {
	return prometheus.Enabled() ||
		newrelic.Enabled() ||
		datadog.Enabled()
}

func StartRequest(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	promCancel := prometheus.StartRequest()
	ctx, nrCancel, rw := newrelic.StartTransaction(ctx, rw, r)
	ctx, ddCancel, rw := datadog.StartRootSpan(ctx, rw, r)

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
	}

	return ctx, cancel, rw
}

func StartDownloadingSegment(ctx context.Context) context.CancelFunc {
	promCancel := prometheus.StartDownloadingSegment()
	nrCancel := newrelic.StartSegment(ctx, "Downloading image")
	ddCancel := datadog.StartSpan(ctx, "downloading_image")

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
	}

	return cancel
}

func StartProcessingSegment(ctx context.Context) context.CancelFunc {
	promCancel := prometheus.StartProcessingSegment()
	nrCancel := newrelic.StartSegment(ctx, "Processing image")
	ddCancel := datadog.StartSpan(ctx, "processing_image")

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
	}

	return cancel
}

func SendError(ctx context.Context, errType string, err error) {
	prometheus.IncrementErrorsTotal(errType)
	newrelic.SendError(ctx, err)
	datadog.SendError(ctx, err)
}

func SendTimeout(ctx context.Context, d time.Duration) {
	prometheus.IncrementErrorsTotal("timeout")
	newrelic.SendTimeout(ctx, d)
	datadog.SendTimeout(ctx, d)
}
