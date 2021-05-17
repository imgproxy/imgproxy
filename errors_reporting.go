package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/airbrake/gobrake/v5"
	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/getsentry/sentry-go"
	"github.com/honeybadger-io/honeybadger-go"
)

var (
	bugsnagEnabled     bool
	honeybadgerEnabled bool
	sentryEnabled      bool
	airbrakeEnabled    bool
	airbrake           *gobrake.Notifier

	headersReplacer = strings.NewReplacer("-", "_")
	sentryTimeout   = 5 * time.Second
)

func initErrorsReporting() {
	if len(conf.BugsnagKey) > 0 {
		bugsnag.Configure(bugsnag.Configuration{
			APIKey:       conf.BugsnagKey,
			ReleaseStage: conf.BugsnagStage,
		})
		bugsnagEnabled = true
	}

	if len(conf.HoneybadgerKey) > 0 {
		honeybadger.Configure(honeybadger.Configuration{
			APIKey: conf.HoneybadgerKey,
			Env:    conf.HoneybadgerEnv,
		})
		honeybadgerEnabled = true
	}

	if len(conf.SentryDSN) > 0 {
		sentry.Init(sentry.ClientOptions{
			Dsn:         conf.SentryDSN,
			Release:     conf.SentryRelease,
			Environment: conf.SentryEnvironment,
		})

		sentryEnabled = true
	}

	if len(conf.AirbrakeProjecKey) > 0 {
		airbrake = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
			ProjectId:   int64(conf.AirbrakeProjecID),
			ProjectKey:  conf.AirbrakeProjecKey,
			Environment: conf.AirbrakeEnv,
		})

		airbrakeEnabled = true
	}
}

func closeErrorsReporting() {
	if airbrake != nil {
		airbrake.Close()
	}
}

func reportError(err error, req *http.Request) {
	if bugsnagEnabled {
		bugsnag.Notify(err, req)
	}

	if honeybadgerEnabled {
		headers := make(honeybadger.CGIData)

		for k, v := range req.Header {
			key := "HTTP_" + headersReplacer.Replace(strings.ToUpper(k))
			headers[key] = v[0]
		}

		honeybadger.Notify(err, req.URL, headers)
	}

	if sentryEnabled {
		hub := sentry.CurrentHub().Clone()
		hub.Scope().SetRequest(req)
		hub.Scope().SetLevel(sentry.LevelError)
		eventID := hub.CaptureException(err)
		if eventID != nil {
			hub.Flush(sentryTimeout)
		}
	}

	if airbrakeEnabled {
		airbrake.Notify(err, req)
	}
}
