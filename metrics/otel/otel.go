package otel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	ec2 "go.opentelemetry.io/contrib/detectors/aws/ec2/v2"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/detectors/aws/eks"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/metrics/errformat"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/version"
)

type hasSpanCtxKey struct{}

type GaugeFunc func() float64

var (
	enabled        bool
	enabledMetrics bool

	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer

	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter

	propagator propagation.TextMapPropagator

	bufferSizeHist     metric.Int64Histogram
	bufferDefaultSizes = make(map[string]int)
	bufferMaxSizes     = make(map[string]int)
	bufferStatsMutex   sync.Mutex
)

func Init() error {
	mapDeprecatedConfig()

	if !config.OpenTelemetryEnable {
		return nil
	}

	otel.SetErrorHandler(&errorHandler{entry: logrus.WithField("from", "opentelemetry")})

	var (
		traceExporter  *otlptrace.Exporter
		metricExporter sdkmetric.Exporter
		err            error
	)

	protocol := "grpc"
	configurators.String(&protocol, "OTEL_EXPORTER_OTLP_PROTOCOL")

	switch protocol {
	case "grpc":
		traceExporter, metricExporter, err = buildGRPCExporters()
	case "http/protobuf", "http", "https":
		traceExporter, metricExporter, err = buildHTTPExporters()
	default:
		return fmt.Errorf("Unsupported OpenTelemetry protocol: %s", protocol)
	}

	if err != nil {
		return err
	}

	if len(os.Getenv("OTEL_SERVICE_NAME")) == 0 {
		os.Setenv("OTEL_SERVICE_NAME", "imgproxy")
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceVersionKey.String(version.Version),
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
		logrus.Warnf("Can't add AWS attributes to OpenTelemetry: %s", merr)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	}

	switch g := config.OpenTelemetryTraceIDGenerator; g {
	case "xray":
		idg := xray.NewIDGenerator()
		opts = append(opts, sdktrace.WithIDGenerator(idg))
	case "random":
		// Do nothing. OTel uses random generator by default
	default:
		return fmt.Errorf("Unknown Trace ID generator: %s", g)
	}

	tracerProvider = sdktrace.NewTracerProvider(opts...)

	tracer = tracerProvider.Tracer("imgproxy")

	var propagatorNames []string
	configurators.StringSlice(&propagatorNames, "OTEL_PROPAGATORS")

	if len(propagatorNames) > 0 {
		propagator, err = autoprop.TextMapPropagator(propagatorNames...)
		if err != nil {
			return err
		}
	}

	enabled = true

	if metricExporter == nil {
		return nil
	}

	metricReader := sdkmetric.NewPeriodicReader(
		metricExporter,
		sdkmetric.WithInterval(5*time.Second),
	)

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricReader),
	)

	meter = meterProvider.Meter("imgproxy")

	if err = addDefaultMetrics(); err != nil {
		return err
	}

	enabledMetrics = true

	return nil
}

func mapDeprecatedConfig() {
	endpoint := os.Getenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT")
	if len(endpoint) > 0 {
		logger.Deprecated(
			"IMGPROXY_OPEN_TELEMETRY_ENDPOINT",
			"IMGPROXY_OPEN_TELEMETRY_ENABLE and OTEL_EXPORTER_OTLP_ENDPOINT",
			"See https://docs.imgproxy.net/latest/monitoring/open_telemetry#deprecated-environment-variables",
		)
		config.OpenTelemetryEnable = true
	}

	if !config.OpenTelemetryEnable {
		return
	}

	protocol := "grpc"

	if prot := os.Getenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL"); len(prot) > 0 {
		logger.Deprecated(
			"IMGPROXY_OPEN_TELEMETRY_PROTOCOL",
			"OTEL_EXPORTER_OTLP_PROTOCOL",
			"See https://docs.imgproxy.net/latest/monitoring/open_telemetry#deprecated-environment-variables",
		)
		protocol = prot
		os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", protocol)
	}

	if len(endpoint) > 0 {
		schema := "https"

		switch protocol {
		case "grpc":
			if insecure, _ := strconv.ParseBool(os.Getenv("IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE")); insecure {
				logger.Deprecated(
					"IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE",
					"OTEL_EXPORTER_OTLP_ENDPOINT with the `http://` schema",
					"See https://docs.imgproxy.net/latest/monitoring/open_telemetry#deprecated-environment-variables",
				)
				schema = "http"
			}
		case "http":
			schema = "http"
		}

		os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", fmt.Sprintf("%s://%s", schema, endpoint))
	}

	if serviceName := os.Getenv("IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME"); len(serviceName) > 0 {
		logger.Deprecated(
			"IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME",
			"OTEL_SERVICE_NAME",
			"See https://docs.imgproxy.net/latest/monitoring/open_telemetry#deprecated-environment-variables",
		)
		os.Setenv("OTEL_SERVICE_NAME", serviceName)
	}

	if propagators := os.Getenv("IMGPROXY_OPEN_TELEMETRY_PROPAGATORS"); len(propagators) > 0 {
		logger.Deprecated(
			"IMGPROXY_OPEN_TELEMETRY_PROPAGATORS",
			"OTEL_PROPAGATORS",
			"See https://docs.imgproxy.net/latest/monitoring/open_telemetry#deprecated-environment-variables",
		)
		os.Setenv("OTEL_PROPAGATORS", propagators)
	}

	if timeout := os.Getenv("IMGPROXY_OPEN_TELEMETRY_CONNECTION_TIMEOUT"); len(timeout) > 0 {
		logger.Deprecated(
			"IMGPROXY_OPEN_TELEMETRY_CONNECTION_TIMEOUT",
			"OTEL_EXPORTER_OTLP_TIMEOUT",
			"See https://docs.imgproxy.net/latest/monitoring/open_telemetry#deprecated-environment-variables",
		)
		if to, _ := strconv.Atoi(timeout); to > 0 {
			os.Setenv("OTEL_EXPORTER_OTLP_TIMEOUT", strconv.Itoa(to*1000))
		}
	}
}

func buildGRPCExporters() (*otlptrace.Exporter, sdkmetric.Exporter, error) {
	tracerOpts := []otlptracegrpc.Option{}
	meterOpts := []otlpmetricgrpc.Option{}

	if tlsConf, err := buildTLSConfig(); tlsConf != nil && err == nil {
		creds := credentials.NewTLS(tlsConf)
		tracerOpts = append(tracerOpts, otlptracegrpc.WithTLSCredentials(creds))
		meterOpts = append(meterOpts, otlpmetricgrpc.WithTLSCredentials(creds))
	} else if err != nil {
		return nil, nil, err
	}

	tracesConnTimeout, metricsConnTimeout, err := getConnectionTimeouts()
	if err != nil {
		return nil, nil, err
	}

	trctx, trcancel := context.WithTimeout(context.Background(), tracesConnTimeout)
	defer trcancel()

	traceExporter, err := otlptracegrpc.New(trctx, tracerOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	if !config.OpenTelemetryEnableMetrics {
		return traceExporter, nil, err
	}

	mtctx, mtcancel := context.WithTimeout(context.Background(), metricsConnTimeout)
	defer mtcancel()

	metricExporter, err := otlpmetricgrpc.New(mtctx, meterOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	return traceExporter, metricExporter, err
}

func buildHTTPExporters() (*otlptrace.Exporter, sdkmetric.Exporter, error) {
	tracerOpts := []otlptracehttp.Option{}
	meterOpts := []otlpmetrichttp.Option{}

	if tlsConf, err := buildTLSConfig(); tlsConf != nil && err == nil {
		tracerOpts = append(tracerOpts, otlptracehttp.WithTLSClientConfig(tlsConf))
		meterOpts = append(meterOpts, otlpmetrichttp.WithTLSClientConfig(tlsConf))
	} else if err != nil {
		return nil, nil, err
	}

	tracesConnTimeout, metricsConnTimeout, err := getConnectionTimeouts()
	if err != nil {
		return nil, nil, err
	}

	trctx, trcancel := context.WithTimeout(context.Background(), tracesConnTimeout)
	defer trcancel()

	traceExporter, err := otlptracehttp.New(trctx, tracerOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	if !config.OpenTelemetryEnableMetrics {
		return traceExporter, nil, err
	}

	mtctx, mtcancel := context.WithTimeout(context.Background(), metricsConnTimeout)
	defer mtcancel()

	metricExporter, err := otlpmetrichttp.New(mtctx, meterOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	return traceExporter, metricExporter, err
}

func getConnectionTimeouts() (time.Duration, time.Duration, error) {
	connTimeout := 10000
	configurators.Int(&connTimeout, "OTEL_EXPORTER_OTLP_TIMEOUT")

	tracesConnTimeout := connTimeout
	configurators.Int(&tracesConnTimeout, "OTEL_EXPORTER_OTLP_TRACES_TIMEOUT")

	metricsConnTimeout := connTimeout
	configurators.Int(&metricsConnTimeout, "OTEL_EXPORTER_OTLP_METRICS_TIMEOUT")

	if tracesConnTimeout <= 0 {
		return 0, 0, errors.New("Opentelemetry traces timeout should be greater than 0")
	}

	if metricsConnTimeout <= 0 {
		return 0, 0, errors.New("Opentelemetry metrics timeout should be greater than 0")
	}

	return time.Duration(tracesConnTimeout) * time.Millisecond,
		time.Duration(metricsConnTimeout) * time.Millisecond,
		nil
}

func buildTLSConfig() (*tls.Config, error) {
	if len(config.OpenTelemetryServerCert) == 0 {
		return nil, nil
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(prepareKeyCert(config.OpenTelemetryServerCert)) {
		return nil, errors.New("Can't load OpenTelemetry server cert")
	}

	tlsConf := tls.Config{RootCAs: certPool}

	if len(config.OpenTelemetryClientCert) > 0 && len(config.OpenTelemetryClientKey) > 0 {
		cert, err := tls.X509KeyPair(
			prepareKeyCert(config.OpenTelemetryClientCert),
			prepareKeyCert(config.OpenTelemetryClientKey),
		)
		if err != nil {
			return nil, fmt.Errorf("Can't load OpenTelemetry client cert/key pair: %s", err)
		}

		tlsConf.Certificates = []tls.Certificate{cert}
	}

	return &tlsConf, nil
}

func prepareKeyCert(str string) []byte {
	return []byte(strings.ReplaceAll(str, `\n`, "\n"))
}

func Stop() {
	if enabled {
		trctx, trcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer trcancel()

		tracerProvider.Shutdown(trctx)

		if meterProvider != nil {
			mtctx, mtcancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer mtcancel()

			meterProvider.Shutdown(mtctx)
		}
	}
}

func Enabled() bool {
	return enabled
}

func StartRootSpan(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !enabled {
		return ctx, func() {}, rw
	}

	if propagator != nil {
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
	}

	server := r.Host
	if len(server) == 0 {
		server = "imgproxy"
	}

	ctx, span := tracer.Start(
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

func SetMetadata(ctx context.Context, key string, value interface{}) {
	if !enabled {
		return
	}

	if ctx.Value(hasSpanCtxKey{}) != nil {
		if span := trace.SpanFromContext(ctx); span != nil {
			setMetadata(span, key, value)
		}
	}
}

func StartSpan(ctx context.Context, name string, meta map[string]any) context.CancelFunc {
	if !enabled {
		return func() {}
	}

	if ctx.Value(hasSpanCtxKey{}) != nil {
		_, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))

		for k, v := range meta {
			setMetadata(span, k, v)
		}

		return func() { span.End() }
	}

	return func() {}
}

func SendError(ctx context.Context, errType string, err error) {
	if !enabled {
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

func addDefaultMetrics() error {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return fmt.Errorf("Can't initialize process data for OpenTelemetry: %s", err)
	}

	processResidentMemory, err := meter.Int64ObservableGauge(
		"process_resident_memory_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Resident memory size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("Can't add process_resident_memory_bytes gauge to OpenTelemetry: %s", err)
	}

	processVirtualMemory, err := meter.Int64ObservableGauge(
		"process_virtual_memory_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Virtual memory size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("Can't add process_virtual_memory_bytes gauge to OpenTelemetry: %s", err)
	}

	goMemstatsSys, err := meter.Int64ObservableGauge(
		"go_memstats_sys_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Number of bytes obtained from system."),
	)
	if err != nil {
		return fmt.Errorf("Can't add go_memstats_sys_bytes gauge to OpenTelemetry: %s", err)
	}

	goMemstatsHeapIdle, err := meter.Int64ObservableGauge(
		"go_memstats_heap_idle_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Number of heap bytes waiting to be used."),
	)
	if err != nil {
		return fmt.Errorf("Can't add go_memstats_heap_idle_bytes gauge to OpenTelemetry: %s", err)
	}

	goMemstatsHeapInuse, err := meter.Int64ObservableGauge(
		"go_memstats_heap_inuse_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Number of heap bytes that are in use."),
	)
	if err != nil {
		return fmt.Errorf("Can't add go_memstats_heap_inuse_bytes gauge to OpenTelemetry: %s", err)
	}

	goGoroutines, err := meter.Int64ObservableGauge(
		"go_goroutines",
		metric.WithUnit("1"),
		metric.WithDescription("Number of goroutines that currently exist."),
	)
	if err != nil {
		return fmt.Errorf("Can't add go_goroutines gauge to OpenTelemetry: %s", err)
	}

	goThreads, err := meter.Int64ObservableGauge(
		"go_threads",
		metric.WithUnit("1"),
		metric.WithDescription("Number of OS threads created."),
	)
	if err != nil {
		return fmt.Errorf("Can't add go_threads gauge to OpenTelemetry: %s", err)
	}

	workersGauge, err := meter.Int64ObservableGauge(
		"workers",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of running workers."),
	)
	if err != nil {
		return fmt.Errorf("Can't add workets gauge to OpenTelemetry: %s", err)
	}

	requestsInProgressGauge, err := meter.Float64ObservableGauge(
		"requests_in_progress",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of requests currently being in progress."),
	)
	if err != nil {
		return fmt.Errorf("Can't add requests_in_progress gauge to OpenTelemetry: %s", err)
	}

	imagesInProgressGauge, err := meter.Float64ObservableGauge(
		"images_in_progress",
		metric.WithUnit("1"),
		metric.WithDescription("A gauge of the number of images currently being in progress."),
	)
	if err != nil {
		return fmt.Errorf("Can't add images_in_progress gauge to OpenTelemetry: %s", err)
	}

	workersUtilizationGauge, err := meter.Float64ObservableGauge(
		"workers_utilization",
		metric.WithUnit("%"),
		metric.WithDescription("A gauge of the workers utilization in percents."),
	)
	if err != nil {
		return fmt.Errorf("Can't add workers_utilization gauge to OpenTelemetry: %s", err)
	}

	bufferDefaultSizeGauge, err := meter.Int64ObservableGauge(
		"buffer_default_size_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("A gauge of the buffer default size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("Can't add buffer_default_size_bytes gauge to OpenTelemetry: %s", err)
	}

	bufferMaxSizeGauge, err := meter.Int64ObservableGauge(
		"buffer_max_size_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("A gauge of the buffer max size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("Can't add buffer_max_size_bytes gauge to OpenTelemetry: %s", err)
	}

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			memStats, merr := proc.MemoryInfo()
			if merr != nil {
				return merr
			}

			o.ObserveInt64(processResidentMemory, int64(memStats.RSS))
			o.ObserveInt64(processVirtualMemory, int64(memStats.VMS))

			goMemStats := &runtime.MemStats{}
			runtime.ReadMemStats(goMemStats)

			o.ObserveInt64(goMemstatsSys, int64(goMemStats.Sys))
			o.ObserveInt64(goMemstatsHeapIdle, int64(goMemStats.HeapIdle))
			o.ObserveInt64(goMemstatsHeapInuse, int64(goMemStats.HeapInuse))

			threadsNum, _ := runtime.ThreadCreateProfile(nil)
			o.ObserveInt64(goGoroutines, int64(runtime.NumGoroutine()))
			o.ObserveInt64(goThreads, int64(threadsNum))

			o.ObserveInt64(workersGauge, int64(config.Workers))
			o.ObserveFloat64(requestsInProgressGauge, stats.RequestsInProgress())
			o.ObserveFloat64(imagesInProgressGauge, stats.ImagesInProgress())
			o.ObserveFloat64(workersUtilizationGauge, stats.WorkersUtilization())

			bufferStatsMutex.Lock()
			defer bufferStatsMutex.Unlock()

			for t, v := range bufferDefaultSizes {
				o.ObserveInt64(bufferDefaultSizeGauge, int64(v), metric.WithAttributes(attribute.String("type", t)))
			}
			for t, v := range bufferMaxSizes {
				o.ObserveInt64(bufferMaxSizeGauge, int64(v), metric.WithAttributes(attribute.String("type", t)))
			}
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
		bufferDefaultSizeGauge,
		bufferMaxSizeGauge,
	)
	if err != nil {
		return fmt.Errorf("Can't register OpenTelemetry callbacks: %s", err)
	}

	bufferSizeHist, err = meter.Int64Histogram(
		"buffer_size_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("A histogram of the buffer size in bytes."),
	)
	if err != nil {
		return fmt.Errorf("Can't add buffer_size_bytes histogram to OpenTelemetry: %s", err)
	}

	return nil
}

func AddGaugeFunc(name, desc, u string, f GaugeFunc) {
	if meter == nil {
		return
	}

	_, err := meter.Float64ObservableGauge(
		name,
		metric.WithUnit(u),
		metric.WithDescription(desc),
		metric.WithFloat64Callback(func(_ context.Context, obsrv metric.Float64Observer) error {
			obsrv.Observe(f())
			return nil
		}),
	)
	if err != nil {
		logrus.Warnf("Can't add %s gauge to OpenTelemetry: %s", name, err)
	}
}

func ObserveBufferSize(t string, size int) {
	if enabledMetrics {
		bufferSizeHist.Record(context.Background(), int64(size), metric.WithAttributes(attribute.String("type", t)))
	}
}

func SetBufferDefaultSize(t string, size int) {
	if enabledMetrics {
		bufferStatsMutex.Lock()
		defer bufferStatsMutex.Unlock()

		bufferDefaultSizes[t] = size
	}
}

func SetBufferMaxSize(t string, size int) {
	if enabledMetrics {
		bufferStatsMutex.Lock()
		defer bufferStatsMutex.Unlock()

		bufferMaxSizes[t] = size
	}
}

type errorHandler struct {
	entry *logrus.Entry
}

func (h *errorHandler) Handle(err error) {
	h.entry.Warn(err.Error())
}
