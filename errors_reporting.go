package main

import (
	"net/http"
	"strings"

	"github.com/bugsnag/bugsnag-go"
	"github.com/honeybadger-io/honeybadger-go"
)

var (
	bugsnagEnabled     bool
	honeybadgerEnabled bool

	headersReplacer = strings.NewReplacer("-", "_")
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
}
