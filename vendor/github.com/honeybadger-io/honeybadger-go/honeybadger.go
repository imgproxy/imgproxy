package honeybadger

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// VERSION defines the version of the honeybadger package.
const VERSION = "0.4.0"

var (
	// client is a pre-defined "global" client.
	DefaultClient = New(Configuration{})

	// Config is a pointer to the global client's Config.
	Config = DefaultClient.Config

	// Notices is the feature for sending error reports.
	Notices = Feature{"notices"}
)

// Feature references a resource provided by the API service. Its Endpoint maps
// to the collection endpoint of the /v1 API.
type Feature struct {
	Endpoint string
}

// CGIData stores variables from the server/request environment indexed by key.
// Header keys should be converted to upercase, all non-alphanumeric characters
// replaced with underscores, and prefixed with HTTP_. For example, the header
// "Content-Type" would become "HTTP_CONTENT_TYPE".
type CGIData hash

// Params stores the form or url values from an HTTP request.
type Params url.Values

// Tags represents tags of the error which is classified errors in Honeybadger.
type Tags []string

// hash is used internally to construct JSON payloads.
type hash map[string]interface{}

func (h *hash) toJSON() []byte {
	out, err := json.Marshal(h)
	if err == nil {
		return out
	}
	panic(err)
}

// Configure updates configuration of the global client.
func Configure(c Configuration) {
	DefaultClient.Configure(c)
}

// SetContext merges c Context into the Context of the global client.
func SetContext(c Context) {
	DefaultClient.SetContext(c)
}

// Notify reports the error err to the Honeybadger service.
//
// The first argument err may be an error, a string, or any other type in which
// case its formatted value will be used.
//
// It returns a string UUID which can be used to reference the error from the
// Honeybadger service, and an error as a second argument.
func Notify(err interface{}, extra ...interface{}) (string, error) {
	return DefaultClient.Notify(newError(err, 2), extra...)
}

// Monitor is used to automatically notify Honeybadger service of panics which
// happen inside the current function. In order to monitor for panics, defer a
// call to Monitor. For example:
// 	func main {
// 		defer honeybadger.Monitor()
// 		// Do risky stuff...
// 	}
// The Monitor function re-panics after the notification has been sent, so it's
// still up to the user to recover from panics if desired.
func Monitor() {
	if err := recover(); err != nil {
		DefaultClient.Notify(newError(err, 2))
		DefaultClient.Flush()
		panic(err)
	}
}

// Flush blocks until all data (normally sent in the background) has been sent
// to the Honeybadger service.
func Flush() {
	DefaultClient.Flush()
}

// Handler returns an http.Handler function which automatically reports panics
// to Honeybadger and then re-panics.
func Handler(h http.Handler) http.Handler {
	return DefaultClient.Handler(h)
}

// BeforeNotify adds a callback function which is run before a notice is
// reported to Honeybadger. If any function returns an error the notification
// will be skipped, otherwise it will be sent.
func BeforeNotify(handler func(notice *Notice) error) {
	DefaultClient.BeforeNotify(handler)
}
