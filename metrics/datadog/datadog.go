package datadog

import (
	"context"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/felixge/httpsnoop"
	log "github.com/sirupsen/logrus"

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

	logStartup := false
	if b, err := strconv.ParseBool(os.Getenv("DD_TRACE_STARTUP_LOGS")); err == nil {
		logStartup = b
	}

	tracer.Start(
		tracer.WithService(name),
		tracer.WithServiceVersion(version.Version),
		tracer.WithLogger(dataDogLogger{}),
		tracer.WithLogStartup(logStartup),
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
			"version:" + version.Version,
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

func SetMetadata(ctx context.Context, key string, value any) {
	if !enabled {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(*tracer.Span); ok {
		setMetadata(rootSpan, key, value)
	}
}

func StartSpan(ctx context.Context, name string, meta map[string]any) context.CancelFunc {
	if !enabled {
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

func SendError(ctx context.Context, errType string, err error) {
	if !enabled {
		return
	}

	if rootSpan, ok := ctx.Value(spanCtxKey{}).(*tracer.Span); ok {
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

			statsdClient.Gauge("imgproxy.workers", float64(config.Workers), nil, 1)
			statsdClient.Gauge("imgproxy.requests_in_progress", stats.RequestsInProgress(), nil, 1)
			statsdClient.Gauge("imgproxy.images_in_progress", stats.ImagesInProgress(), nil, 1)
			statsdClient.Gauge("imgproxy.workers_utilization", stats.WorkersUtilization(), nil, 1)
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
