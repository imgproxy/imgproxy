package newrelic

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/go-agent/internal"
	"github.com/newrelic/go-agent/internal/logger"
)

var (
	// NEW_RELIC_DEBUG_LOGGING can be set to anything to enable additional
	// debug logging: the agent will log every transaction's data at info
	// level.
	envDebugLogging = "NEW_RELIC_DEBUG_LOGGING"
	debugLogging    = os.Getenv(envDebugLogging)
)

type dataConsumer interface {
	Consume(internal.AgentRunID, internal.Harvestable)
}

type appData struct {
	id   internal.AgentRunID
	data internal.Harvestable
}

type app struct {
	config      Config
	rpmControls internal.RpmControls
	testHarvest *internal.Harvest

	// placeholderRun is used when the application is not connected.
	placeholderRun *appRun

	// initiateShutdown is used to tell the processor to shutdown.
	initiateShutdown chan struct{}

	// shutdownStarted and shutdownComplete are closed by the processor
	// goroutine to indicate the shutdown status.  Two channels are used so
	// that the call of app.Shutdown() can block until shutdown has
	// completed but other goroutines can exit when shutdown has started.
	// This is not just an optimization:  This prevents a deadlock if
	// harvesting data during the shutdown fails and an attempt is made to
	// merge the data into the next harvest.
	shutdownStarted  chan struct{}
	shutdownComplete chan struct{}

	// Sends to these channels should not occur without a <-shutdownStarted
	// select option to prevent deadlock.
	dataChan           chan appData
	collectorErrorChan chan error
	connectChan        chan *appRun

	harvestTicker *time.Ticker

	// This mutex protects both `run` and `err`, both of which should only
	// be accessed using getState and setState.
	sync.RWMutex
	// run is non-nil when the app is successfully connected.  It is
	// immutable.
	run *appRun
	// err is non-nil if the application will never be connected again
	// (disconnect, license exception, shutdown).
	err error
}

// appRun contains information regarding a single connection session with the
// collector.  It is immutable after creation at application connect.
type appRun struct {
	*internal.ConnectReply

	// AttributeConfig is calculated on every connect since it depends on
	// the security policies.
	AttributeConfig *internal.AttributeConfig
}

func newAppRun(config Config, reply *internal.ConnectReply) *appRun {
	return &appRun{
		ConnectReply: reply,
		AttributeConfig: internal.CreateAttributeConfig(internal.AttributeConfigInput{
			Attributes:        convertAttributeDestinationConfig(config.Attributes),
			ErrorCollector:    convertAttributeDestinationConfig(config.ErrorCollector.Attributes),
			TransactionEvents: convertAttributeDestinationConfig(config.TransactionEvents.Attributes),
			TransactionTracer: convertAttributeDestinationConfig(config.TransactionTracer.Attributes),
		}, reply.SecurityPolicies.AttributesInclude.Enabled()),
	}
}

func isFatalHarvestError(e error) bool {
	return internal.IsDisconnect(e) ||
		internal.IsLicenseException(e) ||
		internal.IsRestartException(e)
}

func shouldSaveFailedHarvest(e error) bool {
	if e == internal.ErrPayloadTooLarge || e == internal.ErrUnsupportedMedia {
		return false
	}
	return true
}

func (app *app) doHarvest(h *internal.Harvest, harvestStart time.Time, run *appRun) {
	h.CreateFinalMetrics()
	h.Metrics = h.Metrics.ApplyRules(run.MetricRules)

	payloads := h.Payloads(app.config.DistributedTracer.Enabled)
	for _, p := range payloads {
		cmd := p.EndpointMethod()
		data, err := p.Data(run.RunID.String(), harvestStart)

		if nil == data && nil == err {
			continue
		}

		if nil == err {
			call := internal.RpmCmd{
				Collector: run.Collector,
				RunID:     run.RunID.String(),
				Name:      cmd,
				Data:      data,
			}

			// The reply from harvest calls is always unused.
			_, err = internal.CollectorRequest(call, app.rpmControls)
		}

		if nil == err {
			continue
		}

		if isFatalHarvestError(err) {
			select {
			case app.collectorErrorChan <- err:
			case <-app.shutdownStarted:
			}
			return
		}

		app.config.Logger.Warn("harvest failure", map[string]interface{}{
			"cmd":   cmd,
			"error": err.Error(),
		})

		if shouldSaveFailedHarvest(err) {
			app.Consume(run.RunID, p)
		}
	}
}

func connectAttempt(app *app) (*appRun, error) {
	reply, err := internal.ConnectAttempt(config{app.config}, app.config.SecurityPoliciesToken, app.rpmControls)
	if nil != err {
		return nil, err
	}
	return newAppRun(app.config, reply), nil
}

func (app *app) connectRoutine() {
	for {
		run, err := connectAttempt(app)
		if nil == err {
			select {
			case app.connectChan <- run:
			case <-app.shutdownStarted:
			}
			return
		}

		if internal.IsDisconnect(err) || internal.IsLicenseException(err) {
			select {
			case app.collectorErrorChan <- err:
			case <-app.shutdownStarted:
			}
			return
		}

		app.config.Logger.Warn("application connect failure", map[string]interface{}{
			"error": err.Error(),
		})

		time.Sleep(internal.ConnectBackoff)
	}
}

func debug(data internal.Harvestable, lg Logger) {
	now := time.Now()
	h := internal.NewHarvest(now)
	data.MergeIntoHarvest(h)
	ps := h.Payloads(false)
	for _, p := range ps {
		cmd := p.EndpointMethod()
		d, err := p.Data("agent run id", now)
		if nil == d && nil == err {
			continue
		}
		if nil != err {
			lg.Info("integration", map[string]interface{}{
				"cmd":   cmd,
				"error": err.Error(),
			})
			continue
		}
		lg.Info("integration", map[string]interface{}{
			"cmd":  cmd,
			"data": internal.JSONString(d),
		})
	}
}

func processConnectMessages(run *appRun, lg Logger) {
	for _, msg := range run.Messages {
		event := "collector message"
		cn := map[string]interface{}{"msg": msg.Message}

		switch strings.ToLower(msg.Level) {
		case "error":
			lg.Error(event, cn)
		case "warn":
			lg.Warn(event, cn)
		case "info":
			lg.Info(event, cn)
		case "debug", "verbose":
			lg.Debug(event, cn)
		}
	}
}

func (app *app) process() {
	// Both the harvest and the run are non-nil when the app is connected,
	// and nil otherwise.
	var h *internal.Harvest
	var run *appRun

	for {
		select {
		case <-app.harvestTicker.C:
			if nil != run {
				now := time.Now()
				go app.doHarvest(h, now, run)
				h = internal.NewHarvest(now)
			}
		case d := <-app.dataChan:
			if nil != run && run.RunID == d.id {
				d.data.MergeIntoHarvest(h)
			}
		case <-app.initiateShutdown:
			close(app.shutdownStarted)

			// Remove the run before merging any final data to
			// ensure a bounded number of receives from dataChan.
			app.setState(nil, errors.New("application shut down"))
			app.harvestTicker.Stop()

			if nil != run {
				for done := false; !done; {
					select {
					case d := <-app.dataChan:
						if run.RunID == d.id {
							d.data.MergeIntoHarvest(h)
						}
					default:
						done = true
					}
				}
				app.doHarvest(h, time.Now(), run)
			}

			close(app.shutdownComplete)
			return
		case err := <-app.collectorErrorChan:
			run = nil
			h = nil
			app.setState(nil, nil)

			switch {
			case internal.IsDisconnect(err):
				app.setState(nil, err)
				app.config.Logger.Error("application disconnected", map[string]interface{}{
					"app": app.config.AppName,
					"err": err.Error(),
				})
			case internal.IsLicenseException(err):
				app.setState(nil, err)
				app.config.Logger.Error("invalid license", map[string]interface{}{
					"app":     app.config.AppName,
					"license": app.config.License,
				})
			case internal.IsRestartException(err):
				app.config.Logger.Info("application restarted", map[string]interface{}{
					"app": app.config.AppName,
				})
				go app.connectRoutine()
			}
		case run = <-app.connectChan:
			h = internal.NewHarvest(time.Now())
			app.setState(run, nil)

			app.config.Logger.Info("application connected", map[string]interface{}{
				"app": app.config.AppName,
				"run": run.RunID.String(),
			})
			processConnectMessages(run, app.config.Logger)
		}
	}
}

func (app *app) Shutdown(timeout time.Duration) {
	if !app.config.Enabled {
		return
	}

	select {
	case app.initiateShutdown <- struct{}{}:
	default:
	}

	// Block until shutdown is done or timeout occurs.
	t := time.NewTimer(timeout)
	select {
	case <-app.shutdownComplete:
	case <-t.C:
	}
	t.Stop()

	app.config.Logger.Info("application shutdown", map[string]interface{}{
		"app": app.config.AppName,
	})
}

func convertAttributeDestinationConfig(c AttributeDestinationConfig) internal.AttributeDestinationConfig {
	return internal.AttributeDestinationConfig{
		Enabled: c.Enabled,
		Include: c.Include,
		Exclude: c.Exclude,
	}
}

func runSampler(app *app, period time.Duration) {
	previous := internal.GetSample(time.Now(), app.config.Logger)
	t := time.NewTicker(period)
	for {
		select {
		case now := <-t.C:
			current := internal.GetSample(now, app.config.Logger)
			run, _ := app.getState()
			app.Consume(run.RunID, internal.GetStats(internal.Samples{
				Previous: previous,
				Current:  current,
			}))
			previous = current
		case <-app.shutdownStarted:
			t.Stop()
			return
		}
	}
}

func (app *app) WaitForConnection(timeout time.Duration) error {
	if !app.config.Enabled {
		return nil
	}
	deadline := time.Now().Add(timeout)
	pollPeriod := 50 * time.Millisecond

	for {
		run, err := app.getState()
		if nil != err {
			return err
		}
		if run.RunID != "" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout out after %s", timeout.String())
		}
		time.Sleep(pollPeriod)
	}
}

func newApp(c Config) (Application, error) {
	c = copyConfigReferenceFields(c)
	if err := c.Validate(); nil != err {
		return nil, err
	}
	if nil == c.Logger {
		c.Logger = logger.ShimLogger{}
	}
	app := &app{
		config: c,

		placeholderRun: newAppRun(c, internal.ConnectReplyDefaults()),

		// This channel must be buffered since Shutdown makes a
		// non-blocking send attempt.
		initiateShutdown: make(chan struct{}, 1),

		shutdownStarted:    make(chan struct{}),
		shutdownComplete:   make(chan struct{}),
		connectChan:        make(chan *appRun, 1),
		collectorErrorChan: make(chan error, 1),
		dataChan:           make(chan appData, internal.AppDataChanSize),
		rpmControls: internal.RpmControls{
			License: c.License,
			Client: &http.Client{
				Transport: c.Transport,
				Timeout:   internal.CollectorTimeout,
			},
			Logger:       c.Logger,
			AgentVersion: Version,
		},
	}

	app.config.Logger.Info("application created", map[string]interface{}{
		"app":     app.config.AppName,
		"version": Version,
		"enabled": app.config.Enabled,
	})

	if !app.config.Enabled {
		return app, nil
	}

	app.harvestTicker = time.NewTicker(internal.HarvestPeriod)

	go app.process()
	go app.connectRoutine()

	if app.config.RuntimeSampler.Enabled {
		go runSampler(app, internal.RuntimeSamplerPeriod)
	}

	return app, nil
}

type expectApp interface {
	internal.Expect
	Application
}

func newTestApp(replyfn func(*internal.ConnectReply), cfg Config) (expectApp, error) {
	cfg.Enabled = false
	application, err := newApp(cfg)
	if nil != err {
		return nil, err
	}
	app := application.(*app)
	if nil != replyfn {
		replyfn(app.placeholderRun.ConnectReply)
		app.placeholderRun = newAppRun(cfg, app.placeholderRun.ConnectReply)
	}
	app.testHarvest = internal.NewHarvest(time.Now())

	return app, nil
}

func (app *app) getState() (*appRun, error) {
	app.RLock()
	defer app.RUnlock()

	run := app.run
	if nil == run {
		run = app.placeholderRun
	}
	return run, app.err
}

func (app *app) setState(run *appRun, err error) {
	app.Lock()
	defer app.Unlock()

	app.run = run
	app.err = err
}

func transportTypeFromRequest(r *http.Request) TransportType {
	if strings.HasPrefix(r.Proto, "HTTP") {
		if r.TLS != nil {
			return TransportHTTPS
		}
		return TransportHTTP
	}
	return TransportUnknown
}

// StartTransaction implements newrelic.Application's StartTransaction.
func (app *app) StartTransaction(name string, w http.ResponseWriter, r *http.Request) Transaction {
	run, _ := app.getState()
	txn := upgradeTxn(newTxn(txnInput{
		Config:     app.config,
		Reply:      run.ConnectReply,
		W:          w,
		Consumer:   app,
		attrConfig: run.AttributeConfig,
	}, r, name))

	if nil != r {
		if p := r.Header.Get(DistributedTracePayloadHeader); p != "" {
			txn.AcceptDistributedTracePayload(transportTypeFromRequest(r), p)
		}
	}
	return txn
}

var (
	errHighSecurityEnabled        = errors.New("high security enabled")
	errCustomEventsDisabled       = errors.New("custom events disabled")
	errCustomEventsRemoteDisabled = errors.New("custom events disabled by server")
)

// RecordCustomEvent implements newrelic.Application's RecordCustomEvent.
func (app *app) RecordCustomEvent(eventType string, params map[string]interface{}) error {
	if app.config.HighSecurity {
		return errHighSecurityEnabled
	}

	if !app.config.CustomInsightsEvents.Enabled {
		return errCustomEventsDisabled
	}

	event, e := internal.CreateCustomEvent(eventType, params, time.Now())
	if nil != e {
		return e
	}

	run, _ := app.getState()
	if !run.CollectCustomEvents {
		return errCustomEventsRemoteDisabled
	}

	if !run.SecurityPolicies.CustomEvents.Enabled() {
		return errSecurityPolicy
	}

	app.Consume(run.RunID, event)

	return nil
}

var (
	errMetricInf       = errors.New("invalid metric value: inf")
	errMetricNaN       = errors.New("invalid metric value: NaN")
	errMetricNameEmpty = errors.New("missing metric name")
)

// RecordCustomMetric implements newrelic.Application's RecordCustomMetric.
func (app *app) RecordCustomMetric(name string, value float64) error {
	if math.IsNaN(value) {
		return errMetricNaN
	}
	if math.IsInf(value, 0) {
		return errMetricInf
	}
	if "" == name {
		return errMetricNameEmpty
	}
	run, _ := app.getState()
	app.Consume(run.RunID, internal.CustomMetric{
		RawInputName: name,
		Value:        value,
	})
	return nil
}

func (app *app) Consume(id internal.AgentRunID, data internal.Harvestable) {
	if "" != debugLogging {
		debug(data, app.config.Logger)
	}

	if nil != app.testHarvest {
		data.MergeIntoHarvest(app.testHarvest)
		return
	}

	if "" == id {
		return
	}

	select {
	case app.dataChan <- appData{id, data}:
	case <-app.shutdownStarted:
	}
}

func (app *app) ExpectCustomEvents(t internal.Validator, want []internal.WantEvent) {
	internal.ExpectCustomEvents(internal.ExtendValidator(t, "custom events"), app.testHarvest.CustomEvents, want)
}

func (app *app) ExpectErrors(t internal.Validator, want []internal.WantError) {
	t = internal.ExtendValidator(t, "traced errors")
	internal.ExpectErrors(t, app.testHarvest.ErrorTraces, want)
}

func (app *app) ExpectErrorEvents(t internal.Validator, want []internal.WantEvent) {
	t = internal.ExtendValidator(t, "error events")
	internal.ExpectErrorEvents(t, app.testHarvest.ErrorEvents, want)
}

func (app *app) ExpectErrorEventsPresent(t internal.Validator, want []internal.WantEvent) {
	t = internal.ExtendValidator(t, "error events")
	internal.ExpectErrorEventsPresent(t, app.testHarvest.ErrorEvents, want)
}

func (app *app) ExpectErrorEventsAbsent(t internal.Validator, names []string) {
	t = internal.ExtendValidator(t, "error events")
	internal.ExpectErrorEventsAbsent(t, app.testHarvest.ErrorEvents, names)
}

func (app *app) ExpectSpanEvents(t internal.Validator, want []internal.WantEvent) {
	t = internal.ExtendValidator(t, "txn events")
	internal.ExpectSpanEvents(t, app.testHarvest.SpanEvents, want)
}

func (app *app) ExpectSpanEventsPresent(t internal.Validator, want []internal.WantEvent) {
	t = internal.ExtendValidator(t, "span events")
	internal.ExpectSpanEventsPresent(t, app.testHarvest.SpanEvents, want)
}

func (app *app) ExpectSpanEventsAbsent(t internal.Validator, names []string) {
	t = internal.ExtendValidator(t, "span events")
	internal.ExpectSpanEventsAbsent(t, app.testHarvest.SpanEvents, names)
}

func (app *app) ExpectSpanEventsCount(t internal.Validator, c int) {
	t = internal.ExtendValidator(t, "span events")
	internal.ExpectSpanEventsCount(t, app.testHarvest.SpanEvents, c)
}

func (app *app) ExpectTxnEvents(t internal.Validator, want []internal.WantEvent) {
	t = internal.ExtendValidator(t, "txn events")
	internal.ExpectTxnEvents(t, app.testHarvest.TxnEvents, want)
}

func (app *app) ExpectTxnEventsPresent(t internal.Validator, want []internal.WantEvent) {
	t = internal.ExtendValidator(t, "txn events")
	internal.ExpectTxnEventsPresent(t, app.testHarvest.TxnEvents, want)
}

func (app *app) ExpectTxnEventsAbsent(t internal.Validator, names []string) {
	t = internal.ExtendValidator(t, "txn events")
	internal.ExpectTxnEventsAbsent(t, app.testHarvest.TxnEvents, names)
}

func (app *app) ExpectMetrics(t internal.Validator, want []internal.WantMetric) {
	t = internal.ExtendValidator(t, "metrics")
	internal.ExpectMetrics(t, app.testHarvest.Metrics, want)
}

func (app *app) ExpectMetricsPresent(t internal.Validator, want []internal.WantMetric) {
	t = internal.ExtendValidator(t, "metrics")
	internal.ExpectMetricsPresent(t, app.testHarvest.Metrics, want)
}

func (app *app) ExpectTxnTraces(t internal.Validator, want []internal.WantTxnTrace) {
	t = internal.ExtendValidator(t, "txn traces")
	internal.ExpectTxnTraces(t, app.testHarvest.TxnTraces, want)
}

func (app *app) ExpectSlowQueries(t internal.Validator, want []internal.WantSlowQuery) {
	t = internal.ExtendValidator(t, "slow queries")
	internal.ExpectSlowQueries(t, app.testHarvest.SlowSQLs, want)
}
