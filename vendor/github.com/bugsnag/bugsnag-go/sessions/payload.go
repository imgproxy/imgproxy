package sessions

import (
	"runtime"
	"time"

	"github.com/bugsnag/bugsnag-go/device"
)

// notifierPayload defines the .notifier subobject of the payload
type notifierPayload struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

// appPayload defines the .app subobject of the payload
type appPayload struct {
	Type         string `json:"type,omitempty"`
	ReleaseStage string `json:"releaseStage,omitempty"`
	Version      string `json:"version,omitempty"`
}

// devicePayload defines the .device subobject of the payload
type devicePayload struct {
	OsName   string `json:"osName,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// sessionCountsPayload defines the .sessionCounts subobject of the payload
type sessionCountsPayload struct {
	StartedAt       string `json:"startedAt"`
	SessionsStarted int    `json:"sessionsStarted"`
}

// sessionPayload defines the top level payload object
type sessionPayload struct {
	Notifier      *notifierPayload       `json:"notifier"`
	App           *appPayload            `json:"app"`
	Device        *devicePayload         `json:"device"`
	SessionCounts []sessionCountsPayload `json:"sessionCounts"`
}

// makeSessionPayload creates a sessionPayload based off of the given sessions and config
func makeSessionPayload(sessions []*Session, config *SessionTrackingConfiguration) *sessionPayload {
	releaseStage := config.ReleaseStage
	if releaseStage == "" {
		releaseStage = "production"
	}
	hostname := config.Hostname
	if hostname == "" {
		hostname = device.GetHostname()
	}

	return &sessionPayload{
		Notifier: &notifierPayload{
			Name:    "Bugsnag Go",
			URL:     "https://github.com/bugsnag/bugsnag-go",
			Version: config.Version,
		},
		App: &appPayload{
			Type:         config.AppType,
			Version:      config.AppVersion,
			ReleaseStage: releaseStage,
		},
		Device: &devicePayload{
			OsName:   runtime.GOOS,
			Hostname: hostname,
		},
		SessionCounts: []sessionCountsPayload{
			{
				//This timestamp assumes that we're sending these off once a minute
				StartedAt:       sessions[0].StartedAt.UTC().Format(time.RFC3339),
				SessionsStarted: len(sessions),
			},
		},
	}
}
