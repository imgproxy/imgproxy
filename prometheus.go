package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	prometheusEnabled = false

	prometheusRequestsTotal      prometheus.Counter
	prometheusErrorsTotal        *prometheus.CounterVec
	prometheusRequestDuration    prometheus.Histogram
	prometheusDownloadDuration   prometheus.Histogram
	prometheusProcessingDuration prometheus.Histogram
	prometheusBufferSize         *prometheus.HistogramVec
	prometheusBufferDefaultSize  *prometheus.GaugeVec
	prometheusBufferMaxSize      *prometheus.GaugeVec
	prometheusVipsMemory         prometheus.Gauge
	prometheusVipsMaxMemory      prometheus.Gauge
	prometheusVipsAllocs         prometheus.Gauge
)

func initPrometheus() {
	if len(conf.PrometheusBind) == 0 {
		return
	}

	prometheusRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "A counter of the total number of HTTP requests imgproxy processed.",
	})

	prometheusErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "A counter of the occurred errors separated by type.",
	}, []string{"type"})

	prometheusRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "A histogram of the response latency.",
	})

	prometheusDownloadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "download_duration_seconds",
		Help: "A histogram of the source image downloading latency.",
	})

	prometheusProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "processing_duration_seconds",
		Help: "A histogram of the image processing latency.",
	})

	prometheusBufferSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "buffer_size_bytes",
		Help: "A histogram of the buffer size in bytes.",
	}, []string{"type"})

	prometheusBufferDefaultSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "buffer_default_size_bytes",
		Help: "A gauge of the buffer default size in bytes.",
	}, []string{"type"})

	prometheusBufferMaxSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "buffer_max_size_bytes",
		Help: "A gauge of the buffer max size in bytes.",
	}, []string{"type"})

	prometheusVipsMemory = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vips_memory_bytes",
		Help: "A gauge of the vips tracked memory usage in bytes.",
	})

	prometheusVipsMaxMemory = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vips_max_memory_bytes",
		Help: "A gauge of the max vips tracked memory usage in bytes.",
	})

	prometheusVipsAllocs = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vips_allocs",
		Help: "A gauge of the number of active vips allocations.",
	})

	prometheus.MustRegister(
		prometheusRequestsTotal,
		prometheusErrorsTotal,
		prometheusRequestDuration,
		prometheusDownloadDuration,
		prometheusProcessingDuration,
		prometheusBufferSize,
		prometheusBufferDefaultSize,
		prometheusBufferMaxSize,
		prometheusVipsMemory,
		prometheusVipsMaxMemory,
		prometheusVipsAllocs,
	)

	prometheusEnabled = true

	s := http.Server{Handler: promhttp.Handler()}

	go func() {
		l, err := listenReuseport("tcp", conf.PrometheusBind)
		if err != nil {
			logFatal(err.Error())
		}

		logNotice("Starting Prometheus server at %s\n", conf.PrometheusBind)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			logFatal(err.Error())
		}
	}()
}

func startPrometheusDuration(m prometheus.Histogram) func() {
	t := time.Now()
	return func() {
		m.Observe(time.Since(t).Seconds())
	}
}

func incrementPrometheusErrorsTotal(t string) {
	prometheusErrorsTotal.With(prometheus.Labels{"type": t}).Inc()
}

func observePrometheusBufferSize(t string, size int) {
	prometheusBufferSize.With(prometheus.Labels{"type": t}).Observe(float64(size))
}

func setPrometheusBufferDefaultSize(t string, size int) {
	prometheusBufferDefaultSize.With(prometheus.Labels{"type": t}).Set(float64(size))
}

func setPrometheusBufferMaxSize(t string, size int) {
	prometheusBufferMaxSize.With(prometheus.Labels{"type": t}).Set(float64(size))
}
