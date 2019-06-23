package bugsnag

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/bugsnag/bugsnag-go/device"
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/bugsnag/bugsnag-go/sessions"

	// Fixes a bug with SHA-384 intermediate certs on some platforms.
	// - https://github.com/bugsnag/bugsnag-go/issues/9
	_ "crypto/sha512"
)

// VERSION defines the version of this Bugsnag notifier
const VERSION = "1.4.0"

var panicHandlerOnce sync.Once
var sessionTrackerOnce sync.Once
var middleware middlewareStack

// Config is the configuration for the default bugsnag notifier.
var Config Configuration
var sessionTrackingConfig sessions.SessionTrackingConfiguration

// DefaultSessionPublishInterval defines how often sessions should be sent to
// Bugsnag.
// Deprecated: Exposed for developer sanity in testing. Modify at own risk.
var DefaultSessionPublishInterval = 60 * time.Second
var defaultNotifier = Notifier{&Config, nil}
var sessionTracker sessions.SessionTracker

// Configure Bugsnag. The only required setting is the APIKey, which can be
// obtained by clicking on "Settings" in your Bugsnag dashboard. This function
// is also responsible for installing the global panic handler, so it should be
// called as early as possible in your initialization process.
func Configure(config Configuration) {
	Config.update(&config)
	updateSessionConfig()
	// Only do once in case the user overrides the default panichandler, and
	// configures multiple times.
	panicHandlerOnce.Do(Config.PanicHandler)
}

// StartSession creates new context from the context.Context instance with
// Bugsnag session data attached. Will start the session tracker if not already
// started
func StartSession(ctx context.Context) context.Context {
	sessionTrackerOnce.Do(startSessionTracking)
	return sessionTracker.StartSession(ctx)
}

// Notify sends an error.Error to Bugsnag along with the current stack trace.
// If at all possible, it is recommended to pass in a context.Context, e.g.
// from a http.Request or bugsnag.StartSession() as Bugsnag will be able to
// extract additional information in some cases. The rawData is used to send
// extra information along with the error. For example you can pass the current
// http.Request to Bugsnag to see information about it in the dashboard, or set
// the severity of the notification. For a detailed list of the information
// that can be extracted, see
// https://docs.bugsnag.com/platforms/go/reporting-handled-errors/
func Notify(err error, rawData ...interface{}) error {
	if e := checkForEmptyError(err); e != nil {
		return e
	}
	// Stripping one stackframe to not include this function in the stacktrace
	// for a manual notification.
	skipFrames := 1
	return defaultNotifier.Notify(errors.New(err, skipFrames), rawData...)
}

// AutoNotify logs a panic on a goroutine and then repanics.
// It should only be used in places that have existing panic handlers further
// up the stack.
// Although it's not strictly enforced, it's highly recommended to pass a
// context.Context object that has at one-point been returned from
// bugsnag.StartSession. Doing so ensures your stability score remains accurate,
// and future versions of Bugsnag may extract more useful information from this
// context.
// The rawData is used to send extra information along with any
// panics that are handled this way.
// Usage:
//  go func() {
//      ctx := bugsnag.StartSession(context.Background())
//		defer bugsnag.AutoNotify(ctx)
//      // (possibly crashy code)
//  }()
// See also: bugsnag.Recover()
func AutoNotify(rawData ...interface{}) {
	if err := recover(); err != nil {
		severity := defaultNotifier.getDefaultSeverity(rawData, SeverityError)
		state := HandledState{SeverityReasonHandledPanic, severity, true, ""}
		rawData = append([]interface{}{state}, rawData...)
		// We strip the following stackframes as they don't add much info
		// - runtime/$arch - e.g. runtime/asm_amd64.s#call32
		// - runtime/panic.go#gopanic
		// Panics have their own stacktrace, so no stripping of the current stack
		skipFrames := 2
		defaultNotifier.NotifySync(errors.New(err, skipFrames), true, rawData...)
		sessionTracker.FlushSessions()
		panic(err)
	}
}

// Recover logs a panic on a goroutine and then recovers.
// Although it's not strictly enforced, it's highly recommended to pass a
// context.Context object that has at one-point been returned from
// bugsnag.StartSession. Doing so ensures your stability score remains accurate,
// and future versions of Bugsnag may extract more useful information from this
// context.
// The rawData is used to send extra information along with
// any panics that are handled this way
// Usage:
// go func() {
//     ctx := bugsnag.StartSession(context.Background())
//     defer bugsnag.Recover(ctx)
//     // (possibly crashy code)
// }()
// If you wish that any panics caught by the call to Recover shall affect your
// stability score (it does not by default):
// go func() {
//     ctx := bugsnag.StartSession(context.Background())
//     defer bugsnag.Recover(ctx, bugsnag.HandledState{Unhandled: true})
//     // (possibly crashy code)
// }()
// See also: bugsnag.AutoNotify()
func Recover(rawData ...interface{}) {
	if err := recover(); err != nil {
		severity := defaultNotifier.getDefaultSeverity(rawData, SeverityWarning)
		state := HandledState{SeverityReasonHandledPanic, severity, false, ""}
		rawData = append([]interface{}{state}, rawData...)
		// We strip the following stackframes as they don't add much info
		// - runtime/$arch - e.g. runtime/asm_amd64.s#call32
		// - runtime/panic.go#gopanic
		// Panics have their own stacktrace, so no stripping of the current stack
		skipFrames := 2
		defaultNotifier.Notify(errors.New(err, skipFrames), rawData...)
	}
}

// OnBeforeNotify adds a callback to be run before a notification is sent to
// Bugsnag.  It can be used to modify the event or its MetaData. Changes made
// to the configuration are local to notifying about this event. To prevent the
// event from being sent to Bugsnag return an error, this error will be
// returned from bugsnag.Notify() and the event will not be sent.
func OnBeforeNotify(callback func(event *Event, config *Configuration) error) {
	middleware.OnBeforeNotify(callback)
}

// Handler creates an http Handler that notifies Bugsnag any panics that
// happen. It then repanics so that the default http Server panic handler can
// handle the panic too. The rawData is used to send extra information along
// with any panics that are handled this way.
func Handler(h http.Handler, rawData ...interface{}) http.Handler {
	notifier := New(rawData...)
	if h == nil {
		h = http.DefaultServeMux
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r

		// Record a session if auto notify session is enabled
		ctx := r.Context()
		if Config.IsAutoCaptureSessions() {
			ctx = StartSession(ctx)
		}
		ctx = AttachRequestData(ctx, request)
		request = r.WithContext(ctx)
		defer notifier.AutoNotify(ctx, request)
		h.ServeHTTP(w, request)
	})
}

// HandlerFunc creates an http HandlerFunc that notifies Bugsnag about any
// panics that happen. It then repanics so that the default http Server panic
// handler can handle the panic too. The rawData is used to send extra
// information along with any panics that are handled this way. If you have
// already wrapped your http server using bugsnag.Handler() you don't also need
// to wrap each HandlerFunc.
func HandlerFunc(h http.HandlerFunc, rawData ...interface{}) http.HandlerFunc {
	notifier := New(rawData...)

	return func(w http.ResponseWriter, r *http.Request) {
		request := r
		// Record a session if auto notify session is enabled
		ctx := request.Context()
		if notifier.Config.IsAutoCaptureSessions() {
			ctx = StartSession(ctx)
		}
		ctx = AttachRequestData(ctx, request)
		request = request.WithContext(ctx)
		defer notifier.AutoNotify(ctx)
		h(w, request)
	}
}

// checkForEmptyError checks if the given error (to be reported to Bugsnag) is
// nil. If it is, then log an error message and return another error wrapping
// this error message.
func checkForEmptyError(err error) error {
	if err != nil {
		return nil
	}
	msg := "attempted to notify Bugsnag without supplying an error. Bugsnag not notified"
	Config.Logger.Printf("ERROR: " + msg)
	return fmt.Errorf(msg)
}

func init() {
	// Set up builtin middlewarez
	OnBeforeNotify(httpRequestMiddleware)

	// Default configuration
	sourceRoot := ""
	if gopath := os.Getenv("GOPATH"); len(gopath) > 0 {
		sourceRoot = filepath.Join(gopath, "src") + "/"
	} else {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src") + "/"
	}
	Config.update(&Configuration{
		APIKey: "",
		Endpoints: Endpoints{
			Notify:   "https://notify.bugsnag.com",
			Sessions: "https://sessions.bugsnag.com",
		},
		Hostname:            device.GetHostname(),
		AppType:             "",
		AppVersion:          "",
		AutoCaptureSessions: true,
		ReleaseStage:        "",
		ParamsFilters:       []string{"password", "secret", "authorization", "cookie"},
		SourceRoot:          sourceRoot,
		// * for app-engine
		ProjectPackages:     []string{"main*"},
		NotifyReleaseStages: nil,
		Logger:              log.New(os.Stdout, log.Prefix(), log.Flags()),
		PanicHandler:        defaultPanicHandler,
		Transport:           http.DefaultTransport,

		flushSessionsOnRepanic: true,
	})
	updateSessionConfig()
}

func startSessionTracking() {
	if sessionTracker == nil {
		updateSessionConfig()
		sessionTracker = sessions.NewSessionTracker(&sessionTrackingConfig)
	}
}

func updateSessionConfig() {
	sessionTrackingConfig.Update(&sessions.SessionTrackingConfiguration{
		APIKey:              Config.APIKey,
		AutoCaptureSessions: Config.AutoCaptureSessions,
		Endpoint:            Config.Endpoints.Sessions,
		Version:             VERSION,
		PublishInterval:     DefaultSessionPublishInterval,
		Transport:           Config.Transport,
		ReleaseStage:        Config.ReleaseStage,
		Hostname:            Config.Hostname,
		AppType:             Config.AppType,
		AppVersion:          Config.AppVersion,
		NotifyReleaseStages: Config.NotifyReleaseStages,
		Logger:              Config.Logger,
	})
}
