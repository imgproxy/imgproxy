package otel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/detectors/aws/eks"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
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
)

func Init() error {
	if len(config.OpenTelemetryEndpoint) == 0 {
		return nil
	}

	otel.SetErrorHandler(&errorHandler{entry: logrus.WithField("from", "opentelemetry")})

	var (
		traceExporter  *otlptrace.Exporter
		metricExporter sdkmetric.Exporter
		err            error
	)

	switch config.OpenTelemetryProtocol {
	case "grpc":
		traceExporter, metricExporter, err = buildGRPCExporters()
	case "https":
		traceExporter, metricExporter, err = buildHTTPExporters(false)
	case "http":
		traceExporter, metricExporter, err = buildHTTPExporters(true)
	default:
		return fmt.Errorf("Unknown OpenTelemetry protocol: %s", config.OpenTelemetryProtocol)
	}

	if err != nil {
		return err
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String(config.OpenTelemetryServiceName),
			semconv.ServiceVersionKey.String(version.Version()),
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

	if opts, err = addTraceIDRatioSampler(opts); err != nil {
		return err
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

	if len(config.OpenTelemetryPropagators) > 0 {
		propagator, err = autoprop.TextMapPropagator(config.OpenTelemetryPropagators...)
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

	enabledMetrics = true

	AddGaugeFunc(
		"requests_in_progress",
		"A gauge of the number of requests currently being in progress.",
		"1",
		stats.RequestsInProgress,
	)
	AddGaugeFunc(
		"images_in_progress",
		"A gauge of the number of images currently being in progress.",
		"1",
		stats.ImagesInProgress,
	)

	return nil
}

func buildGRPCExporters() (*otlptrace.Exporter, sdkmetric.Exporter, error) {
	tracerOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.OpenTelemetryEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	}

	meterOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(config.OpenTelemetryEndpoint),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	}

	tlsConf, err := buildTLSConfig()
	if err != nil {
		return nil, nil, err
	}

	if tlsConf != nil {
		creds := credentials.NewTLS(tlsConf)
		tracerOpts = append(tracerOpts, otlptracegrpc.WithTLSCredentials(creds))
		meterOpts = append(meterOpts, otlpmetricgrpc.WithTLSCredentials(creds))
	} else if config.OpenTelemetryGRPCInsecure {
		tracerOpts = append(tracerOpts, otlptracegrpc.WithInsecure())
		meterOpts = append(meterOpts, otlpmetricgrpc.WithInsecure())
	}

	trctx, trcancel := context.WithTimeout(
		context.Background(),
		time.Duration(config.OpenTelemetryConnectionTimeout)*time.Second,
	)
	defer trcancel()

	traceExporter, err := otlptracegrpc.New(trctx, tracerOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	if !config.OpenTelemetryEnableMetrics {
		return traceExporter, nil, err
	}

	mtctx, mtcancel := context.WithTimeout(
		context.Background(),
		time.Duration(config.OpenTelemetryConnectionTimeout)*time.Second,
	)
	defer mtcancel()

	metricExporter, err := otlpmetricgrpc.New(mtctx, meterOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	return traceExporter, metricExporter, err
}

func buildHTTPExporters(insecure bool) (*otlptrace.Exporter, sdkmetric.Exporter, error) {
	tracerOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.OpenTelemetryEndpoint),
	}

	meterOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(config.OpenTelemetryEndpoint),
	}

	if insecure {
		tracerOpts = append(tracerOpts, otlptracehttp.WithInsecure())
		meterOpts = append(meterOpts, otlpmetrichttp.WithInsecure())
	} else {
		tlsConf, err := buildTLSConfig()
		if err != nil {
			return nil, nil, err
		}

		if tlsConf != nil {
			tracerOpts = append(tracerOpts, otlptracehttp.WithTLSClientConfig(tlsConf))
			meterOpts = append(meterOpts, otlpmetrichttp.WithTLSClientConfig(tlsConf))
		}
	}

	trctx, trcancel := context.WithTimeout(
		context.Background(),
		time.Duration(config.OpenTelemetryConnectionTimeout)*time.Second,
	)
	defer trcancel()

	traceExporter, err := otlptracehttp.New(trctx, tracerOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	if !config.OpenTelemetryEnableMetrics {
		return traceExporter, nil, err
	}

	mtctx, mtcancel := context.WithTimeout(
		context.Background(),
		time.Duration(config.OpenTelemetryConnectionTimeout)*time.Second,
	)
	defer mtcancel()

	metricExporter, err := otlpmetrichttp.New(mtctx, meterOpts...)
	if err != nil {
		err = fmt.Errorf("Can't connect to OpenTelemetry collector: %s", err)
	}

	return traceExporter, metricExporter, err
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

	ctx, span := tracer.Start(
		ctx, "/request",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("imgproxy", "/", r)...),
	)
	ctx = context.WithValue(ctx, hasSpanCtxKey{}, struct{}{})

	newRw := httpsnoop.Wrap(rw, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				attrs := semconv.HTTPAttributesFromHTTPStatusCode(statusCode)
				spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(statusCode, trace.SpanKindServer)
				span.SetAttributes(attrs...)
				span.SetStatus(spanStatus, spanMessage)

				next(statusCode)
			}
		},
	})

	cancel := func() { span.End() }
	return ctx, cancel, newRw
}

func StartSpan(ctx context.Context, name string) context.CancelFunc {
	if !enabled {
		return func() {}
	}

	if ctx.Value(hasSpanCtxKey{}) != nil {
		_, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))

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

	span.AddEvent(semconv.ExceptionEventName, trace.WithAttributes(attributes...))
}

func AddGaugeFunc(name, desc, u string, f GaugeFunc) {
	if !enabledMetrics {
		return
	}

	gauge, _ := meter.AsyncFloat64().Gauge(
		name,
		instrument.WithUnit(unit.Unit(u)),
		instrument.WithDescription(desc),
	)

	if err := meter.RegisterCallback(
		[]instrument.Asynchronous{
			gauge,
		},
		func(ctx context.Context) {
			gauge.Observe(ctx, f())
		},
	); err != nil {
		logrus.Warnf("Can't add %s gauge to OpenTelemetry: %s", name, err)
	}
}

type errorHandler struct {
	entry *logrus.Entry
}

func (h *errorHandler) Handle(err error) {
	h.entry.Warn(err.Error())
}
