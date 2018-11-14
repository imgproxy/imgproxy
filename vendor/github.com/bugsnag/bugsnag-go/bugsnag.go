package bugsnag

import (
	"github.com/bugsnag/bugsnag-go/errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	// Fixes a bug with SHA-384 intermediate certs on some platforms.
	// - https://github.com/bugsnag/bugsnag-go/issues/9
	_ "crypto/sha512"
)

// The current version of bugsnag-go.
const VERSION = "1.3.2"

var once sync.Once
var middleware middlewareStack

// The configuration for the default bugsnag notifier.
var Config Configuration

var defaultNotifier = Notifier{&Config, nil}

// Configure Bugsnag. The only required setting is the APIKey, which can be
// obtained by clicking on "Settings" in your Bugsnag dashboard. This function
// is also responsible for installing the global panic handler, so it should be
// called as early as possible in your initialization process.
func Configure(config Configuration) {
	Config.update(&config)
	once.Do(Config.PanicHandler)
}

// Notify sends an error to Bugsnag along with the current stack trace. The
// rawData is used to send extra information along with the error. For example
// you can pass the current http.Request to Bugsnag to see information about it
// in the dashboard, or set the severity of the notification.
func Notify(err error, rawData ...interface{}) error {
	return defaultNotifier.Notify(errors.New(err, 1), rawData...)
}

// AutoNotify logs a panic on a goroutine and then repanics.
// It should only be used in places that have existing panic handlers further
// up the stack. The rawData is used to send extra information along with any
// panics that are handled this way.
// Usage:
//  go func() {
//		defer bugsnag.AutoNotify()
//      // (possibly crashy code)
//  }()
// See also: bugsnag.Recover()
func AutoNotify(rawData ...interface{}) {
	if err := recover(); err != nil {
		severity := defaultNotifier.getDefaultSeverity(rawData, SeverityError)
		state := HandledState{SeverityReasonHandledPanic, severity, true, ""}
		rawData = append([]interface{}{state}, rawData...)
		defaultNotifier.NotifySync(errors.New(err, 2), true, rawData...)
		panic(err)
	}
}

// Recover logs a panic on a goroutine and then recovers.
// The rawData is used to send extra information along with
// any panics that are handled this way
// Usage: defer bugsnag.Recover()
func Recover(rawData ...interface{}) {
	if err := recover(); err != nil {
		severity := defaultNotifier.getDefaultSeverity(rawData, SeverityWarning)
		state := HandledState{SeverityReasonHandledPanic, severity, false, ""}
		rawData = append([]interface{}{state}, rawData...)
		defaultNotifier.Notify(errors.New(err, 2), rawData...)
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
		defer notifier.AutoNotify(r)
		h.ServeHTTP(w, r)
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
		defer notifier.AutoNotify(r)
		h(w, r)
	}
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
		APIKey:        "",
		Endpoint:      "https://notify.bugsnag.com/",
		Hostname:      "",
		AppType:       "",
		AppVersion:    "",
		ReleaseStage:  "",
		ParamsFilters: []string{"password", "secret"},
		SourceRoot:    sourceRoot,
		// * for app-engine
		ProjectPackages:     []string{"main*"},
		NotifyReleaseStages: nil,
		Logger:              log.New(os.Stdout, log.Prefix(), log.Flags()),
		PanicHandler:        defaultPanicHandler,
		Transport:           http.DefaultTransport,
	})

	hostname, err := os.Hostname()
	if err == nil {
		Config.Hostname = hostname
	}
}
