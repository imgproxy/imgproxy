package prometheus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/monitoring/format"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	vipsstats "github.com/imgproxy/imgproxy/v3/vips/stats"
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
	if !config.Enabled() {
		return nil, nil
	}

	p := &Prometheus{
		config: config,
		stats:  stats,
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
		Help:      "A histogram of the request spans duration separated by span name.",
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
	}, vipsstats.Memory)

	vipsMaxMemoryBytes := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "vips_max_memory_bytes",
		Help:      "A gauge of the max vips tracked memory usage in bytes.",
	}, vipsstats.MemoryHighwater)

	vipsAllocs := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: config.Namespace,
		Name:      "vips_allocs",
		Help:      "A gauge of the number of active vips allocations.",
	}, vipsstats.Allocs)

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

// Stop stops the Prometheus monitoring
func (p *Prometheus) Stop(ctx context.Context) {
	// No-op
}

// StartServer starts the Prometheus metrics server
func (p *Prometheus) StartServer(cancel context.CancelFunc) error {
	s := http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           promhttp.Handler(),
	}

	l, err := net.Listen("tcp", p.config.Bind)
	if err != nil {
		return fmt.Errorf("can't start Prometheus metrics server: %w", err)
	}

	go func() {
		slog.Info(fmt.Sprintf("Starting Prometheus server at %s", p.config.Bind))
		if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
		}
		cancel()
	}()

	return nil
}

// StartRequest starts a new request for Prometheus monitoring
func (p *Prometheus) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	p.requestsTotal.Inc()

	newRw := httpsnoop.Wrap(rw, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				p.statusCodesTotal.With(prometheus.Labels{"status": strconv.Itoa(statusCode)}).Inc()
				next(statusCode)
			}
		},
	})

	return ctx, p.startDuration(p.requestDuration), newRw
}

// StartSpan starts a new span for Prometheus monitoring
func (p *Prometheus) StartSpan(
	ctx context.Context,
	name string,
	meta map[string]any,
) (context.Context, context.CancelFunc) {
	return ctx, p.startDuration(
		p.requestSpanDuration.With(prometheus.Labels{"span": format.FormatSegmentName(name)}),
	)
}

// SetMetadata sets metadata for Prometheus monitoring
func (p *Prometheus) SetMetadata(ctx context.Context, key string, value any) {
	// Prometheus does not support request tracing
}

// SendError records an error occurrence in Prometheus metrics
func (p *Prometheus) SendError(ctx context.Context, errType string, err errctx.Error) {
	p.errorsTotal.With(prometheus.Labels{"type": errType}).Inc()
}

// InjectHeaders adds monitoring headers to the provided HTTP headers.
func (p *Prometheus) InjectHeaders(ctx context.Context, headers http.Header) {
	// Prometheus does not support request tracing
}

// startDuration starts a timer and returns a cancel function to record the duration
func (p *Prometheus) startDuration(m prometheus.Observer) context.CancelFunc {
	t := time.Now()
	return func() {
		m.Observe(time.Since(t).Seconds())
	}
}
