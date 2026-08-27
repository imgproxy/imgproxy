package sentry_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/sentry"
	"github.com/stretchr/testify/suite"
)

type mockTransport struct {
	events []*sentrygo.Event
}

func (m *mockTransport) Flush(timeout time.Duration) bool {
	return true
}

func (m *mockTransport) FlushWithContext(ctx context.Context) bool {
	return true
}

func (m *mockTransport) Configure(options sentrygo.ClientOptions) {
}

func (m *mockTransport) SendEvent(event *sentrygo.Event) {
	m.events = append(m.events, event)
}

func (m *mockTransport) Close() {
}

type SentryTestSuite struct {
	suite.Suite

	transport *mockTransport
	reporter  *sentry.ReporterIface
}

func TestSentry(t *testing.T) {
	suite.Run(t, new(SentryTestSuite))
}

func (s *SentryTestSuite) SetupTest() {
	s.transport = &mockTransport{}

	client, err := sentrygo.NewClient(sentrygo.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: s.transport,
	})
	s.Require().NoError(err)

	s.reporter = sentry.NewWithClient(client)
}

func (s *SentryTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *SentryTestSuite) TestReport() {
	cases := []struct {
		name        string
		withRequest bool
		meta        map[string]any
		wantRequest bool
		wantMeta    map[string]any
	}{
		{
			name:        "request and meta",
			withRequest: true,
			meta:        map[string]any{"Request ID": "abc123", "Documentation URL": "https://example.com/docs"},
			wantRequest: true,
			wantMeta:    map[string]any{"Request ID": "abc123", "Documentation URL": "https://example.com/docs"},
		},
		{
			name:        "request and nil meta",
			withRequest: true,
			wantRequest: true,
		},
		{
			name:     "nil request and meta",
			meta:     map[string]any{"Request ID": "abc123"},
			wantMeta: map[string]any{"Request ID": "abc123"},
		},
		{
			name: "nil request and nil meta",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			var req *http.Request
			if tc.withRequest {
				req = httptest.NewRequest(http.MethodGet, "http://example.com/image.jpg", nil)
			}

			s.Require().NotPanics(func() {
				s.reporter.Report(errctx.NewTextError("boom", 0), req, tc.meta)
			})

			s.Require().Len(s.transport.events, 1)
			event := s.transport.events[0]

			if tc.wantRequest {
				s.Require().NotNil(event.Request)
				s.Require().Equal(http.MethodGet, event.Request.Method)
			} else {
				s.Require().Nil(event.Request)
			}

			if tc.wantMeta != nil {
				s.Require().Equal(tc.wantMeta, event.Contexts["Processing context"])
			} else {
				s.Require().NotContains(event.Contexts, "Processing context")
			}
		})
	}
}

func (s *SentryTestSuite) TestReportOverridesWrappedErrorType() {
	base := errors.New("boom")
	wrapped := errctx.Wrap(base)

	s.reporter.Report(wrapped, nil, nil)

	s.Require().Len(s.transport.events, 1)
	event := s.transport.events[0]

	s.Require().NotEmpty(event.Exception)
	s.Require().Equal(errctx.ErrorType(wrapped), event.Exception[len(event.Exception)-1].Type)
	s.Require().NotEqual("*errctx.WrappedError", event.Exception[len(event.Exception)-1].Type)
}
