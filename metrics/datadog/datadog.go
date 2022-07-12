package datadog

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/metrics/errformat"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/version"
)

type spanCtxKey struct{}

type GaugeFunc func() float64

var (
	enabled        bool
	enabledMetrics bool

	statsdClient     *statsd.Client
	statsdClientStop chan struct{}

	gaugeFuncs      = make(map[string]GaugeFunc)
	gaugeFuncsMutex sync.RWMutex
)

func Init() {
	if !config.DataDogEnable {
		return
	}

	name := os.Getenv("DD_SERVICE")
	if len(name) == 0 {
		name = "imgproxy"
	}

	tracer.Start(
		tracer.WithService(name),
		tracer.WithServiceVersion(version.Version()),
		tracer.WithLogger(dataDogLogger{}),
	)

	enabled = true

	statsdHost, statsdPort := os.Getenv("DD_AGENT_HOST"), os.Getenv("DD_DOGSTATSD_PORT")
	if len(statsdHost) == 0 {
		statsdHost = "localhost"
	}
	if len(statsdPort) == 0 {
		statsdPort = "8125"
	}

	if !config.DataDogEnableMetrics {
		return
	}

	var err error
	statsdClient, err = statsd.New(
		net.JoinHostPort(statsdHost, statsdPort),
		statsd.WithTags([]string{
			"service:" + name,
			"version:" + version.Version(),
		}),
	)
	if err == nil {
		statsdClientStop = make(chan struct{})
		enabledMetrics = true
		go runMetricsCollector()
	} else {
		log.Warnf("Can't initialize DogStatsD client: %s", err)
	}
}

func Stop() {
	if enabled {
		tracer.Stop()

		if statsdClient != nil {
			close(statsdClientStop)
			statsdClient.Close()
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

	span := tracer.StartSpan(
		"request",
		tracer.Measured(),
		tracer.SpanType("web"),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, r.RequestURI),
	)
	cancel := func() { span.Finish() }
	newRw := dataDogResponseWriter{rw, span}

	return context.WithValue(ctx, spanCtxKey{}, span), cancel, newRw
}

func StartSpan(ctx context.Context, name string) context.CancelFunc {
	if !enabled {
		return func() {}
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(tracer.Span); ok {
		span := tracer.StartSpan(name, tracer.Measured(), tracer.ChildOf(rootSpan.Context()))
		return func() { span.Finish() }
	}

	return func() {}
}

func SendError(ctx context.Context, errType string, err error) {
	if !enabled {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(tracer.Span); ok {
		rootSpan.SetTag(ext.Error, err)
		rootSpan.SetTag(ext.ErrorType, errformat.FormatErrType(errType, err))
	}
}

func AddGaugeFunc(name string, f GaugeFunc) {
	gaugeFuncsMutex.Lock()
	defer gaugeFuncsMutex.Unlock()

	gaugeFuncs["imgproxy."+name] = f
}

func ObserveBufferSize(t string, size int) {
	if enabledMetrics {
		statsdClient.Histogram("imgproxy.buffer.size", float64(size), []string{"type:" + t}, 1)
	}
}

func SetBufferDefaultSize(t string, size int) {
	if enabledMetrics {
		statsdClient.Gauge("imgproxy.buffer.default_size", float64(size), []string{"type:" + t}, 1)
	}
}

func SetBufferMaxSize(t string, size int) {
	if enabledMetrics {
		statsdClient.Gauge("imgproxy.buffer.max_size", float64(size), []string{"type:" + t}, 1)
	}
}

func runMetricsCollector() {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			func() {
				gaugeFuncsMutex.RLock()
				defer gaugeFuncsMutex.RUnlock()

				for name, f := range gaugeFuncs {
					statsdClient.Gauge(name, f(), nil, 1)
				}
			}()

			statsdClient.Gauge("imgproxy.requests_in_progress", stats.RequestsInProgress(), nil, 1)
			statsdClient.Gauge("imgproxy.images_in_progress", stats.ImagesInProgress(), nil, 1)
		case <-statsdClientStop:
			return
		}
	}
}

type dataDogLogger struct {
}

func (l dataDogLogger) Log(msg string) {
	log.Info(msg)
}

type dataDogResponseWriter struct {
	rw   http.ResponseWriter
	span tracer.Span
}

func (ddrw dataDogResponseWriter) Header() http.Header {
	return ddrw.rw.Header()
}
func (ddrw dataDogResponseWriter) Write(data []byte) (int, error) {
	return ddrw.rw.Write(data)
}
func (ddrw dataDogResponseWriter) WriteHeader(statusCode int) {
	ddrw.span.SetTag(ext.HTTPCode, statusCode)
	ddrw.rw.WriteHeader(statusCode)
}
