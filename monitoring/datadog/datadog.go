package datadog

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/felixge/httpsnoop"

	"github.com/imgproxy/imgproxy/v3/monitoring/errformat"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// spanCtxKey is the context key type for storing the root span in the request context
type spanCtxKey struct{}

// dataDogLogger is a custom logger for DataDog
type dataDogLogger struct{}

func (l dataDogLogger) Log(msg string) {
	slog.Info(msg)
}

// DataDog holds DataDog client and configuration
type DataDog struct {
	stats  *stats.Stats
	config *Config

	statsdClient     *statsd.Client
	statsdClientStop chan struct{}
}

// New creates a new DataDog instance
func New(config *Config, stats *stats.Stats) (*DataDog, error) {
	dd := &DataDog{
		stats:  stats,
		config: config,
	}

	if !config.Enabled() {
		return dd, nil
	}

	tracer.Start(
		tracer.WithService(config.Service),
		tracer.WithServiceVersion(version.Version),
		tracer.WithLogger(dataDogLogger{}),
		tracer.WithLogStartup(config.TraceStartupLogs),
		tracer.WithAgentAddr(net.JoinHostPort(config.AgentHost, strconv.Itoa(config.TracePort))),
	)

	// If additional metrics collection is not enabled, return early
	if !config.EnableMetrics {
		return dd, nil
	}

	var err error

	dd.statsdClient, err = statsd.New(
		net.JoinHostPort(config.AgentHost, strconv.Itoa(config.StatsDPort)),
		statsd.WithTags([]string{
			"service:" + config.Service,
			"version:" + version.Version,
		}),
	)

	if err == nil {
		dd.statsdClientStop = make(chan struct{})
		go dd.runMetricsCollector()
	} else {
		slog.Warn(fmt.Sprintf("can't initialize DogStatsD client: %s", err))
	}

	return dd, nil
}

// Enabled returns true if DataDog is enabled
func (dd *DataDog) Enabled() bool {
	return dd.config.Enabled()
}

// Stop stops the DataDog tracer and metrics collection
func (dd *DataDog) Stop() {
	if !dd.Enabled() {
		return
	}

	tracer.Stop()

	if dd.statsdClient != nil {
		close(dd.statsdClientStop)
		dd.statsdClient.Close()
	}
}

func (dd *DataDog) StartRootSpan(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !dd.Enabled() {
		return ctx, func() {}, rw
	}

	span := tracer.StartSpan(
		"request",
		tracer.Measured(),
		tracer.SpanType("web"),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, r.RequestURI),
	)
	cancel := func() { span.Finish() }

	newRw := httpsnoop.Wrap(rw, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				span.SetTag(ext.HTTPCode, statusCode)
				next(statusCode)
			}
		},
	})

	return context.WithValue(ctx, spanCtxKey{}, span), cancel, newRw
}

func setMetadata(span *tracer.Span, key string, value any) {
	if len(key) == 0 || value == nil {
		return
	}

	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		for _, k := range rv.MapKeys() {
			setMetadata(span, key+"."+k.String(), rv.MapIndex(k).Interface())
		}
		return
	}

	span.SetTag(key, value)
}

func (dd *DataDog) SetMetadata(ctx context.Context, key string, value any) {
	if !dd.Enabled() {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(*tracer.Span); ok {
		setMetadata(rootSpan, key, value)
	}
}

func (dd *DataDog) StartSpan(ctx context.Context, name string, meta map[string]any) context.CancelFunc {
	if !dd.Enabled() {
		return func() {}
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(*tracer.Span); ok {
		span := rootSpan.StartChild(name, tracer.Measured())

		for k, v := range meta {
			setMetadata(span, k, v)
		}

		return func() { span.Finish() }
	}

	return func() {}
}

func (dd *DataDog) SendError(ctx context.Context, errType string, err error) {
	if !dd.Enabled() {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(*tracer.Span); ok {
		rootSpan.SetTag(ext.Error, err)
		rootSpan.SetTag(ext.ErrorType, errformat.FormatErrType(errType, err))
	}
}

func (dd *DataDog) runMetricsCollector() {
	tick := time.NewTicker(dd.config.MetricsInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			dd.statsdClient.Gauge("imgproxy.workers", float64(dd.stats.WorkersNumber), nil, 1)
			dd.statsdClient.Gauge("imgproxy.requests_in_progress", dd.stats.RequestsInProgress(), nil, 1)
			dd.statsdClient.Gauge("imgproxy.images_in_progress", dd.stats.ImagesInProgress(), nil, 1)
			dd.statsdClient.Gauge("imgproxy.workers_utilization", dd.stats.WorkersUtilization(), nil, 1)

			dd.statsdClient.Gauge("imgproxy.vips.memory", vips.GetMem(), nil, 1)
			dd.statsdClient.Gauge("imgproxy.vips.max_memory", vips.GetMemHighwater(), nil, 1)
			dd.statsdClient.Gauge("imgproxy.vips.allocs", vips.GetAllocs(), nil, 1)
		case <-dd.statsdClientStop:
			return
		}
	}
}
