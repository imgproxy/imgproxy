package newrelic

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/monitoring/format"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	vipsstats "github.com/imgproxy/imgproxy/v3/vips/stats"
)

// attributable is an interface for New Relic entities that can have attributes set on them
type attributable interface {
	AddAttribute(key string, value any)
}

// NewRelic holds New Relic APM agent and configuration
type NewRelic struct {
	stats  *stats.Stats
	config *Config

	app *newrelic.Application

	metricsCtx       context.Context //nolint:containedctx
	metricsCtxCancel context.CancelFunc
}

// New creates a new [NewRelic] instance
func New(config *Config, stats *stats.Stats) (*NewRelic, error) {
	if !config.Enabled() {
		return nil, nil
	}

	nl := &NewRelic{
		config: config,
		stats:  stats,
	}

	var err error

	// Initialize New Relic APM agent
	nl.app, err = newrelic.NewApplication(
		newrelic.ConfigAppName(config.AppName),
		newrelic.ConfigLicense(config.Key),
		func(c *newrelic.Config) {
			if len(config.Labels) > 0 {
				c.Labels = config.Labels
			}
		},
	)

	if err != nil {
		return nil, fmt.Errorf("can't init New Relic agent: %w", err)
	}

	nl.metricsCtx, nl.metricsCtxCancel = context.WithCancel(context.Background())
	go nl.runMetricsCollector()

	return nl, nil
}

// Stop stops the New Relic APM agent and Telemetry SDK harvester
func (nl *NewRelic) Stop(ctx context.Context) {
	if nl.metricsCtxCancel != nil {
		nl.metricsCtxCancel()
	}

	if nl.app != nil {
		nl.app.Shutdown(5 * time.Second)
	}
}

// StartRequest starts a new New Relic transaction for the incoming HTTP request
func (nl *NewRelic) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	txn := nl.app.StartTransaction("request")
	txn.SetWebRequestHTTP(r)
	newRw := txn.SetWebResponse(rw)
	cancel := func() { txn.End() }
	return newrelic.NewContext(ctx, txn), cancel, newRw
}

// setMetadata sets metadata on the given New Relic attributable entity
func setMetadata(span attributable, key string, value any) {
	if len(key) == 0 || value == nil {
		return
	}

	if stringer, ok := value.(fmt.Stringer); ok {
		span.AddAttribute(key, stringer.String())
		return
	}

	rv := reflect.ValueOf(value)
	switch {
	case rv.Kind() == reflect.String || rv.Kind() == reflect.Bool:
		span.AddAttribute(key, value)
	case rv.CanInt():
		span.AddAttribute(key, rv.Int())
	case rv.CanUint():
		span.AddAttribute(key, rv.Uint())
	case rv.CanFloat():
		span.AddAttribute(key, rv.Float())
	case rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String:
		for _, k := range rv.MapKeys() {
			setMetadata(span, key+"."+k.String(), rv.MapIndex(k).Interface())
		}
	default:
		span.AddAttribute(key, fmt.Sprintf("%v", value))
	}
}

// SetMetadata sets metadata for the current transaction
func (nl *NewRelic) SetMetadata(ctx context.Context, key string, value any) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		setMetadata(txn, key, value)
	}
}

// StartSpan starts a new span for New Relic monitoring
func (nl *NewRelic) StartSpan(
	ctx context.Context,
	name string,
	meta map[string]any,
) (context.Context, context.CancelFunc) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		segment := txn.NewGoroutine().StartSegment(name)

		for k, v := range meta {
			setMetadata(segment, k, v)
		}

		return ctx, func() { segment.End() }
	}

	return ctx, func() {}
}

// SendError sends an error to New Relic APM
func (nl *NewRelic) SendError(ctx context.Context, errType string, err errctx.Error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		txn.NoticeError(newrelic.Error{
			Message: err.Error(),
			Class:   format.FormatErrType(errType, err),
			Stack:   err.StackTrace(),
		})
	}
}

// InjectHeaders adds monitoring headers to the provided HTTP headers.
func (nl *NewRelic) InjectHeaders(ctx context.Context, headers http.Header) {
	if !nl.config.PropagateExt {
		return
	}

	if txn := newrelic.FromContext(ctx); txn != nil {
		txn.InsertDistributedTraceHeaders(headers)
	}
}

// runMetricsCollector periodically collects and sends custom metrics to New Relic
func (nl *NewRelic) runMetricsCollector() {
	tick := time.NewTicker(nl.config.MetricsInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			nl.app.RecordCustomMetric("imgproxy/workers", float64(nl.stats.WorkersNumber))
			nl.app.RecordCustomMetric("imgproxy/requests_in_progress", float64(nl.stats.RequestsInProgress()))
			nl.app.RecordCustomMetric("imgproxy/images_in_progress", float64(nl.stats.ImagesInProgress()))
			nl.app.RecordCustomMetric("imgproxy/workers_utilization", nl.stats.WorkersUtilization())

			nl.app.RecordCustomMetric("imgproxy/vips/memory", float64(vipsstats.Memory()))
			nl.app.RecordCustomMetric("imgproxy/vips/max_memory", float64(vipsstats.MemoryHighwater()))
			nl.app.RecordCustomMetric("imgproxy/vips/allocs", float64(vipsstats.Allocs()))
		case <-nl.metricsCtx.Done():
			return
		}
	}
}
