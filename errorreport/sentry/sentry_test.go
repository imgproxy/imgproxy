package sentry_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/sentry"
	"github.com/stretchr/testify/suite"
)

type SentryTestSuite struct {
	suite.Suite

	transport *sentrygo.MockTransport
	reporter  *sentry.ReporterIface
}

func TestSentry(t *testing.T) {
	suite.Run(t, new(SentryTestSuite))
}

func (s *SentryTestSuite) SetupTest() {
	s.transport = &sentrygo.MockTransport{}

	r, err := sentry.New(&sentry.Config{
		DSN:       "https://public@example.com/1",
		Transport: s.transport,
	})
	s.Require().NoError(err)
	s.Require().NotNil(r)

	s.reporter = r
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

			s.Require().Len(s.transport.Events(), 1)
			event := s.transport.Events()[0]

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

	s.Require().Len(s.transport.Events(), 1)
	event := s.transport.Events()[0]

	s.Require().NotEmpty(event.Exception)
	s.Require().Equal(errctx.ErrorType(wrapped), event.Exception[len(event.Exception)-1].Type)
	s.Require().NotEqual("*errctx.WrappedError", event.Exception[len(event.Exception)-1].Type)
}
