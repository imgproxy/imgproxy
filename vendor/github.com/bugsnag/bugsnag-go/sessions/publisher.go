package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bugsnag/bugsnag-go/headers"
)

// sessionPayloadVersion defines the current version of the payload that's
// being sent to the session server.
const sessionPayloadVersion = "1.0"

type sessionPublisher interface {
	publish(sessions []*Session) error
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type publisher struct {
	config *SessionTrackingConfiguration
	client httpClient
}

// publish builds a payload from the given sessions and publishes them to the
// session server. Returns any errors that happened as part of publishing.
func (p *publisher) publish(sessions []*Session) error {
	if p.config.Endpoint == "" {
		// Session tracking is disabled, likely because the notify endpoint was
		// changed without changing the sessions endpoint
		// We've already logged a warning in this case, so no need to spam the
		// log every minute
		return nil
	}
	if apiKey := p.config.APIKey; len(apiKey) != 32 {
		return fmt.Errorf("bugsnag/sessions/publisher.publish invalid API key: '%s'", apiKey)
	}
	nrs, rs := p.config.NotifyReleaseStages, p.config.ReleaseStage
	if rs != "" && (nrs != nil && !contains(nrs, rs)) {
		// Always send sessions if the release stage is not set, but don't send any
		// sessions when notify release stages don't match the current release stage
		return nil
	}
	if len(sessions) == 0 {
		return fmt.Errorf("bugsnag/sessions/publisher.publish requested publication of 0")
	}
	p.config.mutex.Lock()
	defer p.config.mutex.Unlock()
	payload := makeSessionPayload(sessions, p.config)
	buf, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("bugsnag/sessions/publisher.publish unable to marshal json: %v", err)
	}
	req, err := http.NewRequest("POST", p.config.Endpoint, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("bugsnag/sessions/publisher.publish unable to create request: %v", err)
	}
	for k, v := range headers.PrefixedHeaders(p.config.APIKey, sessionPayloadVersion) {
		req.Header.Add(k, v)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("bugsnag/sessions/publisher.publish unable to deliver session: %v", err)
	}
	defer func(res *http.Response) {
		if err := res.Body.Close(); err != nil {
			p.config.logf("%v", err)
		}
	}(res)
	if res.StatusCode != 202 {
		return fmt.Errorf("bugsnag/session.publish expected 202 response status, got HTTP %s", res.Status)
	}
	return nil
}

func contains(coll []string, e string) bool {
	for _, s := range coll {
		if s == e {
			return true
		}
	}
	return false
}
