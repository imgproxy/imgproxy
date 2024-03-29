package sentry

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/imgproxy/imgproxy/v3/config"
)

var (
	enabled bool

	timeout = 5 * time.Second
)

func Init() {
	if len(config.SentryDSN) > 0 {
		sentry.Init(sentry.ClientOptions{
			Dsn:         config.SentryDSN,
			Release:     config.SentryRelease,
			Environment: config.SentryEnvironment,
		})

		enabled = true
	}
}

func Report(err error, req *http.Request, meta map[string]any) {
	if !enabled {
		return
	}

	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetRequest(req)
	hub.Scope().SetLevel(sentry.LevelError)

	if meta != nil {
		hub.Scope().SetContext("Processing context", meta)
	}

	eventID := hub.CaptureException(err)
	if eventID != nil {
		hub.Flush(timeout)
	}
}
