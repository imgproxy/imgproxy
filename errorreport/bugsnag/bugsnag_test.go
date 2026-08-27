package bugsnag_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bugsnagvendor "github.com/bugsnag/bugsnag-go/v2"
	bugsnagtestutil "github.com/bugsnag/bugsnag-go/v2/testutil"
	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/bugsnag"
	"github.com/stretchr/testify/suite"
)

type capturedRequest struct {
	HTTPMethod string `json:"httpMethod"`
	URL        string `json:"url"`
}

type capturedException struct {
	ErrorClass string `json:"errorClass"`
}

type capturedEvent struct {
	Request    *capturedRequest          `json:"request"`
	Exceptions []capturedException       `json:"exceptions"`
	MetaData   map[string]map[string]any `json:"metaData"`
}

type capturedReport struct {
	Events []capturedEvent `json:"events"`
}

type BugsnagTestSuite struct {
	suite.Suite

	server   *httptest.Server
	reports  chan []byte
	reporter *bugsnag.ReporterIface
}

func TestBugsnag(t *testing.T) {
	suite.Run(t, new(BugsnagTestSuite))
}

func (s *BugsnagTestSuite) SetupTest() {
	s.server, s.reports = bugsnagtestutil.Setup()

	cfg := &bugsnag.Config{
		Key:       bugsnagtestutil.TestAPIKey,
		Stage:     "test",
		Endpoints: bugsnagvendor.Endpoints{Notify: s.server.URL},
	}

	r, err := bugsnag.New(cfg)
	s.Require().NoError(err)
	s.Require().NotNil(r)

	s.reporter = r
}

func (s *BugsnagTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *BugsnagTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *BugsnagTestSuite) TearDownSubTest() {
	s.TearDownTest()
}

func (s *BugsnagTestSuite) awaitReport() capturedReport {
	select {
	case body := <-s.reports:
		var report capturedReport
		s.Require().NoError(json.Unmarshal(body, &report))
		s.Require().Len(report.Events, 1)
		return report
	case <-time.After(2 * time.Second):
		s.Require().FailNow("timed out waiting for a Bugsnag report")
		return capturedReport{}
	}
}

func (s *BugsnagTestSuite) TestReport() {
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

			report := s.awaitReport()
			event := report.Events[0]

			if tc.wantRequest {
				s.Require().NotNil(event.Request)
				s.Require().Equal(http.MethodGet, event.Request.HTTPMethod)
				s.Require().Equal("http://example.com/image.jpg", event.Request.URL)
			} else {
				s.Require().Nil(event.Request)
			}

			if tc.wantMeta != nil {
				s.Require().Equal(tc.wantMeta, event.MetaData["Processing Context"])
			} else {
				s.Require().Empty(event.MetaData["Processing Context"])
			}
		})
	}
}

func (s *BugsnagTestSuite) TestReportOverridesWrappedErrorType() {
	base := errors.New("boom")
	wrapped := errctx.Wrap(base)

	s.reporter.Report(wrapped, nil, nil)

	report := s.awaitReport()
	event := report.Events[0]

	s.Require().NotEmpty(event.Exceptions)
	s.Require().Equal(errctx.ErrorType(wrapped), event.Exceptions[0].ErrorClass)
	s.Require().NotEqual("*errctx.WrappedError", event.Exceptions[0].ErrorClass)
}
