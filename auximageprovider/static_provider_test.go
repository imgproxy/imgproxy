package auximageprovider

import (
	"encoding/base64"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type ImageProviderTestSuite struct {
	testutil.LazySuite

	testData    []byte
	testDataB64 string

	testServer testutil.LazyTestServer
	idf        *imagedata.Factory
}

func (s *ImageProviderTestSuite) SetupSuite() {
	s.testData = testutil.NewTestDataProvider(s.T).Read("test1.jpg")
	s.testDataB64 = base64.StdEncoding.EncodeToString(s.testData)

	fc := fetcher.NewDefaultConfig()
	fc.Transport.HTTP.AllowLoopbackSourceAddresses = true

	f, err := fetcher.New(&fc)
	s.Require().NoError(err)

	s.idf = imagedata.NewFactory(f, nil)

	s.testServer, _ = testutil.NewLazySuiteTestServer(
		s,
		func(srv *testutil.TestServer) error {
			srv.SetHeaders(
				httpheaders.ContentType, "image/jpeg",
				httpheaders.ContentLength, strconv.Itoa(len(s.testData)),
			).SetBody(s.testData)

			return nil
		},
	)
}

func (s *ImageProviderTestSuite) SetupSubTest() {
	// We use t.Run() a lot, so we need to reset lazy objects at the beginning of each subtest
	s.ResetLazyObjects()
}

// Helper function to read data from ImageData
func (s *ImageProviderTestSuite) readImageData(provider Provider) []byte {
	imgData, _, err := provider.Get(s.T().Context(), options.New())
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
			config: &StaticConfig{URL: s.testServer().URL()},
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
			config: &StaticConfig{URL: s.testServer().URL()},
			setupFunc: func() {
				s.testServer().SetHeaders(
					"X-Custom-Header", "test-value",
					httpheaders.CacheControl, "max-age=3600",
				)
			},
			validateFunc: func(provider Provider) {
				imgData, headers, err := provider.Get(s.T().Context(), options.New())
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

			provider, err := NewStaticProvider(s.T().Context(), tt.config, "test image", s.idf)

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
