package honeybadger_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	honeybadgervendor "github.com/honeybadger-io/honeybadger-go"
	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/honeybadger"
	"github.com/stretchr/testify/suite"
)

var errSkipSend = errors.New("skip send")

type HoneybadgerTestSuite struct {
	suite.Suite

	captured *honeybadgervendor.Notice
	reporter *honeybadger.ReporterIface
}

func TestHoneybadger(t *testing.T) {
	suite.Run(t, new(HoneybadgerTestSuite))
}

func (s *HoneybadgerTestSuite) SetupTest() {
	s.captured = nil

	client := honeybadgervendor.New(honeybadgervendor.Configuration{APIKey: "test", Env: "test"})
	client.BeforeNotify(func(n *honeybadgervendor.Notice) error {
		s.captured = n
		return errSkipSend
	})

	s.reporter = honeybadger.NewWithClient(client)
}

func (s *HoneybadgerTestSuite) SetupSubTest() {
	s.SetupTest()
}

// Note: imgproxy's Honeybadger backend merges both request headers and meta
// into a single honeybadger.CGIData map (headers get an "HTTP_" prefix, meta
// keys don't), not into honeybadger.Notice.Context.
func (s *HoneybadgerTestSuite) TestReport() {
	cases := []struct {
		name        string
		withRequest bool
		meta        map[string]any
		wantCGIData honeybadgervendor.CGIData
	}{
		{
			name:        "request and meta",
			withRequest: true,
			meta:        map[string]any{"Request ID": "abc123", "Documentation URL": "https://example.com/docs"},
			wantCGIData: honeybadgervendor.CGIData{
				"HTTP_X_REQUEST_ID": "req-1",
				"REQUEST_ID":        "abc123",
				"DOCUMENTATION_URL": "https://example.com/docs",
			},
		},
		{
			name:        "request and nil meta",
			withRequest: true,
			wantCGIData: honeybadgervendor.CGIData{"HTTP_X_REQUEST_ID": "req-1"},
		},
		{
			name:        "nil request and meta",
			meta:        map[string]any{"Request ID": "abc123"},
			wantCGIData: honeybadgervendor.CGIData{"REQUEST_ID": "abc123"},
		},
		{
			name:        "nil request and nil meta",
			wantCGIData: honeybadgervendor.CGIData{},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			var req *http.Request
			if tc.withRequest {
				req = httptest.NewRequest(http.MethodGet, "http://example.com/image.jpg", nil)
				req.Header.Set("X-Request-Id", "req-1")
			}

			s.Require().NotPanics(func() {
				s.reporter.Report(errctx.NewTextError("boom", 0), req, tc.meta)
			})

			s.Require().NotNil(s.captured)
			s.Require().Equal(tc.wantCGIData, s.captured.CGIData)
		})
	}
}

func (s *HoneybadgerTestSuite) TestReportOverridesWrappedErrorType() {
	base := errors.New("boom")
	wrapped := errctx.Wrap(base)

	s.reporter.Report(wrapped, nil, nil)

	s.Require().NotNil(s.captured)
	s.Require().Equal(errctx.ErrorType(wrapped), s.captured.ErrorClass)
	s.Require().NotEqual("*errctx.WrappedError", s.captured.ErrorClass)
}
