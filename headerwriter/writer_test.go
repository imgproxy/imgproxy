package headerwriter

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/stretchr/testify/suite"
)

type HeaderWriterSuite struct {
	suite.Suite
}

type writerTestCase struct {
	name   string
	url    string
	req    http.Header
	res    http.Header
	config Config
	fn     func(*Writer)
}

func (s *HeaderWriterSuite) TestHeaderCases() {
	expires := time.Date(2030, 8, 1, 0, 0, 0, 0, time.UTC)
	expiresSeconds := strconv.Itoa(int(time.Until(expires).Seconds()))

	shortExpires := time.Now().Add(10 * time.Second)
	shortExpiresSeconds := strconv.Itoa(int(time.Until(shortExpires).Seconds()))

	tt := []writerTestCase{
		{
			name: "MinimalHeaders",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader:      false,
				DefaultTTL:              0,
				CacheControlPassthrough: false,
				LastModifiedEnabled:     false,
				EnableClientHints:       false,
				SetVaryAccept:           false,
			},
		},
		{
			name: "PassthroughCacheControl",
			req: http.Header{
				httpheaders.CacheControl: []string{"no-cache, no-store, must-revalidate"},
			},
			res: http.Header{
				httpheaders.CacheControl:          []string{"no-cache, no-store, must-revalidate"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				CacheControlPassthrough: true,
				DefaultTTL:              3600,
			},
		},
		{
			name: "PassthroughCacheControlExpires",
			req: http.Header{
				httpheaders.Expires: []string{expires.Format(http.TimeFormat)},
			},
			res: http.Header{
				httpheaders.CacheControl:          []string{fmt.Sprintf("max-age=%s, public", expiresSeconds)},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				CacheControlPassthrough: true,
				DefaultTTL:              3600,
			},
		},
		{
			name: "PassthroughCacheControlExpiredInThePast",
			req: http.Header{
				httpheaders.Expires: []string{time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)},
			},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				CacheControlPassthrough: true,
				DefaultTTL:              3600,
			},
		},
		{
			name: "Canonical_ValidURL",
			req:  http.Header{},
			url:  "https://example.com/image.jpg",
			res: http.Header{
				httpheaders.Link:                  []string{"<https://example.com/image.jpg>; rel=\"canonical\""},
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader: true,
				DefaultTTL:         3600,
			},
			fn: func(w *Writer) {
				w.SetCanonical()
			},
		},
		{
			name: "Canonical_InvalidURL",
			url:  "ftp://example.com/image.jpg",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader: true,
				DefaultTTL:         3600,
			},
		},
		{
			name: "WriteCanonical_Disabled",
			req:  http.Header{},
			url:  "https://example.com/image.jpg",
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader: false,
				DefaultTTL:         3600,
			},
			fn: func(w *Writer) {
				w.SetCanonical()
			},
		},
		{
			name: "LastModified",
			req: http.Header{
				httpheaders.LastModified: []string{expires.Format(http.TimeFormat)},
			},
			res: http.Header{
				httpheaders.LastModified:          []string{expires.Format(http.TimeFormat)},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
			},
			config: Config{
				LastModifiedEnabled: true,
				DefaultTTL:          3600,
			},
			fn: func(w *Writer) {
				w.SetLastModified()
			},
		},
		{
			name: "SetMaxAgeTTL",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=1, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				DefaultTTL:       3600,
				FallbackImageTTL: 1,
			},
			fn: func(w *Writer) {
				w.SetIsFallbackImage()
			},
		},
		{
			name: "SetMaxAgeExpires",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{fmt.Sprintf("max-age=%s, public", expiresSeconds)},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				DefaultTTL: math.MaxInt32,
			},
			fn: func(w *Writer) {
				w.SetExpires(&expires)
			},
		},
		{
			name: "SetMaxAgeTTLOutlivesExpires",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{fmt.Sprintf("max-age=%s, public", shortExpiresSeconds)},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				DefaultTTL:       math.MaxInt32,
				FallbackImageTTL: 600,
			},
			fn: func(w *Writer) {
				w.SetIsFallbackImage()
				w.SetExpires(&shortExpires)
			},
		},
		{
			name: "SetVaryHeader",
			req:  http.Header{},
			res: http.Header{
				httpheaders.Vary:                  []string{"Accept, Sec-CH-DPR, DPR, Sec-CH-Width, Width"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				EnableClientHints: true,
				SetVaryAccept:     true,
			},
			fn: func(w *Writer) {
				w.SetVary()
			},
		},
		{
			name: "PassthroughHeaders",
			req: http.Header{
				"X-Test": []string{"foo", "bar"},
			},
			res: http.Header{
				"X-Test":                          []string{"foo", "bar"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{},
			fn: func(w *Writer) {
				w.Passthrough([]string{"X-Test"})
			},
		},
		{
			name: "CopyFromHeaders",
			req:  http.Header{},
			res: http.Header{
				"X-From":                          []string{"baz"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{},
			fn: func(w *Writer) {
				h := http.Header{}
				h.Set("X-From", "baz")
				w.CopyFrom(h, []string{"X-From"})
			},
		},
		{
			name: "WriteContentLength",
			req:  http.Header{},
			res: http.Header{
				httpheaders.ContentLength:         []string{"123"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{},
			fn: func(w *Writer) {
				w.SetContentLength(123)
			},
		},
		{
			name: "WriteContentType",
			req:  http.Header{},
			res: http.Header{
				httpheaders.ContentType:           []string{"image/png"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{},
			fn: func(w *Writer) {
				w.SetContentType("image/png")
			},
		},
		{
			name: "SetMaxAgeFromExpiresNil",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				DefaultTTL: 3600,
			},
			fn: func(w *Writer) {
				w.SetExpires(nil)
			},
		},
		{
			name: "WriteVaryAcceptOnly",
			req:  http.Header{},
			res: http.Header{
				httpheaders.Vary:                  []string{"Accept"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetVaryAccept: true,
			},
			fn: func(w *Writer) {
				w.SetVary()
			},
		},
		{
			name: "WriteVaryClientHintsOnly",
			req:  http.Header{},
			res: http.Header{
				httpheaders.Vary:                  []string{"Sec-CH-DPR, DPR, Sec-CH-Width, Width"},
				httpheaders.CacheControl:          []string{"no-cache"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				EnableClientHints: true,
			},
			fn: func(w *Writer) {
				w.SetVary()
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			writer := New(&tc.config, tc.req, tc.url)

			if tc.fn != nil {
				tc.fn(writer)
			}

			r := httptest.NewRecorder()
			writer.Write(r)

			s.Require().Equal(tc.res, r.Header())
		})
	}
}

func TestHeaderWriter(t *testing.T) {
	suite.Run(t, new(HeaderWriterSuite))
}
