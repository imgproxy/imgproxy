package processing

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ConditionalHeadersSuite struct {
	testutil.LazySuite

	Config testutil.LazyObj[*Config]
}

func (s *ConditionalHeadersSuite) SetupSuite() {
	s.Config, _ = testutil.NewLazySuiteObj(s, func() (*Config, error) {
		cfg := NewDefaultConfig()
		return &cfg, nil
	})
}

func (s *ConditionalHeadersSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

// helper creates a minimal request object with given config and optional headers
func (s *ConditionalHeadersSuite) makeReq(ifMod, ifNone string) *http.Request {
	req := &http.Request{Header: make(http.Header)}

	if len(ifMod) > 0 {
		req.Header.Set(httpheaders.IfModifiedSince, ifMod)
	}

	if len(ifNone) > 0 {
		req.Header.Set(httpheaders.IfNoneMatch, ifNone)
	}

	return req
}

func (s *ConditionalHeadersSuite) TestNewFromRequest() {
	inMod := "Mon, 02 Jan 2006 15:04:05 GMT"
	inNone := `"etag"`

	r := s.makeReq(inMod, inNone)
	c := NewConditionalHeadersFromRequest(s.Config(), r)

	req := s.Require()
	req.Equal(inMod, c.ifModifiedSince)
	req.Equal(inNone, c.ifNoneMatch)
}

func (s *ConditionalHeadersSuite) TestInjectImageRequestHeaders() {
	buster := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	inModAfterBuster := buster.Add(+time.Hour).Format(http.TimeFormat)
	inModBeforeBuster := buster.Add(-time.Hour).Format(http.TimeFormat)

	s.Config().LastModifiedBuster = buster

	cases := []struct {
		name                string
		ifModifiedSince     string
		ifNoneMatch         string
		etagSentToSource    string
		lastModifiedEnabled bool
		lastModifiedBuster  time.Time
		etagEnabled         bool
		etagBuster          string
		wantIfModifiedSince bool
		wantIfNoneMatch     bool
	}{
		{
			name: "all disabled",
		},
		{
			name:                "last modified enabled",
			ifModifiedSince:     inModAfterBuster,
			lastModifiedEnabled: true,
			wantIfModifiedSince: true,
		},
		{
			name:             "etag enabled",
			ifNoneMatch:      "etag1",
			etagSentToSource: "etag1",
			etagEnabled:      true,
			wantIfNoneMatch:  true,
		},
		{
			name:                "etag/lm enabled with no buster",
			ifModifiedSince:     inModAfterBuster,
			ifNoneMatch:         "etag1",
			etagSentToSource:    "etag1",
			lastModifiedEnabled: true,
			etagEnabled:         true,
			wantIfModifiedSince: true,
			wantIfNoneMatch:     true,
		},
		{
			name:                "last modified after buster",
			ifModifiedSince:     inModAfterBuster,
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			wantIfModifiedSince: true,
		},
		{
			name:                "last modified before buster",
			ifModifiedSince:     inModBeforeBuster,
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			wantIfModifiedSince: false,
		},
		{
			name:                "last modified enabled with buster, invalid header",
			ifModifiedSince:     "not-a-time",
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			wantIfModifiedSince: false, // invalid header should be treated as if it's before buster
		},
		{
			name:             "etag buster",
			ifNoneMatch:      "buster/" + base64.RawURLEncoding.EncodeToString([]byte("etag1")),
			etagSentToSource: "etag1",
			etagEnabled:      true,
			etagBuster:       "buster",
			wantIfNoneMatch:  true,
		},
		{
			name:                "both enabled, but last modified buster check fails, etag still works",
			ifModifiedSince:     inModBeforeBuster,
			ifNoneMatch:         "buster/" + base64.RawURLEncoding.EncodeToString([]byte("etag1")),
			etagSentToSource:    "etag1",
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			etagEnabled:         true,
			etagBuster:          "buster",
			wantIfModifiedSince: false, // should not be sent because of buster
			wantIfNoneMatch:     true,  // should not be sent because of buster
		},
		{
			name:                "both enabled, if-none-match with invalid base64",
			ifModifiedSince:     inModAfterBuster,
			ifNoneMatch:         "buster/@@@",
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			etagEnabled:         true,
			etagBuster:          "buster",
			wantIfModifiedSince: false, // should be sent because last modified buster check passes
			wantIfNoneMatch:     false, // should not be sent because of etag buster check fails
		},
		{
			name:                "both enabled, but etag buster different",
			ifModifiedSince:     inModAfterBuster,
			ifNoneMatch:         "buster/" + base64.RawURLEncoding.EncodeToString([]byte("etag1")),
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			etagEnabled:         true,
			etagBuster:          "different-buster",
			wantIfModifiedSince: false, // should be sent because last modified buster check passes
			wantIfNoneMatch:     false, // should not be sent because of etag buster check fails
		},
		{
			name:                "both enabled, but etag with no buster",
			ifModifiedSince:     inModAfterBuster,
			ifNoneMatch:         "W/etag1",
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			etagEnabled:         true,
			etagBuster:          "buster",
			wantIfModifiedSince: false, // should be sent because last modified buster check passes
			wantIfNoneMatch:     false, // should not be sent because of etag buster check fails
		},
	}

	for _, c := range cases {
		s.Run(c.name, func() {
			s.Config().LastModifiedEnabled = c.lastModifiedEnabled
			s.Config().LastModifiedBuster = c.lastModifiedBuster
			s.Config().ETagEnabled = c.etagEnabled
			s.Config().ETagBuster = c.etagBuster

			r := s.makeReq(c.ifModifiedSince, c.ifNoneMatch)
			ch := NewConditionalHeadersFromRequest(s.Config(), r)
			h := make(http.Header)
			ch.InjectImageRequestHeaders(h)

			if c.wantIfModifiedSince {
				s.Require().Equal(c.ifModifiedSince, h.Get(httpheaders.IfModifiedSince))
			} else {
				s.Require().Empty(h.Get(httpheaders.IfModifiedSince))
			}
			if c.wantIfNoneMatch {
				s.Require().Equal(c.etagSentToSource, h.Get(httpheaders.IfNoneMatch))
			} else {
				s.Require().Empty(h.Get(httpheaders.IfNoneMatch))
			}
		})
	}
}

func (s *ConditionalHeadersSuite) TestInjectUserResponseHeaders() {
	buster := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	busterFmt := buster.Format(http.TimeFormat)
	lastModAfterBuster := buster.Add(time.Hour).Format(http.TimeFormat)
	lastModBeforeBuster := buster.Add(-time.Hour).Format(http.TimeFormat)

	cases := []struct {
		name                string
		imageEtag           string
		imageLastModified   string
		lastModifiedEnabled bool
		lastModifiedBuster  time.Time
		etagEnabled         bool
		etagBuster          string
		wantLastModified    string
		wantEtag            string
	}{
		{
			name:                "all disabled",
			lastModifiedEnabled: false,
			etagEnabled:         false,
		},
		{
			name:                "lastmod enabled no header",
			lastModifiedEnabled: true,
		},
		{
			name:                "lastmod older than buster",
			imageLastModified:   lastModBeforeBuster,
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			wantLastModified:    busterFmt,
		},
		{
			name:                "lastmod after buster",
			imageLastModified:   lastModAfterBuster,
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			wantLastModified:    lastModAfterBuster,
		},
		{
			name:                "lastmod invalid value",
			imageLastModified:   "not-a-time",
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			wantLastModified:    "",
		},
		{
			name:        "etag enabled empty",
			etagEnabled: true,
		},
		{
			name:        "etag enabled with value",
			imageEtag:   "etag1",
			etagEnabled: true,
			etagBuster:  "buster",
			wantEtag:    `"buster/` + base64.RawURLEncoding.EncodeToString([]byte("etag1")) + `"`,
		},
		{
			name:                "both enabled, both present, lm older",
			imageLastModified:   lastModBeforeBuster,
			imageEtag:           "etag2",
			lastModifiedEnabled: true,
			lastModifiedBuster:  buster,
			etagEnabled:         true,
			etagBuster:          "buster",
			wantLastModified:    busterFmt,
			wantEtag:            `"buster/` + base64.RawURLEncoding.EncodeToString([]byte("etag2")) + `"`,
		},
	}

	for _, c := range cases {
		s.Run(c.name, func() {
			s.Config().LastModifiedEnabled = c.lastModifiedEnabled
			s.Config().LastModifiedBuster = c.lastModifiedBuster
			s.Config().ETagEnabled = c.etagEnabled
			s.Config().ETagBuster = c.etagBuster

			// prepare image response headers
			imageRes := make(http.Header)
			if len(c.imageEtag) > 0 {
				imageRes.Set(httpheaders.Etag, c.imageEtag)
			}
			if len(c.imageLastModified) > 0 {
				imageRes.Set(httpheaders.LastModified, c.imageLastModified)
			}

			// prepare request and response writer
			r := s.makeReq("", "")
			rec := httptest.NewRecorder()

			ch := NewConditionalHeadersFromRequest(s.Config(), r)
			ch.SetOriginHeaders(imageRes)
			ch.InjectUserResponseHeaders(rec)

			// assert Last-Modified
			lm := rec.Header().Get(httpheaders.LastModified)
			if c.wantLastModified == "" {
				s.Require().Empty(lm)
			} else {
				s.Require().Equal(c.wantLastModified, lm)
			}

			// assert ETag
			etag := rec.Header().Get(httpheaders.Etag)
			if c.wantEtag == "" {
				s.Require().Empty(etag)
			} else {
				s.Require().Equal(c.wantEtag, etag)
			}
		})
	}
}

func TestConditionalHeadersSuite(t *testing.T) {
	suite.Run(t, new(ConditionalHeadersSuite))
}
