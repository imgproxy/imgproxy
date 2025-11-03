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

	"github.com/imgproxy/imgproxy/v3/monitoring/format"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/version"
	vipsstats "github.com/imgproxy/imgproxy/v3/vips/stats"
)

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
	if !config.Enabled() {
		return nil, nil
	}

	dd := &DataDog{
		stats:  stats,
		config: config,
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

// Stop stops the DataDog tracer and metrics collection
func (dd *DataDog) Stop(ctx context.Context) {
	tracer.Stop()

	if dd.statsdClient != nil {
		close(dd.statsdClientStop)
		dd.statsdClient.Close()
	}
}

// StartRequest starts a new DataDog span for the incoming HTTP request
func (dd *DataDog) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	// Extract parent span context from incoming request headers if any
	parentSpanCtx, err := tracer.Extract(tracer.HTTPHeadersCarrier(r.Header))
	if err != nil {
		parentSpanCtx = nil
	}

	span := tracer.StartSpan(
		"request",
		tracer.Measured(),
		tracer.SpanType("web"),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, r.RequestURI),
		// ChildOf is deprecated, but there's no alternative yet,
		// and DD devs recommend to keep using it:
		// https://github.com/DataDog/dd-trace-go/issues/3598
		//nolint:staticcheck
		tracer.ChildOf(parentSpanCtx),
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

	return tracer.ContextWithSpan(ctx, span), cancel, newRw
}

// setMetadata sets metadata on the given DataDog span
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

// SetMetadata sets metadata for the current span
func (dd *DataDog) SetMetadata(ctx context.Context, key string, value any) {
	if span, ok := tracer.SpanFromContext(ctx); ok {
		setMetadata(span, key, value)
	}
}

// StartSpan starts a new span for DataDog monitoring
func (dd *DataDog) StartSpan(
	ctx context.Context,
	name string,
	meta map[string]any,
) (context.Context, context.CancelFunc) {
	if rootSpan, ok := tracer.SpanFromContext(ctx); ok {
		span := rootSpan.StartChild(format.FormatSegmentName(name), tracer.Measured())

		for k, v := range meta {
			setMetadata(span, k, v)
		}

		ctx = tracer.ContextWithSpan(ctx, span)

		return ctx, func() { span.Finish() }
	}

	return ctx, func() {}
}

// SendError sends an error to DataDog APM
func (dd *DataDog) SendError(ctx context.Context, errType string, err error) {
	if span, ok := tracer.SpanFromContext(ctx); ok {
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorType, format.FormatErrType(errType, err))
	}
}

// InjectHeaders adds monitoring headers to the provided HTTP headers.
func (dd *DataDog) InjectHeaders(ctx context.Context, headers http.Header) {
	if !dd.config.PropagateExt {
		return
	}

	if span, ok := tracer.SpanFromContext(ctx); ok {
		carrier := tracer.HTTPHeadersCarrier(headers)
		tracer.Inject(span.Context(), carrier)
	}
}

// runMetricsCollector periodically collects and sends custom metrics to DataDog
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

			dd.statsdClient.Gauge("imgproxy.vips.memory", vipsstats.Memory(), nil, 1)
			dd.statsdClient.Gauge("imgproxy.vips.max_memory", vipsstats.MemoryHighwater(), nil, 1)
			dd.statsdClient.Gauge("imgproxy.vips.allocs", vipsstats.Allocs(), nil, 1)
		case <-dd.statsdClientStop:
			return
		}
	}
}
