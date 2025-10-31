package newrelic

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/imgproxy/imgproxy/v3/monitoring/format"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// attributable is an interface for New Relic entities that can have attributes set on them
type attributable interface {
	AddAttribute(key string, value any)
}

type NewRelic struct {
	stats  *stats.Stats
	config *Config

	app *newrelic.Application

	metricsCtx       context.Context
	metricsCtxCancel context.CancelFunc
}

func New(config *Config, stats *stats.Stats) (*NewRelic, error) {
	nl := &NewRelic{
		config: config,
		stats:  stats,
	}

	if !config.Enabled() {
		return nl, nil
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
		return nil, fmt.Errorf("can't init New Relic agent: %s", err)
	}

	nl.metricsCtx, nl.metricsCtxCancel = context.WithCancel(context.Background())
	go nl.runMetricsCollector()

	return nl, nil
}

// Enabled returns true if New Relic is enabled
func (nl *NewRelic) Enabled() bool {
	return nl.config.Enabled()
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

func (nl *NewRelic) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !nl.Enabled() {
		return ctx, func() {}, rw
	}

	txn := nl.app.StartTransaction("request")
	txn.SetWebRequestHTTP(r)
	newRw := txn.SetWebResponse(rw)
	cancel := func() { txn.End() }
	return newrelic.NewContext(ctx, txn), cancel, newRw
}

func setMetadata(span attributable, key string, value interface{}) {
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

func (nl *NewRelic) SetMetadata(ctx context.Context, key string, value interface{}) {
	if !nl.Enabled() {
		return
	}

	if txn := newrelic.FromContext(ctx); txn != nil {
		setMetadata(txn, key, value)
	}
}

func (nl *NewRelic) StartSpan(
	ctx context.Context,
	name string,
	meta map[string]any,
) context.CancelFunc {
	if !nl.Enabled() {
		return func() {}
	}

	if txn := newrelic.FromContext(ctx); txn != nil {
		segment := txn.NewGoroutine().StartSegment(name)

		for k, v := range meta {
			setMetadata(segment, k, v)
		}

		return func() { segment.End() }
	}

	return func() {}
}

func (nl *NewRelic) SendError(ctx context.Context, errType string, err error) {
	if !nl.Enabled() {
		return
	}

	if txn := newrelic.FromContext(ctx); txn != nil {
		txn.NoticeError(newrelic.Error{
			Message: err.Error(),
			Class:   format.FormatErrType(errType, err),
		})
	}
}

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

			nl.app.RecordCustomMetric("imgproxy/vips/memory", float64(vips.GetMem()))
			nl.app.RecordCustomMetric("imgproxy/vips/max_memory", float64(vips.GetMemHighwater()))
			nl.app.RecordCustomMetric("imgproxy/vips/allocs", float64(vips.GetAllocs()))
		case <-nl.metricsCtx.Done():
			return
		}
	}
}
