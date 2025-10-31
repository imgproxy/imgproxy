package prometheus

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/imgproxy/imgproxy/v3/monitoring/format"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// Prometheus holds Prometheus metrics and configuration
type Prometheus struct {
	config *Config
	stats  *stats.Stats

	requestsTotal    prometheus.Counter
	statusCodesTotal *prometheus.CounterVec
	errorsTotal      *prometheus.CounterVec

	requestDuration     prometheus.Histogram
	requestSpanDuration *prometheus.HistogramVec

	workers prometheus.Gauge
}

// New creates a new Prometheus instance
func New(config *Config, stats *stats.Stats) (*Prometheus, error) {
	p := &Prometheus{
		config: config,
		stats:  stats,
	}

	if !config.Enabled() {
		return p, nil
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	p.requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: config.Namespace,
		Name:      "requests_total",
		Help:      "A counter of the total number of HTTP requests imgproxy processed.",
	})

	p.statusCodesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.Namespace,
		Name:      "status_codes_total",
		Help:      "A counter of the response status codes.",
	}, []string{"status"})

	p.errorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.Namespace,
		Name:      "errors_total",
		Help:      "A counter of the occurred errors separated by type.",
	}, []string{"type"})

	p.requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: config.Namespace,
		Name:      "request_duration_seconds",
		Help:      "A histogram of the response latency.",
	})

	p.requestSpanDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.Namespace,
		Name:      "request_span_duration_seconds",
		Help:      "A histogram of the queue latency.",
	}, []string{"span"})

	p.workers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "workers",
		Help:      "A gauge of the number of running workers.",
	})
	p.workers.Set(float64(stats.WorkersNumber))

	requestsInProgress := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "requests_in_progress",
		Help:      "A gauge of the number of requests currently being in progress.",
	}, stats.RequestsInProgress)

	imagesInProgress := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "images_in_progress",
		Help:      "A gauge of the number of images currently being in progress.",
	}, stats.ImagesInProgress)

	workersUtilization := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "workers_utilization",
		Help:      "A gauge of the workers utilization in percents.",
	}, stats.WorkersUtilization)

	vipsMemoryBytes := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "vips_memory_bytes",
		Help:      "A gauge of the vips tracked memory usage in bytes.",
	}, vips.GetMem)

	vipsMaxMemoryBytes := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "vips_max_memory_bytes",
		Help:      "A gauge of the max vips tracked memory usage in bytes.",
	}, vips.GetMemHighwater)

	vipsAllocs := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "vips_allocs",
		Help:      "A gauge of the number of active vips allocations.",
	}, vips.GetAllocs)

	prometheus.MustRegister(
		p.requestsTotal,
		p.statusCodesTotal,
		p.errorsTotal,
		p.requestDuration,
		p.requestSpanDuration,
		p.workers,
		requestsInProgress,
		imagesInProgress,
		workersUtilization,
		vipsMemoryBytes,
		vipsMaxMemoryBytes,
		vipsAllocs,
	)

	return p, nil
}

// Enabled returns true if Prometheus monitoring is enabled
func (p *Prometheus) Enabled() bool {
	return p.config.Enabled()
}

// StartServer starts the Prometheus metrics server
func (p *Prometheus) StartServer(cancel context.CancelFunc) error {
	// If not enabled, do nothing
	if !p.Enabled() {
		return nil
	}

	s := http.Server{Handler: promhttp.Handler()}

	l, err := net.Listen("tcp", p.config.Bind)
	if err != nil {
		return fmt.Errorf("can't start Prometheus metrics server: %s", err)
	}

	go func() {
		slog.Info(fmt.Sprintf("Starting Prometheus server at %s", p.config.Bind))
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error())
		}
		cancel()
	}()

	return nil
}

func (p *Prometheus) StartRequest(rw http.ResponseWriter) (context.CancelFunc, http.ResponseWriter) {
	if !p.Enabled() {
		return func() {}, rw
	}

	p.requestsTotal.Inc()

	newRw := httpsnoop.Wrap(rw, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				p.statusCodesTotal.With(prometheus.Labels{"status": strconv.Itoa(statusCode)}).Inc()
				next(statusCode)
			}
		},
	})

	return p.startDuration(p.requestDuration), newRw
}

func (p *Prometheus) StartSpan(name string) context.CancelFunc {
	if !p.Enabled() {
		return func() {}
	}

	return p.startDuration(
		p.requestSpanDuration.With(prometheus.Labels{"span": format.FormatSegmentName(name)}),
	)
}

func (p *Prometheus) startDuration(m prometheus.Observer) context.CancelFunc {
	t := time.Now()
	return func() {
		m.Observe(time.Since(t).Seconds())
	}
}

func (p *Prometheus) IncrementErrorsTotal(t string) {
	if !p.Enabled() {
		return
	}

	p.errorsTotal.With(prometheus.Labels{"type": t}).Inc()
}
