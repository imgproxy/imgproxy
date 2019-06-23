package sessions

import (
	"context"
	"net/http"
	"os"

	"github.com/bugsnag/panicwrap"
)

// SendStartupSession is called by Bugsnag on startup, which will send a
// session to Bugsnag and return a context to represent the session of the main
// goroutine. This is the session associated with any fatal panics that are
// caught by panicwrap.
func SendStartupSession(config *SessionTrackingConfiguration) context.Context {
	ctx := context.Background()
	session := newSession()
	if !config.IsAutoCaptureSessions() || isApplicationProcess() {
		return ctx
	}
	publisher := &publisher{
		config: config,
		client: &http.Client{Transport: config.Transport},
	}
	go publisher.publish([]*Session{session})
	return context.WithValue(ctx, contextSessionKey, session)
}

// Checks to see if this is the application process, as opposed to the process
// that monitors for panics
func isApplicationProcess() bool {
	// Application process is run first, and this will only have been set when
	// the monitoring process runs
	return "" == os.Getenv(panicwrap.DEFAULT_COOKIE_KEY)
}
