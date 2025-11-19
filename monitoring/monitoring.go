package monitoring

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/monitoring/cloudwatch"
	"github.com/imgproxy/imgproxy/v3/monitoring/datadog"
	"github.com/imgproxy/imgproxy/v3/monitoring/newrelic"
	"github.com/imgproxy/imgproxy/v3/monitoring/otel"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
)

// monitor is an interface for monitoring services
type monitor interface {
	Stop(ctx context.Context)
	StartRequest(
		ctx context.Context,
		rw http.ResponseWriter,
		r *http.Request,
	) (context.Context, context.CancelFunc, http.ResponseWriter)
	StartSpan(
		ctx context.Context,
		name string,
		meta map[string]any,
	) (context.Context, context.CancelFunc)
	SetMetadata(ctx context.Context, key string, value any)
	SendError(ctx context.Context, errType string, err errctx.Error)
	InjectHeaders(ctx context.Context, headers http.Header)
}

// Monitoring holds all monitoring service instances
type Monitoring struct {
	config *Config
	stats  *stats.Stats

	monitors []monitor
}

// New creates a new Monitoring instance
func New(ctx context.Context, config *Config, workersNumber int) (*Monitoring, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	st := stats.New(workersNumber)
	monitors := make([]monitor, 0)

	if m, err := prometheus.New(&config.Prometheus, st); err != nil {
		return nil, err
	} else if m != nil {
		monitors = append(monitors, m)
	}

	if m, err := newrelic.New(&config.NewRelic, st); err != nil {
		return nil, err
	} else if m != nil {
		monitors = append(monitors, m)
	}

	if m, err := datadog.New(&config.DataDog, st); err != nil {
		return nil, err
	} else if m != nil {
		monitors = append(monitors, m)
	}

	if m, err := otel.New(&config.OpenTelemetry, st); err != nil {
		return nil, err
	} else if m != nil {
		monitors = append(monitors, m)
	}

	if m, err := cloudwatch.New(ctx, &config.CloudWatch, st); err != nil {
		return nil, err
	} else if m != nil {
		monitors = append(monitors, m)
	}

	m := &Monitoring{
		config:   config,
		stats:    st,
		monitors: monitors,
	}

	return m, nil
}

// Enabled returns true if at least one monitoring service is enabled
func (m *Monitoring) Enabled() bool {
	return len(m.monitors) > 0
}

// Stats returns the stats instance
func (m *Monitoring) Stats() *stats.Stats {
	return m.stats
}

// Stop stops all monitoring services
func (m *Monitoring) Stop(ctx context.Context) {
	for _, monitor := range m.monitors {
		monitor.Stop(ctx)
	}
}

// StartPrometheus starts the Prometheus metrics server
func (m *Monitoring) StartPrometheus(cancel context.CancelFunc) error {
	for _, monitor := range m.monitors {
		if prom, ok := monitor.(*prometheus.Prometheus); ok {
			return prom.StartServer(cancel)
		}
	}
	return nil
}

// StartRequest starts a new request span
func (m *Monitoring) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	cancels := make([]context.CancelFunc, 0, len(m.monitors))

	for _, monitor := range m.monitors {
		var cancel context.CancelFunc
		ctx, cancel, rw = monitor.StartRequest(ctx, rw, r)
		cancels = append(cancels, cancel)
	}

	cancel := func() {
		for _, c := range cancels {
			c()
		}
	}

	return ctx, cancel, rw
}

// SetMetadata sets metadata key-value pair for all monitoring services
func (m *Monitoring) SetMetadata(ctx context.Context, meta Meta) {
	for _, monitor := range m.monitors {
		for key, value := range meta {
			monitor.SetMetadata(ctx, key, value)
		}
	}
}

// StartSpan starts a new trace span as child of the current span
func (m *Monitoring) StartSpan(
	ctx context.Context,
	name string,
	meta Meta,
) (context.Context, context.CancelFunc) {
	cancels := make([]context.CancelFunc, 0, len(m.monitors))

	for _, monitor := range m.monitors {
		var cancel context.CancelFunc
		ctx, cancel = monitor.StartSpan(ctx, name, meta)
		cancels = append(cancels, cancel)
	}

	cancel := func() {
		for _, c := range cancels {
			c()
		}
	}

	return ctx, cancel
}

// SendError sends an error to all monitoring services
func (m *Monitoring) SendError(ctx context.Context, errType string, err errctx.Error) {
	for _, monitor := range m.monitors {
		monitor.SendError(ctx, errType, err)
	}
}

// InjectHeaders adds monitoring headers to the provided HTTP headers.
// These headers can be used to correlate requests across different services.
func (m *Monitoring) InjectHeaders(ctx context.Context, headers http.Header) {
	for _, monitor := range m.monitors {
		monitor.InjectHeaders(ctx, headers)
	}
}
