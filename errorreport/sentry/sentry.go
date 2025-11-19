package sentry

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/imgproxy/imgproxy/v3/errctx"
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

func (r *reporter) Report(err errctx.Error, req *http.Request, meta map[string]any) {
	hub := r.hub.Clone()
	hub.Scope().SetRequest(req)
	hub.Scope().SetLevel(sentry.LevelError)

	if meta != nil {
		hub.Scope().SetContext("Processing context", meta)
	}

	// imgproxy may wrap errors using errctx.WrappedError to add context, so Sentry
	// would report the error type as *errctx.WrappedError.
	//
	// To avoid this, we create the event manually from the original error
	// and set the correct error type.
	if event := hub.Client().EventFromException(err, sentry.LevelError); event != nil {
		// Sentry reports errors in the reverse order: the last one is the outermost error.
		// So we need to set the type on the last exception.
		event.Exception[len(event.Exception)-1].Type = errctx.ErrorType(err)
		hub.CaptureEvent(event)
	}
}

func (r *reporter) Close() {
	r.hub.Flush(flushTimeout)
}
