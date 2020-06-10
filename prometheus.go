package main

import (
	"context"
	"fmt"
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
	prometheusVipsMemory         prometheus.GaugeFunc
	prometheusVipsMaxMemory      prometheus.GaugeFunc
	prometheusVipsAllocs         prometheus.GaugeFunc
)

func initPrometheus() {
	if len(conf.PrometheusBind) == 0 {
		return
	}

	prometheusRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "requests_total",
		Help:      "A counter of the total number of HTTP requests imgproxy processed.",
	})

	prometheusErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "errors_total",
		Help:      "A counter of the occurred errors separated by type.",
	}, []string{"type"})

	prometheusRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "request_duration_seconds",
		Help:      "A histogram of the response latency.",
	})

	prometheusDownloadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "download_duration_seconds",
		Help:      "A histogram of the source image downloading latency.",
	})

	prometheusProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "processing_duration_seconds",
		Help:      "A histogram of the image processing latency.",
	})

	prometheusBufferSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "buffer_size_bytes",
		Help:      "A histogram of the buffer size in bytes.",
	}, []string{"type"})

	prometheusBufferDefaultSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "buffer_default_size_bytes",
		Help:      "A gauge of the buffer default size in bytes.",
	}, []string{"type"})

	prometheusBufferMaxSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "buffer_max_size_bytes",
		Help:      "A gauge of the buffer max size in bytes.",
	}, []string{"type"})

	prometheusVipsMemory = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "vips_memory_bytes",
		Help:      "A gauge of the vips tracked memory usage in bytes.",
	}, vipsGetMem)

	prometheusVipsMaxMemory = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "vips_max_memory_bytes",
		Help:      "A gauge of the max vips tracked memory usage in bytes.",
	}, vipsGetMemHighwater)

	prometheusVipsAllocs = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: conf.PrometheusNamespace,
		Name:      "vips_allocs",
		Help:      "A gauge of the number of active vips allocations.",
	}, vipsGetAllocs)

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
}

func startPrometheusServer(cancel context.CancelFunc) error {
	s := http.Server{Handler: promhttp.Handler()}

	l, err := listenReuseport("tcp", conf.PrometheusBind)
	if err != nil {
		return fmt.Errorf("Can't start Prometheus metrics server: %s", err)
	}

	go func() {
		logNotice("Starting Prometheus server at %s", conf.PrometheusBind)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			logError(err.Error())
		}
		cancel()
	}()

	return nil
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
