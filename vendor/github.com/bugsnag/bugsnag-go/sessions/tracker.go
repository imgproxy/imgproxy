package sessions

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	//contextSessionKey is a unique key for accessing and setting Bugsnag
	//session data on a context.Context object
	contextSessionKey ctxKey = 1
)

// ctxKey is a type alias that ensures uniqueness as a context.Context key
type ctxKey int

// SessionTracker exposes a method for starting sessions that are used for
// gauging your application's health
type SessionTracker interface {
	StartSession(context.Context) context.Context
	FlushSessions()
}

type sessionTracker struct {
	sessionChannel chan *Session
	sessions       []*Session
	config         *SessionTrackingConfiguration
	publisher      sessionPublisher
	sessionsMutex  sync.Mutex
}

// NewSessionTracker creates a new SessionTracker based on the provided config,
func NewSessionTracker(config *SessionTrackingConfiguration) SessionTracker {
	publisher := publisher{
		config: config,
		client: &http.Client{Transport: config.Transport},
	}
	st := sessionTracker{
		sessionChannel: make(chan *Session, 1),
		sessions:       []*Session{},
		config:         config,
		publisher:      &publisher,
	}
	go st.processSessions()
	return &st
}

// IncrementEventCountAndGetSession extracts a Bugsnag session from the given
// context and increments the event count of unhandled or handled events and
// returns the session
func IncrementEventCountAndGetSession(ctx context.Context, unhandled bool) *Session {
	if s := ctx.Value(contextSessionKey); s != nil {
		if session, ok := s.(*Session); ok && !session.StartedAt.IsZero() {
			// It is not just getting back a default value
			ec := session.EventCounts
			if unhandled {
				ec.Unhandled++
			} else {
				ec.Handled++
			}
			return session
		}
	}
	return nil
}

func (s *sessionTracker) StartSession(ctx context.Context) context.Context {
	session := newSession()
	s.sessionChannel <- session
	return context.WithValue(ctx, contextSessionKey, session)
}

func (s *sessionTracker) interval() time.Duration {
	s.config.mutex.Lock()
	defer s.config.mutex.Unlock()
	return s.config.PublishInterval
}

func (s *sessionTracker) processSessions() {
	tic := time.Tick(s.interval())
	shutdown := shutdownSignals()
	for {
		select {
		case session := <-s.sessionChannel:
			s.appendSession(session)
		case <-tic:
			s.publishCollectedSessions()
		case sig := <-shutdown:
			s.flushSessionsAndRepeatSignal(shutdown, sig.(syscall.Signal))
		}
	}
}

func (s *sessionTracker) appendSession(session *Session) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	s.sessions = append(s.sessions, session)
}

func (s *sessionTracker) publishCollectedSessions() {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	oldSessions := s.sessions
	s.sessions = nil
	if len(oldSessions) > 0 {
		go func(s *sessionTracker) {
			err := s.publisher.publish(oldSessions)
			if err != nil {
				s.config.logf("%v", err)
			}
		}(s)
	}
}

func (s *sessionTracker) flushSessionsAndRepeatSignal(shutdown chan<- os.Signal, sig syscall.Signal) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	signal.Stop(shutdown)
	if len(s.sessions) > 0 {
		err := s.publisher.publish(s.sessions)
		if err != nil {
			s.config.logf("%v", err)
		}
	}
	syscall.Kill(syscall.Getpid(), sig)
}

func (s *sessionTracker) FlushSessions() {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	sessions := s.sessions
	s.sessions = nil
	if len(sessions) != 0 {
		if err := s.publisher.publish(sessions); err != nil {
			s.config.logf("%v", err)
		}
	}
}

func shutdownSignals() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	return c
}
