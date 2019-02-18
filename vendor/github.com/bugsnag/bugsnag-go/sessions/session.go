package sessions

import (
	"time"

	uuid "github.com/gofrs/uuid"
)

// EventCounts register how many handled/unhandled events have happened for
// this session
type EventCounts struct {
	Handled   int `json:"handled"`
	Unhandled int `json:"unhandled"`
}

// Session represents a start time and a unique ID that identifies the session.
type Session struct {
	StartedAt   time.Time
	ID          uuid.UUID
	EventCounts *EventCounts
}

func newSession() *Session {
	sessionID, _ := uuid.NewV4()
	return &Session{
		StartedAt:   time.Now(),
		ID:          sessionID,
		EventCounts: &EventCounts{},
	}
}
