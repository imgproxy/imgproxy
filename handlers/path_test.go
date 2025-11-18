package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/stretchr/testify/suite"
)

type PathTestSuite struct {
	suite.Suite
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, new(PathTestSuite))
}

func (s *PathTestSuite) createRequest(path string) *http.Request {
	return httptest.NewRequest("GET", path, nil)
}

func (s *PathTestSuite) TestParsePath() {
	testCases := []struct {
		name          string
		pathPrefix    string
		requestPath   string
		expectedPath  string
		expectedSig   string
		expectedError bool
	}{
		{
			name:          "BasicPath",
			requestPath:   "/dummy_signature/rs:fill:300:200/plain/http://example.com/image.jpg",
			expectedPath:  "rs:fill:300:200/plain/http://example.com/image.jpg",
			expectedSig:   "dummy_signature",
			expectedError: false,
		},
		{
			name:          "PathWithQueryParams",
			requestPath:   "/dummy_signature/rs:fill:300:200/plain/http://example.com/image.jpg?param1=value1&param2=value2",
			expectedPath:  "rs:fill:300:200/plain/http://example.com/image.jpg",
			expectedSig:   "dummy_signature",
			expectedError: false,
		},
		{
			name:          "PathWithPrefix",
			pathPrefix:    "/imgproxy",
			requestPath:   "/imgproxy/dummy_signature/rs:fill:300:200/plain/http://example.com/image.jpg",
			expectedPath:  "rs:fill:300:200/plain/http://example.com/image.jpg",
			expectedSig:   "dummy_signature",
			expectedError: false,
		},
		{
			name:          "PathWithRedenormalization",
			requestPath:   "/dummy_signature/rs:fill:300:200/plain/https:/example.com/path/to/image.jpg",
			expectedPath:  "rs:fill:300:200/plain/https://example.com/path/to/image.jpg",
			expectedSig:   "dummy_signature",
			expectedError: false,
		},
		{
			name:          "NoSignatureSeparator",
			requestPath:   "/invalid_path_without_slash",
			expectedPath:  "",
			expectedSig:   "",
			expectedError: true,
		},
		{
			name:          "EmptyPath",
			requestPath:   "/",
			expectedPath:  "",
			expectedSig:   "",
			expectedError: true,
		},
		{
			name:          "OnlySignature",
			requestPath:   "/signature_only",
			expectedPath:  "",
			expectedSig:   "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {

			req := s.createRequest(tc.requestPath)
			req.Pattern = tc.pathPrefix
			path, signature, err := SplitPathSignature(req)

			if tc.expectedError {
				var ierr errctx.Error

				s.Require().Error(err)
				s.Require().ErrorAs(err, &ierr)
				s.Require().Equal(CategoryPathParsing, ierr.Category())

				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedPath, path)
			s.Require().Equal(tc.expectedSig, signature)
		})
	}
}

func (s *PathTestSuite) TestRedenormalizePathHTTPProtocol() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTP",
			input:    "/plain/http:/example.com/image.jpg",
			expected: "/plain/http://example.com/image.jpg",
		},
		{
			name:     "HTTPS",
			input:    "/plain/https:/example.com/image.jpg",
			expected: "/plain/https://example.com/image.jpg",
		},
		{
			name:     "Local",
			input:    "/plain/local:/image.jpg",
			expected: "/plain/local:///image.jpg",
		},
		{
			name:     "NormalizedPath",
			input:    "/plain/http://example.com/image.jpg",
			expected: "/plain/http://example.com/image.jpg",
		},
		{
			name:     "ProtocolMissing",
			input:    "/rs:fill:300:200/plain/example.com/image.jpg",
			expected: "/rs:fill:300:200/plain/example.com/image.jpg",
		},
		{
			name:     "EmptyString",
			input:    "",
			expected: "",
		},
		{
			name:     "SingleSlash",
			input:    "/",
			expected: "/",
		},
		{
			name:     "NoPlainPrefix",
			input:    "/http:/example.com/image.jpg",
			expected: "/http:/example.com/image.jpg",
		},
		{
			name:     "NoProtocol",
			input:    "/plain/example.com/image.jpg",
			expected: "/plain/example.com/image.jpg",
		},
		{
			name:     "EndsWithProtocol",
			input:    "/plain/http:",
			expected: "/plain/http:",
		},
		{
			name:     "OnlyProtocol",
			input:    "/plain/http:/test",
			expected: "/plain/http://test",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := redenormalizePath(tc.input)
			s.Equal(tc.expected, result)
		})
	}
}
