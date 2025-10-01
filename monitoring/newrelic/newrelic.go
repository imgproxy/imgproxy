package newrelic

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"

	"github.com/imgproxy/imgproxy/v3/monitoring/errformat"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// transactionCtxKey context key for storing New Relic transaction in context
type transactionCtxKey struct{}

// attributable is an interface for New Relic entities that can have attributes set on them
type attributable interface {
	AddAttribute(key string, value any)
}

const (
	// Metric API endpoints. NOTE: Possibly, this should be configurable?
	defaultMetricURL = "https://metric-api.newrelic.com/metric/v1"
	euMetricURL      = "https://metric-api.eu.newrelic.com/metric/v1"
)

type NewRelic struct {
	stats  *stats.Stats
	config *Config

	app       *newrelic.Application
	harvester *telemetry.Harvester

	harvesterCtx       context.Context
	harvesterCtxCancel context.CancelFunc
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

	// Initialize New Relic Telemetry SDK harvester
	harvesterAttributes := map[string]any{"appName": config.AppName}
	for k, v := range config.Labels {
		harvesterAttributes[k] = v
	}

	// Choose metrics endpoint based on license key pattern
	licenseEuRegex := regexp.MustCompile(`(^eu.+?)x`)

	metricsURL := defaultMetricURL
	if licenseEuRegex.MatchString(config.Key) {
		metricsURL = euMetricURL
	}

	// Initialize error logger
	errLogger := slog.NewLogLogger(
		slog.With("source", "newrelic").Handler(),
		slog.LevelWarn,
	)

	// Create harvester
	harvester, err := telemetry.NewHarvester(
		telemetry.ConfigAPIKey(config.Key),
		telemetry.ConfigCommonAttributes(harvesterAttributes),
		telemetry.ConfigHarvestPeriod(0), // Don't harvest automatically
		telemetry.ConfigMetricsURLOverride(metricsURL),
		telemetry.ConfigBasicErrorLogger(errLogger.Writer()),
	)
	if err == nil {
		// In case, there were no errors while starting the harvester, start the metrics collector
		nl.harvester = harvester
		nl.harvesterCtx, nl.harvesterCtxCancel = context.WithCancel(context.Background())
		go nl.runMetricsCollector()
	} else {
		slog.Warn(fmt.Sprintf("Can't init New Relic telemetry harvester: %s", err))
	}

	return nl, nil
}

// Enabled returns true if New Relic is enabled
func (nl *NewRelic) Enabled() bool {
	return nl.config.Enabled()
}

// Stop stops the New Relic APM agent and Telemetry SDK harvester
func (nl *NewRelic) Stop(ctx context.Context) {
	if nl.app != nil {
		nl.app.Shutdown(5 * time.Second)
	}

	if nl.harvester != nil {
		nl.harvesterCtxCancel()
		nl.harvester.HarvestNow(ctx)
	}
}

func (nl *NewRelic) StartTransaction(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !nl.Enabled() {
		return ctx, func() {}, rw
	}

	txn := nl.app.StartTransaction("request")
	txn.SetWebRequestHTTP(r)
	newRw := txn.SetWebResponse(rw)
	cancel := func() { txn.End() }
	return context.WithValue(ctx, transactionCtxKey{}, txn), cancel, newRw
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

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		setMetadata(txn, key, value)
	}
}

func (nl *NewRelic) StartSegment(ctx context.Context, name string, meta map[string]any) context.CancelFunc {
	if !nl.Enabled() {
		return func() {}
	}

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
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

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		txn.NoticeError(newrelic.Error{
			Message: err.Error(),
			Class:   errformat.FormatErrType(errType, err),
		})
	}
}

func (nl *NewRelic) runMetricsCollector() {
	tick := time.NewTicker(nl.config.MetricsInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.workers",
				Value:     float64(nl.stats.WorkersNumber),
				Timestamp: time.Now(),
			})

			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.requests_in_progress",
				Value:     nl.stats.RequestsInProgress(),
				Timestamp: time.Now(),
			})

			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.images_in_progress",
				Value:     nl.stats.ImagesInProgress(),
				Timestamp: time.Now(),
			})

			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.workers_utilization",
				Value:     nl.stats.WorkersUtilization(),
				Timestamp: time.Now(),
			})

			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.vips.memory",
				Value:     vips.GetMem(),
				Timestamp: time.Now(),
			})

			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.vips.max_memory",
				Value:     vips.GetMemHighwater(),
				Timestamp: time.Now(),
			})

			nl.harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.vips.allocs",
				Value:     vips.GetAllocs(),
				Timestamp: time.Now(),
			})

			nl.harvester.HarvestNow(nl.harvesterCtx)
		case <-nl.harvesterCtx.Done():
			return
		}
	}
}
