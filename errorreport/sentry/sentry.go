package sentry

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	// flushTimeout is the maximum time to wait for Sentry to send events
	flushTimeout = 5 * time.Second
)

// reporter is a Sentry error reporter
type reporter struct {
	hub *sentry.Hub
}

// New creates and configures a new Sentry reporter
func New(config *Config) (*reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(config.DSN) == 0 {
		return nil, nil
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         config.DSN,
		Release:     config.Release,
		Environment: config.Environment,
	})
	if err != nil {
		return nil, err
	}

	hub := sentry.NewHub(client, sentry.NewScope())

	return &reporter{hub: hub}, nil
}

func (r *reporter) Report(err error, req *http.Request, meta map[string]any) {
	hub := r.hub.Clone()
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

		hub.CaptureEvent(event)
	}
}

func (r *reporter) Close() {
	r.hub.Flush(flushTimeout)
}
