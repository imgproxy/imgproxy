package sessions

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// SessionTrackingConfiguration defines the configuration options relevant for session tracking.
// These are likely a subset of the global bugsnag.Configuration. Users should
// not modify this struct directly but rather call
// `bugsnag.Configure(bugsnag.Configuration)` which will update this configuration in return.
type SessionTrackingConfiguration struct {
	// PublishInterval defines how often the sessions are sent off to the session server.
	PublishInterval time.Duration

	// AutoCaptureSessions can be set to false to disable automatic session
	// tracking. If you want control over what is deemed a session, you can
	// switch off automatic session tracking with this configuration, and call
	// bugsnag.StartSession() when appropriate for your application. See the
	// official docs for instructions and examples of associating handled
	// errors with sessions and ensuring error rate accuracy on the Bugsnag
	// dashboard. This will default to true, but is stored as an interface to enable
	// us to detect when this option has not been set.
	AutoCaptureSessions interface{}

	// APIKey defines the API key for the Bugsnag project. Same value as for reporting errors.
	APIKey string
	// Endpoint is the URI of the session server to receive session payloads.
	Endpoint string
	// Version defines the current version of the notifier.
	Version string

	// ReleaseStage defines the release stage, e.g. "production" or "staging",
	// that this session occurred in. The release stage, in combination with
	// the app version make up the release that Bugsnag tracks.
	ReleaseStage string
	// Hostname defines the host of the server this application is running on.
	Hostname string
	// AppType defines the type of the application.
	AppType string
	// AppVersion defines the version of the application.
	AppVersion string
	// Transport defines the http.RoundTripper to be used for managing HTTP requests.
	Transport http.RoundTripper

	// The release stages to notify about sessions in. If you set this then
	// bugsnag-go will only send sessions to Bugsnag if the release stage
	// is listed here.
	NotifyReleaseStages []string

	// Logger is the logger that Bugsnag should log to. Uses the same defaults
	// as go's builtin logging package. This logger gets invoked when any error
	// occurs inside the library itself.
	Logger interface {
		Printf(format string, v ...interface{})
	}

	mutex sync.Mutex
}

// Update modifies the values inside the receiver to match the non-default properties of the given config.
// Existing properties will not be cleared when given empty fields.
func (c *SessionTrackingConfiguration) Update(config *SessionTrackingConfiguration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if config.PublishInterval != 0 {
		c.PublishInterval = config.PublishInterval
	}
	if config.APIKey != "" {
		c.APIKey = config.APIKey
	}
	if config.Endpoint != "" {
		c.Endpoint = config.Endpoint
	}
	if config.Version != "" {
		c.Version = config.Version
	}
	if config.ReleaseStage != "" {
		c.ReleaseStage = config.ReleaseStage
	}
	if config.Hostname != "" {
		c.Hostname = config.Hostname
	}
	if config.AppType != "" {
		c.AppType = config.AppType
	}
	if config.AppVersion != "" {
		c.AppVersion = config.AppVersion
	}
	if config.Transport != nil {
		c.Transport = config.Transport
	}
	if config.Logger != nil {
		c.Logger = config.Logger
	}
	if config.NotifyReleaseStages != nil {
		c.NotifyReleaseStages = config.NotifyReleaseStages
	}
	if config.AutoCaptureSessions != nil {
		c.AutoCaptureSessions = config.AutoCaptureSessions
	}
}

func (c *SessionTrackingConfiguration) logf(fmt string, args ...interface{}) {
	if c != nil && c.Logger != nil {
		c.Logger.Printf(fmt, args...)
	} else {
		log.Printf(fmt, args...)
	}
}

// IsAutoCaptureSessions identifies whether or not the notifier should
// automatically capture sessions as requests come in. It's a convenience
// wrapper that allows automatic session capturing to be enabled by default.
func (c *SessionTrackingConfiguration) IsAutoCaptureSessions() bool {
	if c.AutoCaptureSessions == nil {
		return true // enabled by default
	}
	if val, ok := c.AutoCaptureSessions.(bool); ok {
		return val
	}
	// It has been configured to *something* (although not a valid value)
	// assume the user wanted to disable this option.
	return false
}
