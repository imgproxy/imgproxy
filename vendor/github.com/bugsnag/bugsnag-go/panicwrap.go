// +build !appengine

package bugsnag

import (
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/bugsnag/bugsnag-go/sessions"
	"github.com/bugsnag/panicwrap"
)

// Forks and re-runs your program to add panic monitoring. This function does
// not return on one process, instead listening on stderr of the other process,
// which returns nil.
//
// Related: https://godoc.org/github.com/bugsnag/panicwrap#BasicMonitor
func defaultPanicHandler() {
	defer defaultNotifier.dontPanic()
	ctx := sessions.SendStartupSession(&sessionTrackingConfig)

	err := panicwrap.BasicMonitor(func(output string) {
		toNotify, err := errors.ParsePanic(output)

		if err != nil {
			defaultNotifier.Config.logf("bugsnag.handleUncaughtPanic: %v", err)
		}
		state := HandledState{SeverityReasonUnhandledPanic, SeverityError, true, ""}
		defaultNotifier.NotifySync(toNotify, true, state, ctx)

	})

	if err != nil {
		defaultNotifier.Config.logf("bugsnag.handleUncaughtPanic: %v", err)
	}
}
