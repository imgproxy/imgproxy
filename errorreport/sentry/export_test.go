package sentry

import "github.com/getsentry/sentry-go"

// ReporterIface is the internal reporter type, exported for testing.
type ReporterIface = reporter

// NewWithClient builds a reporter around an already-configured Sentry client, for testing.
func NewWithClient(client *sentry.Client) *reporter {
	return &reporter{hub: sentry.NewHub(client, sentry.NewScope())}
}
