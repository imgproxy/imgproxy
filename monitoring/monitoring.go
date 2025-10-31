package monitoring

import (
	"context"
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/monitoring/cloudwatch"
	"github.com/imgproxy/imgproxy/v3/monitoring/datadog"
	"github.com/imgproxy/imgproxy/v3/monitoring/newrelic"
	"github.com/imgproxy/imgproxy/v3/monitoring/otel"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
)

// Monitoring holds all monitoring service instances
type Monitoring struct {
	config *Config
	stats  *stats.Stats

	prometheus *prometheus.Prometheus
	newrelic   *newrelic.NewRelic
	datadog    *datadog.DataDog
	otel       *otel.Otel
	cloudwatch *cloudwatch.CloudWatch
}

// New creates a new Monitoring instance
func New(ctx context.Context, config *Config, workersNumber int) (*Monitoring, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	m := &Monitoring{
		config: config,
		stats:  stats.New(workersNumber),
	}

	var prErr, nlErr, ddErr, otelErr, cwErr error

	m.prometheus, prErr = prometheus.New(&config.Prometheus, m.stats)
	m.newrelic, nlErr = newrelic.New(&config.NewRelic, m.stats)
	m.datadog, ddErr = datadog.New(&config.DataDog, m.stats)
	m.otel, otelErr = otel.New(&config.OpenTelemetry, m.stats)
	m.cloudwatch, cwErr = cloudwatch.New(ctx, &config.CloudWatch, m.stats)

	err := errors.Join(prErr, nlErr, ddErr, otelErr, cwErr)

	return m, err
}

// Enabled returns true if at least one monitoring service is enabled
func (m *Monitoring) Enabled() bool {
	return m.prometheus.Enabled() ||
		m.newrelic.Enabled() ||
		m.datadog.Enabled() ||
		m.otel.Enabled() ||
		m.cloudwatch.Enabled()
}

// Stats returns the stats instance
func (m *Monitoring) Stats() *stats.Stats {
	return m.stats
}

// Stop stops all monitoring services
func (m *Monitoring) Stop(ctx context.Context) {
	m.newrelic.Stop(ctx)
	m.datadog.Stop()
	m.otel.Stop(ctx)
	m.cloudwatch.Stop()
}

// StartPrometheus starts the Prometheus metrics server
func (m *Monitoring) StartPrometheus(cancel context.CancelFunc) error {
	return m.prometheus.StartServer(cancel)
}

func (m *Monitoring) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	promCancel, rw := m.prometheus.StartRequest(rw)
	ctx, nrCancel, rw := m.newrelic.StartRequest(ctx, rw, r)
	ctx, ddCancel, rw := m.datadog.StartRequest(ctx, rw, r)
	ctx, otelCancel, rw := m.otel.StartRequest(ctx, rw, r)

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return ctx, cancel, rw
}

func (m *Monitoring) SetMetadata(ctx context.Context, meta Meta) {
	for key, value := range meta {
		m.newrelic.SetMetadata(ctx, key, value)
		m.datadog.SetMetadata(ctx, key, value)
		m.otel.SetMetadata(ctx, key, value)
	}
}

func (m *Monitoring) StartSpan(
	ctx context.Context,
	name string,
	meta Meta,
) (context.Context, context.CancelFunc) {
	promCancel := m.prometheus.StartSpan(name)
	nrCancel := m.newrelic.StartSpan(ctx, name, meta)
	ctx, ddCancel := m.datadog.StartSpan(ctx, name, meta)
	ctx, otelCancel := m.otel.StartSpan(ctx, name, meta)

	cancel := func() {
		promCancel()
		nrCancel()
		ddCancel()
		otelCancel()
	}

	return ctx, cancel
}

func (m *Monitoring) SendError(ctx context.Context, errType string, err error) {
	m.prometheus.IncrementErrorsTotal(errType)
	m.newrelic.SendError(ctx, errType, err)
	m.datadog.SendError(ctx, errType, err)
	m.otel.SendError(ctx, errType, err)
}

// InjectHeaders adds monitoring headers to the provided HTTP headers.
// These headers can be used to correlate requests across different services.
func (m *Monitoring) InjectHeaders(ctx context.Context, headers http.Header) {
	m.newrelic.InjectHeaders(ctx, headers)
	m.datadog.InjectHeaders(ctx, headers)
	m.otel.InjectHeaders(ctx, headers)
}
