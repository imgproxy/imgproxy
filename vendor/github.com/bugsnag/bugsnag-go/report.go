package bugsnag

import (
	"github.com/bugsnag/bugsnag-go/sessions"
	uuid "github.com/gofrs/uuid"
)

type reportJSON struct {
	APIKey   string       `json:"apiKey"`
	Events   []eventJSON  `json:"events"`
	Notifier notifierJSON `json:"notifier"`
}

type notifierJSON struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

type eventJSON struct {
	App            *appJSON            `json:"app"`
	Context        string              `json:"context,omitempty"`
	Device         *deviceJSON         `json:"device,omitempty"`
	Request        *RequestJSON        `json:"request,omitempty"`
	Exceptions     []exceptionJSON     `json:"exceptions"`
	GroupingHash   string              `json:"groupingHash,omitempty"`
	Metadata       interface{}         `json:"metaData"`
	PayloadVersion string              `json:"payloadVersion"`
	Session        *sessionJSON        `json:"session,omitempty"`
	Severity       string              `json:"severity"`
	SeverityReason *severityReasonJSON `json:"severityReason,omitempty"`
	Unhandled      bool                `json:"unhandled"`
	User           *User               `json:"user,omitempty"`
}

type sessionJSON struct {
	StartedAt string               `json:"startedAt"`
	ID        uuid.UUID            `json:"id"`
	Events    sessions.EventCounts `json:"events"`
}

type appJSON struct {
	ReleaseStage string `json:"releaseStage"`
	Type         string `json:"type,omitempty"`
	Version      string `json:"version,omitempty"`
}

type exceptionJSON struct {
	ErrorClass string       `json:"errorClass"`
	Message    string       `json:"message"`
	Stacktrace []stackFrame `json:"stacktrace"`
}

type severityReasonJSON struct {
	Type SeverityReason `json:"type,omitempty"`
}

type deviceJSON struct {
	Hostname string `json:"hostname,omitempty"`
}

// RequestJSON is the request information that populates the Request tab in the dashboard.
type RequestJSON struct {
	ClientIP   string            `json:"clientIp,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	HTTPMethod string            `json:"httpMethod,omitempty"`
	URL        string            `json:"url,omitempty"`
	Referer    string            `json:"referer,omitempty"`
}
