package bugsnag

import "github.com/bugsnag/bugsnag-go/v2"

// ReporterIface is the internal reporter type, exported for testing.
type ReporterIface = reporter

// NewWithNotifier builds a reporter around an already-configured Bugsnag notifier, for testing.
func NewWithNotifier(notifier *bugsnag.Notifier) *reporter {
	return &reporter{notifier: notifier}
}
