package otel

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/shirou/gopsutil/process"
	ec2 "go.opentelemetry.io/contrib/detectors/aws/ec2/v2"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/detectors/aws/eks"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/monitoring/errformat"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/imgproxy/imgproxy/v3/vips"
)

const (
	// stopTimeout is the maximum time to wait for the shutdown of the tracer and meter providers
	stopTimeout = 5 * time.Second

	// defaultOtelServiceName is the default service name for OpenTelemetry if none is set
	defaultOtelServiceName = "imgproxy"
)

// hasSpanCtxKey is a context key to mark that there is a span in the context
type hasSpanCtxKey struct{}

// errorHandler is an implementation of the OpenTelemetry error handler interface
type errorHandler struct{}

func (h errorHandler) Handle(err error) {
	slog.Warn(err.Error(), "source", "opentelemetry")
}

// Otel holds OpenTelemetry tracer and meter providers and configuration
type Otel struct {
	config *Config
	stats  *stats.Stats

	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer

	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter

	propagator propagation.TextMapPropagator
}

// New creates a new Otel instance
func New(config *Config, stats *stats.Stats) (*Otel, error) {
	o := &Otel{
		config: config,
		stats:  stats,
	}

	if !config.Enabled() {
		return o, nil
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	otel.SetErrorHandler(errorHandler{})

	traceExporter, metricExporter, err := buildProtocolExporter(config)
	if err != nil {
		return nil, err
	}

	// If no service name is set, use "imgproxy" as default, and write it into the environment
	if n, _ := OTEL_SERVICE_NAME.Get(); len(n) == 0 {
		os.Setenv(OTEL_SERVICE_NAME.Name, defaultOtelServiceName)
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceVersion(version.Version),
		),
	)

	awsRes, _ := resource.Detect(
		context.Background(),
		ec2.NewResourceDetector(),
		ecs.NewResourceDetector(),
		eks.NewResourceDetector(),
	)

	if merged, merr := resource.Merge(awsRes, res); merr == nil {
		res = merged
	} else {
		slog.Warn(fmt.Sprintf("can't add AWS attributes to OpenTelemetry: %s", merr))
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	}

	switch g := config.TraceIDGenerator; g {
	case "xray":
		idg := xray.NewIDGenerator()
		opts = append(opts, sdktrace.WithIDGenerator(idg))
	case "random":
		// Do nothing. OTel uses random generator by default
	default:
		return nil, fmt.Errorf("unknown Trace ID generator: %s", g)
	}

	o.tracerProvider = sdktrace.NewTracerProvider(opts...)
	o.tracer = o.tracerProvider.Tracer("imgproxy")

	if len(config.Propagators) > 0 {
		o.propagator, err = autoprop.TextMapPropagator(config.Propagators...)
		if err != nil {
			return nil, err
		}
	}

	if metricExporter == nil {
		return o, nil
	}

	metricReader := sdkmetric.NewPeriodicReader(
		metricExporter,
		sdkmetric.WithInterval(config.MetricsInterval),
	)

	o.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricReader),
	)

	o.meter = o.meterProvider.Meter("imgproxy")

	if err = o.addDefaultMetrics(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Otel) Enabled() bool {
	return o.config.Enabled()
}

func (o *Otel) Stop(ctx context.Context) {
	if o.tracerProvider != nil {
		trctx, trcancel := context.WithTimeout(ctx, stopTimeout)
		defer trcancel()

		o.tracerProvider.Shutdown(trctx)
	}

	if o.meterProvider != nil {
		mtctx, mtcancel := context.WithTimeout(ctx, stopTimeout)
		defer mtcancel()

		o.meterProvider.Shutdown(mtctx)
	}
}

func (o *Otel) StartRootSpan(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !o.Enabled() {
		return ctx, func() {}, rw
	}

	if o.propagator != nil {
		ctx = o.propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
	}

	server := r.Host
	if len(server) == 0 {
		server = "imgproxy"
	}

	ctx, span := o.tracer.Start(
		ctx, "/request",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(httpconv.ServerRequest(server, r)...),
		trace.WithAttributes(semconv.HTTPURL(r.RequestURI)),
	)
	ctx = context.WithValue(ctx, hasSpanCtxKey{}, struct{}{})

	newRw := httpsnoop.Wrap(rw, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				span.SetStatus(httpconv.ServerStatus(statusCode))
				span.SetAttributes(semconv.HTTPStatusCode(statusCode))

				next(statusCode)
			}
		},
	})

	cancel := func() { span.End() }
	return ctx, cancel, newRw
}

func setMetadata(span trace.Span, key string, value interface{}) {
	if len(key) == 0 || value == nil {
		return
	}

	if stringer, ok := value.(fmt.Stringer); ok {
		span.SetAttributes(attribute.String(key, stringer.String()))
		return
	}

	rv := reflect.ValueOf(value)

	switch {
	case rv.Kind() == reflect.String:
		span.SetAttributes(attribute.String(key, value.(string)))
	case rv.Kind() == reflect.Bool:
		span.SetAttributes(attribute.Bool(key, value.(bool)))
	case rv.CanInt():
		span.SetAttributes(attribute.Int64(key, rv.Int()))
	case rv.CanUint():
		span.SetAttributes(attribute.Int64(key, int64(rv.Uint())))
	case rv.CanFloat():
		span.SetAttributes(attribute.Float64(key, rv.Float()))
	case rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String:
		for _, k := range rv.MapKeys() {
			setMetadata(span, key+"."+k.String(), rv.MapIndex(k).Interface())
		}
	default:
		// Theoretically, we can also cover slices and arrays here,
		// but it's pretty complex and not really needed for now
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
	}
}

func (o *Otel) SetMetadata(ctx context.Context, key string, value interface{}) {
	if !o.Enabled() {
		return
	}

	if ctx.Value(hasSpanCtxKey{}) != nil {
		if span := trace.SpanFromContext(ctx); span != nil {
			setMetadata(span, key, value)
		}
	}
}

func (o *Otel) StartSpan(ctx context.Context, name string, meta map[string]any) context.CancelFunc {
	if !o.Enabled() {
		return func() {}
	}

	if ctx.Value(hasSpanCtxKey{}) != nil {
		_, span := o.tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))

		for k, v := range meta {
			setMetadata(span, k, v)
		}

		return func() { span.End() }
	}

	return func() {}
}

func (o *Otel) SendError(ctx context.Context, errType string, err error) {
	if !o.Enabled() {
		return
	}

	span := trace.SpanFromContext(ctx)

	attributes := []attribute.KeyValue{
		semconv.ExceptionTypeKey.String(errformat.FormatErrType(errType, err)),
		semconv.ExceptionMessageKey.String(err.Error()),
	}

	if ierr, ok := err.(*ierrors.Error); ok {
		if stack := ierr.FormatStack(); len(stack) != 0 {
			attributes = append(attributes, semconv.ExceptionStacktraceKey.String(stack))
		}
	}
	span.SetStatus(codes.Error, err.Error())
	span.AddEvent(semconv.ExceptionEventName, trace.WithAttributes(attributes...))
}

func (o *Otel) addDefaultMetrics() error {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return fmt.Errorf("can't initialize process data for OpenTelemetry: %s", err)
	}

	processResidentMemory, err := o.meter.Int64ObservableGauge(
		"process_resident_memory_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Resident memory size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("can't add process_resident_memory_bytes gauge to OpenTelemetry: %s", err)
	}

	processVirtualMemory, err := o.meter.Int64ObservableGauge(
		"process_virtual_memory_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Virtual memory size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("can't add process_virtual_memory_bytes gauge to OpenTelemetry: %s", err)
	}

	goMemstatsSys, err := o.meter.Int64ObservableGauge(
		"go_memstats_sys_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Number of bytes obtained from system."),
	)
	if err != nil {
		return fmt.Errorf("can't add go_memstats_sys_bytes gauge to OpenTelemetry: %s", err)
	}

	goMemstatsHeapIdle, err := o.meter.Int64ObservableGauge(
		"go_memstats_heap_idle_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Number of heap bytes waiting to be used."),
	)
	if err != nil {
		return fmt.Errorf("can't add go_memstats_heap_idle_bytes gauge to OpenTelemetry: %s", err)
	}

	goMemstatsHeapInuse, err := o.meter.Int64ObservableGauge(
		"go_memstats_heap_inuse_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Number of heap bytes that are in use."),
	)
	if err != nil {
		return fmt.Errorf("can't add go_memstats_heap_inuse_bytes gauge to OpenTelemetry: %s", err)
	}

	goGoroutines, err := o.meter.Int64ObservableGauge(
		"go_goroutines",
		metric.WithUnit("1"),
		metric.WithDescription("Number of goroutines that currently exist."),
	)
	if err != nil {
		return fmt.Errorf("can't add go_goroutines gauge to OpenTelemetry: %s", err)
	}

	goThreads, err := o.meter.Int64ObservableGauge(
		"go_threads",
		metric.WithUnit("1"),
		metric.WithDescription("Number of OS threads created."),
	)
	if err != nil {
		return fmt.Errorf("can't add go_threads gauge to OpenTelemetry: %s", err)
	}

	workersGauge, err := o.meter.Int64ObservableGauge(
		"workers",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of running workers."),
	)
	if err != nil {
		return fmt.Errorf("can't add workers gauge to OpenTelemetry: %s", err)
	}

	requestsInProgressGauge, err := o.meter.Float64ObservableGauge(
		"requests_in_progress",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of requests currently being in progress."),
	)
	if err != nil {
		return fmt.Errorf("can't add requests_in_progress gauge to OpenTelemetry: %s", err)
	}

	imagesInProgressGauge, err := o.meter.Float64ObservableGauge(
		"images_in_progress",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of images currently being in progress."),
	)
	if err != nil {
		return fmt.Errorf("can't add images_in_progress gauge to OpenTelemetry: %s", err)
	}

	workersUtilizationGauge, err := o.meter.Float64ObservableGauge(
		"workers_utilization",
		metric.WithUnit("%"),
		metric.WithDescription("A gauge of the workers utilization in percents."),
	)
	if err != nil {
		return fmt.Errorf("can't add workers_utilization gauge to OpenTelemetry: %s", err)
	}

	vipsMemory, err := o.meter.Float64ObservableGauge(
		"vips_memory_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("A gauge of the vips tracked memory usage in bytes."),
	)
	if err != nil {
		return fmt.Errorf("can't add vips_memory_bytes gauge to OpenTelemetry: %s", err)
	}

	vipsMaxMemory, err := o.meter.Float64ObservableGauge(
		"vips_max_memory_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("A gauge of the max vips tracked memory usage in bytes."),
	)
	if err != nil {
		return fmt.Errorf("can't add vips_max_memory_bytes gauge to OpenTelemetry: %s", err)
	}

	vipsAllocs, err := o.meter.Float64ObservableGauge(
		"vips_allocs",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of active vips allocations."),
	)
	if err != nil {
		return fmt.Errorf("can't add vips_allocs gauge to OpenTelemetry: %s", err)
	}

	_, err = o.meter.RegisterCallback(
		func(ctx context.Context, ob metric.Observer) error {
			memStats, merr := proc.MemoryInfo()
			if merr != nil {
				return merr
			}

			ob.ObserveInt64(processResidentMemory, int64(memStats.RSS))
			ob.ObserveInt64(processVirtualMemory, int64(memStats.VMS))

			goMemStats := &runtime.MemStats{}
			runtime.ReadMemStats(goMemStats)

			ob.ObserveInt64(goMemstatsSys, int64(goMemStats.Sys))
			ob.ObserveInt64(goMemstatsHeapIdle, int64(goMemStats.HeapIdle))
			ob.ObserveInt64(goMemstatsHeapInuse, int64(goMemStats.HeapInuse))

			threadsNum, _ := runtime.ThreadCreateProfile(nil)
			ob.ObserveInt64(goGoroutines, int64(runtime.NumGoroutine()))
			ob.ObserveInt64(goThreads, int64(threadsNum))

			ob.ObserveInt64(workersGauge, int64(o.stats.WorkersNumber))
			ob.ObserveFloat64(requestsInProgressGauge, o.stats.RequestsInProgress())
			ob.ObserveFloat64(imagesInProgressGauge, o.stats.ImagesInProgress())
			ob.ObserveFloat64(workersUtilizationGauge, o.stats.WorkersUtilization())

			ob.ObserveFloat64(vipsMemory, vips.GetMem())
			ob.ObserveFloat64(vipsMaxMemory, vips.GetMemHighwater())
			ob.ObserveFloat64(vipsAllocs, vips.GetAllocs())

			return nil
		},
		processResidentMemory,
		processVirtualMemory,
		goMemstatsSys,
		goMemstatsHeapIdle,
		goMemstatsHeapInuse,
		goGoroutines,
		goThreads,
		workersGauge,
		requestsInProgressGauge,
		imagesInProgressGauge,
		workersUtilizationGauge,
		vipsMemory,
		vipsMaxMemory,
		vipsAllocs,
	)
	if err != nil {
		return fmt.Errorf("can't register OpenTelemetry callbacks: %s", err)
	}

	return nil
}
