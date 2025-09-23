package responsewriter

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

type ResponseWriterSuite struct {
	suite.Suite
}

type writerTestCase struct {
	name   string
	req    http.Header
	res    http.Header
	config Config
	fn     func(*Writer)
}

func (s *ResponseWriterSuite) TestHeaderCases() {
	expires := time.Date(2030, 8, 1, 0, 0, 0, 0, time.UTC)
	expiresSeconds := strconv.Itoa(int(time.Until(expires).Seconds()))

	shortExpires := time.Now().Add(10 * time.Second)
	shortExpiresSeconds := strconv.Itoa(int(time.Until(shortExpires).Seconds()))

	writeResponseTimeout := 10 * time.Second

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
				WriteResponseTimeout:    writeResponseTimeout,
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
				WriteResponseTimeout:    writeResponseTimeout,
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
				WriteResponseTimeout:    writeResponseTimeout,
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
				WriteResponseTimeout:    writeResponseTimeout,
			},
		},
		{
			name: "Canonical_ValidURL",
			req:  http.Header{},
			res: http.Header{
				httpheaders.Link:                  []string{"<https://example.com/image.jpg>; rel=\"canonical\""},
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader:   true,
				DefaultTTL:           3600,
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.SetCanonical("https://example.com/image.jpg")
			},
		},
		{
			name: "Canonical_InvalidURL",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader:   true,
				DefaultTTL:           3600,
				WriteResponseTimeout: writeResponseTimeout,
			},
		},
		{
			name: "WriteCanonical_Disabled",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				SetCanonicalHeader:   false,
				DefaultTTL:           3600,
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.SetCanonical("https://example.com/image.jpg")
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
				DefaultTTL:           3600,
				FallbackImageTTL:     1,
				WriteResponseTimeout: writeResponseTimeout,
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
				DefaultTTL:           math.MaxInt32,
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.SetExpires(expires)
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
				DefaultTTL:           math.MaxInt32,
				FallbackImageTTL:     600,
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.SetIsFallbackImage()
				w.SetExpires(shortExpires)
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
				VaryValue:            "Accept, Sec-CH-DPR, DPR, Sec-CH-Width, Width",
				WriteResponseTimeout: writeResponseTimeout,
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
			config: Config{
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.Passthrough("X-Test")
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
			config: Config{
				WriteResponseTimeout: writeResponseTimeout,
			},
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
			config: Config{
				WriteResponseTimeout: writeResponseTimeout,
			},
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
			config: Config{
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.SetContentType("image/png")
			},
		},
		{
			name: "SetMaxAgeFromExpiresZero",
			req:  http.Header{},
			res: http.Header{
				httpheaders.CacheControl:          []string{"max-age=3600, public"},
				httpheaders.ContentSecurityPolicy: []string{"script-src 'none'"},
			},
			config: Config{
				DefaultTTL:           3600,
				WriteResponseTimeout: writeResponseTimeout,
			},
			fn: func(w *Writer) {
				w.SetExpires(time.Time{})
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			factory, err := NewFactory(&tc.config)
			s.Require().NoError(err)

			r := httptest.NewRecorder()

			writer := factory.NewWriter(r)
			writer.SetOriginHeaders(tc.req)

			if tc.fn != nil {
				tc.fn(writer)
			}

			writer.WriteHeader(http.StatusOK)

			s.Require().Equal(tc.res, r.Header())
		})
	}
}

func TestHeaderWriter(t *testing.T) {
	suite.Run(t, new(ResponseWriterSuite))
}
