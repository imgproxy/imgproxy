package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/reuseport"
)

var (
	enabled = false

	requestsTotal      prometheus.Counter
	errorsTotal        *prometheus.CounterVec
	requestDuration    prometheus.Histogram
	downloadDuration   prometheus.Histogram
	processingDuration prometheus.Histogram
	bufferSize         *prometheus.HistogramVec
	bufferDefaultSize  *prometheus.GaugeVec
	bufferMaxSize      *prometheus.GaugeVec
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

	prometheus.MustRegister(
		requestsTotal,
		errorsTotal,
		requestDuration,
		downloadDuration,
		processingDuration,
		bufferSize,
		bufferDefaultSize,
		bufferMaxSize,
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

func StartRequest() context.CancelFunc {
	if !enabled {
		return func() {}
	}

	requestsTotal.Inc()
	return startDuration(requestDuration)
}

func StartDownloadingSegment() context.CancelFunc {
	return startDuration(downloadDuration)
}

func StartProcessingSegment() context.CancelFunc {
	return startDuration(processingDuration)
}

func startDuration(m prometheus.Histogram) context.CancelFunc {
	if !enabled {
		return func() {}
	}

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
