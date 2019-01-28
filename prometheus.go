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
	prometheusBuffersTotal       *prometheus.CounterVec
	prometheusBufferSize         *prometheus.HistogramVec
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
		Help: "A counter of the occured errors separated by type.",
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

	prometheusBuffersTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "buffers_total",
		Help: "A counter of the total number of buffers imgproxy allocated.",
	}, []string{"type"})

	prometheusBufferSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "buffer_size_megabytes",
		Help: "A histogram of the buffer size in megabytes.",
	}, []string{"type"})

	prometheus.MustRegister(
		prometheusRequestsTotal,
		prometheusErrorsTotal,
		prometheusRequestDuration,
		prometheusDownloadDuration,
		prometheusProcessingDuration,
		prometheusBuffersTotal,
		prometheusBufferSize,
	)

	prometheusEnabled = true

	s := http.Server{
		Addr:    conf.PrometheusBind,
		Handler: promhttp.Handler(),
	}

	go func() {
		logNotice("Starting Prometheus server at %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func incrementBuffersTotal(t string) {
	prometheusBuffersTotal.With(prometheus.Labels{"type": t}).Inc()
}

func observeBufferSize(t string, cap int) {
	size := float64(cap) / 1024.0 / 1024.0
	prometheusBufferSize.With(prometheus.Labels{"type": t}).Observe(size)
}
