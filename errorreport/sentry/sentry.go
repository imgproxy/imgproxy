package sentry

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/version"
)

const (
	// flushTimeout is the maximum time to wait for Sentry to send events
	flushTimeout = 5 * time.Second
)

var (
	IMGPROXY_SENTRY_DSN         = env.Describe("IMGPROXY_SENTRY_DSN", "string")
	IMGPROXY_SENTRY_RELEASE     = env.Describe("IMGPROXY_SENTRY_RELEASE", "string")
	IMGPROXY_SENTRY_ENVIRONMENT = env.Describe("IMGPROXY_SENTRY_ENVIRONMENT", "string")
)

// reporter is a Sentry error reporter
type reporter struct{}

// new creates and configures a new Sentry reporter
func New() (*reporter, error) {
	dsn := ""
	environment := "production"
	release := fmt.Sprintf("imgproxy@%s", version.Version)

	err := errors.Join(
		env.String(&dsn, IMGPROXY_SENTRY_DSN),
		env.String(&release, IMGPROXY_SENTRY_RELEASE),
		env.String(&environment, IMGPROXY_SENTRY_ENVIRONMENT),
	)
	if err != nil {
		return nil, err
	}

	if len(dsn) == 0 {
		return nil, nil
	}

	sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Release:     release,
		Environment: environment,
	})

	return &reporter{}, nil
}

func (r *reporter) Report(err error, req *http.Request, meta map[string]any) {
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
			hub.Flush(flushTimeout)
		}
	}
}

func (r *reporter) Close() {
	sentry.Flush(flushTimeout)
}
