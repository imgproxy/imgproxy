package auximageprovider

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

type ImageProviderTestSuite struct {
	suite.Suite

	server      *httptest.Server
	testData    []byte
	testDataB64 string

	// Server state
	status int
	data   []byte
	header http.Header
}

func (s *ImageProviderTestSuite) SetupSuite() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true

	// Load test image data
	f, err := os.Open("../testdata/test1.jpg")
	s.Require().NoError(err)
	defer f.Close()

	data, err := io.ReadAll(f)
	s.Require().NoError(err)

	s.testData = data
	s.testDataB64 = base64.StdEncoding.EncodeToString(data)

	// Create test server
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		for k, vv := range s.header {
			for _, v := range vv {
				rw.Header().Add(k, v)
			}
		}

		data := s.data
		if data == nil {
			data = s.testData
		}

		rw.Header().Set(httpheaders.ContentLength, strconv.Itoa(len(data)))
		rw.WriteHeader(s.status)
		rw.Write(data)
	}))

	s.Require().NoError(imagedata.Init())
}

func (s *ImageProviderTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *ImageProviderTestSuite) SetupTest() {
	s.status = http.StatusOK
	s.data = nil
	s.header = http.Header{}
	s.header.Set(httpheaders.ContentType, "image/jpeg")
}

// Helper function to read data from ImageData
func (s *ImageProviderTestSuite) readImageData(provider Provider) []byte {
	imgData, _, err := provider.Get(s.T().Context(), &options.ProcessingOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(imgData)
	defer imgData.Close()

	reader := imgData.Reader()
	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	return data
}

func (s *ImageProviderTestSuite) TestNewProvider() {
	tests := []struct {
		name         string
		config       *StaticConfig
		setupFunc    func()
		expectError  bool
		expectNil    bool
		validateFunc func(provider Provider)
	}{
		{
			name:   "B64",
			config: &StaticConfig{Base64Data: s.testDataB64},
			validateFunc: func(provider Provider) {
				s.Equal(s.testData, s.readImageData(provider))
			},
		},
		{
			name:   "Path",
			config: &StaticConfig{Path: "../testdata/test1.jpg"},
			validateFunc: func(provider Provider) {
				s.Equal(s.testData, s.readImageData(provider))
			},
		},
		{
			name:   "URL",
			config: &StaticConfig{URL: s.server.URL},
			validateFunc: func(provider Provider) {
				s.Equal(s.testData, s.readImageData(provider))
			},
		},
		{
			name:      "EmptyConfig",
			config:    &StaticConfig{},
			expectNil: true,
		},
		{
			name:        "InvalidURL",
			config:      &StaticConfig{URL: "http://invalid-url-that-does-not-exist.invalid"},
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "InvalidBase64",
			config:      &StaticConfig{Base64Data: "invalid-base64-data!!!"},
			expectError: true,
			expectNil:   true,
		},
		{
			name: "Base64PreferenceOverPath",
			config: &StaticConfig{
				Base64Data: base64.StdEncoding.EncodeToString(s.testData),
				Path:       "../testdata/test2.jpg", // This should be ignored
			},
			validateFunc: func(provider Provider) {
				actualData := s.readImageData(provider)
				s.Equal(s.testData, actualData)
			},
		},
		{
			name:   "HeadersPassedThrough",
			config: &StaticConfig{URL: s.server.URL},
			setupFunc: func() {
				s.header.Set("X-Custom-Header", "test-value")
				s.header.Set(httpheaders.CacheControl, "max-age=3600")
			},
			validateFunc: func(provider Provider) {
				imgData, headers, err := provider.Get(s.T().Context(), &options.ProcessingOptions{})
				s.Require().NoError(err)
				s.Require().NotNil(imgData)
				defer imgData.Close()

				s.Equal("test-value", headers.Get("X-Custom-Header"))
				s.Equal("max-age=3600", headers.Get(httpheaders.CacheControl))
				s.Equal("image/jpeg", headers.Get(httpheaders.ContentType))
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			provider, err := NewStaticProvider(s.T().Context(), tt.config, "test image")

			if tt.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			if tt.expectNil {
				s.Nil(provider)
			} else {
				s.Require().NotNil(provider)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(provider)
			}
		})
	}
}

func TestImageProvider(t *testing.T) {
	suite.Run(t, new(ImageProviderTestSuite))
}
