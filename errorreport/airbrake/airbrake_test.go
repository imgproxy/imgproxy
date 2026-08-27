package airbrake_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/airbrake/gobrake/v5"
	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/airbrake"
	"github.com/stretchr/testify/suite"
)

type AirbrakeTestSuite struct {
	suite.Suite

	captured chan *gobrake.Notice
	reporter *airbrake.ReporterIface
}

func TestAirbrake(t *testing.T) {
	suite.Run(t, new(AirbrakeTestSuite))
}

func (s *AirbrakeTestSuite) SetupTest() {
	notifier := gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
		ProjectId:  1,
		ProjectKey: "test",
	})

	s.captured = make(chan *gobrake.Notice, 1)
	notifier.AddFilter(func(n *gobrake.Notice) *gobrake.Notice {
		s.captured <- n
		return nil
	})

	s.reporter = airbrake.NewWithNotifier(notifier)
}

func (s *AirbrakeTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *AirbrakeTestSuite) awaitNotice() *gobrake.Notice {
	select {
	case n := <-s.captured:
		return n
	case <-time.After(2 * time.Second):
		s.Require().FailNow("timed out waiting for an Airbrake notice")
		return nil
	}
}

func (s *AirbrakeTestSuite) TestReport() {
	cases := []struct {
		name        string
		withRequest bool
		meta        map[string]any
		wantContext map[string]any
		absentKeys  []string
	}{
		{
			name:        "request and meta",
			withRequest: true,
			meta:        map[string]any{"Request ID": "abc123", "Documentation URL": "https://example.com/docs"},
			wantContext: map[string]any{
				"url":               "http://example.com/image.jpg",
				"httpMethod":        http.MethodGet,
				"request-id":        "abc123",
				"documentation-url": "https://example.com/docs",
			},
		},
		{
			name:        "request and nil meta",
			withRequest: true,
			wantContext: map[string]any{
				"url":        "http://example.com/image.jpg",
				"httpMethod": http.MethodGet,
			},
		},
		{
			name:        "nil request and meta",
			meta:        map[string]any{"Request ID": "abc123"},
			wantContext: map[string]any{"request-id": "abc123"},
			absentKeys:  []string{"url", "httpMethod"},
		},
		{
			name:       "nil request and nil meta",
			absentKeys: []string{"url", "httpMethod"},
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

			notice := s.awaitNotice()

			s.Require().Subset(notice.Context, tc.wantContext)
			for _, key := range tc.absentKeys {
				s.Require().NotContains(notice.Context, key)
			}
		})
	}
}

func (s *AirbrakeTestSuite) TestReportOverridesWrappedErrorType() {
	base := errors.New("boom")
	wrapped := errctx.Wrap(base)

	s.reporter.Report(wrapped, nil, nil)

	notice := s.awaitNotice()

	s.Require().NotEmpty(notice.Errors)
	s.Require().Equal(errctx.ErrorType(wrapped), notice.Errors[0].Type)
	s.Require().NotEqual("*errctx.WrappedError", notice.Errors[0].Type)
}
