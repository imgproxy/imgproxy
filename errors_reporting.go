package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/bugsnag/bugsnag-go"
	"github.com/getsentry/sentry-go"
	"github.com/honeybadger-io/honeybadger-go"
)

var (
	bugsnagEnabled     bool
	honeybadgerEnabled bool
	sentryEnabled      bool

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
}
