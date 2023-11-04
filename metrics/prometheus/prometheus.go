package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/reuseport"
)

var (
	enabled = false

	requestsTotal    prometheus.Counter
	statusCodesTotal *prometheus.CounterVec
	errorsTotal      *prometheus.CounterVec

	requestDuration     prometheus.Histogram
	requestSpanDuration *prometheus.HistogramVec
	downloadDuration    prometheus.Histogram
	processingDuration  prometheus.Histogram

	bufferSize        *prometheus.HistogramVec
	bufferDefaultSize *prometheus.GaugeVec
	bufferMaxSize     *prometheus.GaugeVec

	requestsInProgress prometheus.GaugeFunc
	imagesInProgress   prometheus.GaugeFunc
)

func Init() {
	if len(config.PrometheusBind) == 0 {
		return
	}

	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "requests_total",
		Help:      "A counter of the total number of HTTP requests imgproxy processed.",
	})

	statusCodesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "status_codes_total",
		Help:      "A counter of the response status codes.",
	}, []string{"status"})

	errorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "errors_total",
		Help:      "A counter of the occurred errors separated by type.",
	}, []string{"type"})

	requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "request_duration_seconds",
		Help:      "A histogram of the response latency.",
	})

	requestSpanDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "request_span_duration_seconds",
		Help:      "A histogram of the queue latency.",
	}, []string{"span"})

	downloadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "download_duration_seconds",
		Help:      "A histogram of the source image downloading latency.",
	})

	processingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "processing_duration_seconds",
		Help:      "A histogram of the image processing latency.",
	})

	bufferSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "buffer_size_bytes",
		Help:      "A histogram of the buffer size in bytes.",
		Buckets:   prometheus.ExponentialBuckets(1024, 2, 14),
	}, []string{"type"})

	bufferDefaultSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "buffer_default_size_bytes",
		Help:      "A gauge of the buffer default size in bytes.",
	}, []string{"type"})

	bufferMaxSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "buffer_max_size_bytes",
		Help:      "A gauge of the buffer max size in bytes.",
	}, []string{"type"})

	requestsInProgress = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "requests_in_progress",
		Help:      "A gauge of the number of requests currently being in progress.",
	}, stats.RequestsInProgress)

	imagesInProgress = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.PrometheusNamespace,
		Name:      "images_in_progress",
		Help:      "A gauge of the number of images currently being in progress.",
	}, stats.ImagesInProgress)

	prometheus.MustRegister(
		requestsTotal,
		statusCodesTotal,
		errorsTotal,
		requestDuration,
		requestSpanDuration,
		downloadDuration,
		processingDuration,
		bufferSize,
		bufferDefaultSize,
		bufferMaxSize,
		requestsInProgress,
		imagesInProgress,
	)

	enabled = true
}

func Enabled() bool {
	return enabled
}

func StartServer(cancel context.CancelFunc) error {
	if !enabled {
		return nil
	}

	s := http.Server{Handler: promhttp.Handler()}

	l, err := reuseport.Listen("tcp", config.PrometheusBind)
	if err != nil {
		return fmt.Errorf("Can't start Prometheus metrics server: %s", err)
	}

	go func() {
		log.Infof("Starting Prometheus server at %s", config.PrometheusBind)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Error(err)
		}
		cancel()
	}()

	return nil
}

func StartRequest(rw http.ResponseWriter) (context.CancelFunc, http.ResponseWriter) {
	if !enabled {
		return func() {}, rw
	}

	requestsTotal.Inc()

	newRw := httpsnoop.Wrap(rw, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				statusCodesTotal.With(prometheus.Labels{"status": strconv.Itoa(statusCode)}).Inc()
				next(statusCode)
			}
		},
	})

	return startDuration(requestDuration), newRw
}

func StartQueueSegment() context.CancelFunc {
	if !enabled {
		return func() {}
	}

	return startDuration(requestSpanDuration.With(prometheus.Labels{"span": "queue"}))
}

func StartDownloadingSegment() context.CancelFunc {
	if !enabled {
		return func() {}
	}

	cancel := startDuration(requestSpanDuration.With(prometheus.Labels{"span": "downloading"}))
	cancelLegacy := startDuration(downloadDuration)

	return func() {
		cancel()
		cancelLegacy()
	}
}

func StartProcessingSegment() context.CancelFunc {
	if !enabled {
		return func() {}
	}

	cancel := startDuration(requestSpanDuration.With(prometheus.Labels{"span": "processing"}))
	cancelLegacy := startDuration(processingDuration)

	return func() {
		cancel()
		cancelLegacy()
	}
}

func StartStreamingSegment() context.CancelFunc {
	if !enabled {
		return func() {}
	}

	return startDuration(requestSpanDuration.With(prometheus.Labels{"span": "streaming"}))
}

func startDuration(m prometheus.Observer) context.CancelFunc {
	t := time.Now()
	return func() {
		m.Observe(time.Since(t).Seconds())
	}
}

func IncrementErrorsTotal(t string) {
	if enabled {
		errorsTotal.With(prometheus.Labels{"type": t}).Inc()
	}
}

func ObserveBufferSize(t string, size int) {
	if enabled {
		bufferSize.With(prometheus.Labels{"type": t}).Observe(float64(size))
	}
}

func SetBufferDefaultSize(t string, size int) {
	if enabled {
		bufferDefaultSize.With(prometheus.Labels{"type": t}).Set(float64(size))
	}
}

func SetBufferMaxSize(t string, size int) {
	if enabled {
		bufferMaxSize.With(prometheus.Labels{"type": t}).Set(float64(size))
	}
}

func AddGaugeFunc(name, help string, f func() float64) {
	if !enabled {
		return
	}

	gauge := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.PrometheusNamespace,
		Name:      name,
		Help:      help,
	}, f)
	prometheus.MustRegister(gauge)
}
