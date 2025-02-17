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

	// imgproxy wraps almost all errors into *ierrors.Error, so Sentry will show
	// the same error type for all errors. We need to fix it.
	//
	// Instead of using hub.CaptureException(err), we need to create an event
	// manually and replace `*ierrors.Error` with the wrapped error type
	// (which is the previous exception type in the exception chain).
	if event := hub.Client().EventFromException(err, sentry.LevelError); event != nil {
		for i := 1; i < len(event.Exception); i++ {
			if event.Exception[i].Type == "*ierrors.Error" {
				event.Exception[i].Type = event.Exception[i-1].Type
			}
		}

		eventID := hub.CaptureEvent(event)
		if eventID != nil {
			hub.Flush(timeout)
		}
	}
}
