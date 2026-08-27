package airbrake

import "github.com/airbrake/gobrake/v5"

// ReporterIface is the internal reporter type, exported for testing.
type ReporterIface = reporter

// NewWithNotifier builds a reporter around an already-configured Airbrake notifier, for testing.
func NewWithNotifier(notifier *gobrake.Notifier) *reporter {
	return &reporter{notifier: notifier}
}
