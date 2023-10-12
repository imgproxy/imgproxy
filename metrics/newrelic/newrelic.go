package newrelic

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/metrics/errformat"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
)

type transactionCtxKey struct{}

type GaugeFunc func() float64

const (
	defaultMetricURL = "https://metric-api.newrelic.com/metric/v1"
	euMetricURL      = "https://metric-api.eu.newrelic.com/metric/v1"
)

var (
	enabled          = false
	enabledHarvester = false

	app       *newrelic.Application
	harvester *telemetry.Harvester

	harvesterCtx       context.Context
	harvesterCtxCancel context.CancelFunc

	gaugeFuncs      = make(map[string]GaugeFunc)
	gaugeFuncsMutex sync.RWMutex

	bufferSummaries      = make(map[string]*telemetry.Summary)
	bufferSummariesMutex sync.RWMutex

	interval = 10 * time.Second

	licenseEuRegex = regexp.MustCompile(`(^eu.+?)x`)
)

func Init() error {
	if len(config.NewRelicKey) == 0 {
		return nil
	}

	name := config.NewRelicAppName
	if len(name) == 0 {
		name = "imgproxy"
	}

	var err error

	app, err = newrelic.NewApplication(
		newrelic.ConfigAppName(name),
		newrelic.ConfigLicense(config.NewRelicKey),
		func(c *newrelic.Config) {
			if len(config.NewRelicLabels) > 0 {
				c.Labels = config.NewRelicLabels
			}
		},
	)

	if err != nil {
		return fmt.Errorf("Can't init New Relic agent: %s", err)
	}

	harvesterAttributes := map[string]interface{}{"appName": name}
	for k, v := range config.NewRelicLabels {
		harvesterAttributes[k] = v
	}

	metricsURL := defaultMetricURL
	if licenseEuRegex.MatchString(config.NewRelicKey) {
		metricsURL = euMetricURL
	}

	harvester, err = telemetry.NewHarvester(
		telemetry.ConfigAPIKey(config.NewRelicKey),
		telemetry.ConfigCommonAttributes(harvesterAttributes),
		telemetry.ConfigHarvestPeriod(0), // Don't harvest automatically
		telemetry.ConfigMetricsURLOverride(metricsURL),
		telemetry.ConfigBasicErrorLogger(log.StandardLogger().WithField("from", "newrelic").WriterLevel(log.WarnLevel)),
	)
	if err == nil {
		harvesterCtx, harvesterCtxCancel = context.WithCancel(context.Background())
		enabledHarvester = true
		go runMetricsCollector()
	} else {
		log.Warnf("Can't init New Relic telemetry harvester: %s", err)
	}

	enabled = true

	return nil
}

func Stop() {
	if enabled {
		app.Shutdown(5 * time.Second)

		if enabledHarvester {
			harvesterCtxCancel()
			harvester.HarvestNow(context.Background())
		}
	}
}

func Enabled() bool {
	return enabled
}

func StartTransaction(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !enabled {
		return ctx, func() {}, rw
	}

	txn := app.StartTransaction("request")
	txn.SetWebRequestHTTP(r)
	newRw := txn.SetWebResponse(rw)
	cancel := func() { txn.End() }
	return context.WithValue(ctx, transactionCtxKey{}, txn), cancel, newRw
}

func StartSegment(ctx context.Context, name string) context.CancelFunc {
	if !enabled {
		return func() {}
	}

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		segment := txn.StartSegment(name)
		return func() { segment.End() }
	}

	return func() {}
}

func SendError(ctx context.Context, errType string, err error) {
	if !enabled {
		return
	}

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		txn.NoticeError(newrelic.Error{
			Message: err.Error(),
			Class:   errformat.FormatErrType(errType, err),
		})
	}
}

func AddGaugeFunc(name string, f GaugeFunc) {
	gaugeFuncsMutex.Lock()
	defer gaugeFuncsMutex.Unlock()

	gaugeFuncs["imgproxy."+name] = f
}

func ObserveBufferSize(t string, size int) {
	if enabledHarvester {
		bufferSummariesMutex.Lock()
		defer bufferSummariesMutex.Unlock()

		summary, ok := bufferSummaries[t]
		if !ok {
			summary = &telemetry.Summary{
				Name:       "imgproxy.buffer.size",
				Attributes: map[string]interface{}{"buffer_type": t},
				Timestamp:  time.Now(),
			}
			bufferSummaries[t] = summary
		}

		sizef := float64(size)

		summary.Count += 1
		summary.Sum += sizef
		summary.Min = math.Min(summary.Min, sizef)
		summary.Max = math.Max(summary.Max, sizef)
	}
}

func SetBufferDefaultSize(t string, size int) {
	if enabledHarvester {
		harvester.RecordMetric(telemetry.Gauge{
			Name:       "imgproxy.buffer.default_size",
			Value:      float64(size),
			Attributes: map[string]interface{}{"buffer_type": t},
			Timestamp:  time.Now(),
		})
	}
}

func SetBufferMaxSize(t string, size int) {
	if enabledHarvester {
		harvester.RecordMetric(telemetry.Gauge{
			Name:       "imgproxy.buffer.max_size",
			Value:      float64(size),
			Attributes: map[string]interface{}{"buffer_type": t},
			Timestamp:  time.Now(),
		})
	}
}

func runMetricsCollector() {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			func() {
				gaugeFuncsMutex.RLock()
				defer gaugeFuncsMutex.RUnlock()

				for name, f := range gaugeFuncs {
					harvester.RecordMetric(telemetry.Gauge{
						Name:      name,
						Value:     f(),
						Timestamp: time.Now(),
					})
				}
			}()

			func() {
				bufferSummariesMutex.RLock()
				defer bufferSummariesMutex.RUnlock()

				now := time.Now()

				for _, summary := range bufferSummaries {
					summary.Interval = now.Sub(summary.Timestamp)
					harvester.RecordMetric(*summary)

					summary.Timestamp = now
					summary.Count = 0
					summary.Sum = 0
					summary.Min = 0
					summary.Max = 0
				}
			}()

			harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.requests_in_progress",
				Value:     stats.RequestsInProgress(),
				Timestamp: time.Now(),
			})

			harvester.RecordMetric(telemetry.Gauge{
				Name:      "imgproxy.images_in_progress",
				Value:     stats.ImagesInProgress(),
				Timestamp: time.Now(),
			})

			harvester.HarvestNow(harvesterCtx)
		case <-harvesterCtx.Done():
			return
		}
	}
}
