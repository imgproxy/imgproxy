package auximageprovider

import (
	"context"
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
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
)

type StaticProviderTestSuite struct {
	suite.Suite

	server      *httptest.Server
	testData    []byte
	testDataB64 string

	// Server state
	status int
	data   []byte
	header http.Header
}

func (s *StaticProviderTestSuite) SetupSuite() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true

	// Load test image data
	data, err := os.ReadFile("../testdata/test1.jpg")
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

func (s *StaticProviderTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *StaticProviderTestSuite) SetupTest() {
	s.status = http.StatusOK
	s.data = nil
	s.header = http.Header{}
	s.header.Set(httpheaders.ContentType, "image/jpeg")
}

// Helper function to read data from ImageData
func (s *StaticProviderTestSuite) readImageData(provider Provider) []byte {
	imgData, _, err := provider.Get(s.T().Context(), &options.ProcessingOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(imgData)
	defer imgData.Close()

	reader := imgData.Reader()
	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	return data
}

func (s *StaticProviderTestSuite) TestNewFromFile() {
	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{Path: "../testdata/test1.jpg"},
	)
	s.Require().NoError(err)
	s.Require().NotNil(provider)

	// Test Get method
	imgData, headers, err := provider.Get(s.T().Context(), &options.ProcessingOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(imgData)
	s.Require().NotNil(headers)
	defer imgData.Close()

	// Verify image data
	reader := imgData.Reader()
	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal(s.testData, data)

	// Verify image format
	s.Equal(imagetype.JPEG, imgData.Format())
}

func (s *StaticProviderTestSuite) TestNewFromFileNonExistent() {
	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{Path: "../testdata/non-existent.jpg"},
	)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *StaticProviderTestSuite) TestNewFromBase64() {
	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{Base64Data: s.testDataB64},
	)
	s.Require().NoError(err)
	s.Require().NotNil(provider)

	// Test Get method
	imgData, headers, err := provider.Get(s.T().Context(), &options.ProcessingOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(imgData)
	s.Require().NotNil(headers)
	defer imgData.Close()

	// Verify image data
	reader := imgData.Reader()
	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal(s.testData, data)

	// Verify image format
	s.Equal(imagetype.JPEG, imgData.Format())
}

func (s *StaticProviderTestSuite) TestNewFromBase64Invalid() {
	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{Base64Data: "invalid-base64"},
	)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *StaticProviderTestSuite) TestNewFromBase64InvalidImage() {
	invalidB64 := base64.StdEncoding.EncodeToString([]byte("not an image"))
	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{Base64Data: invalidB64},
	)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *StaticProviderTestSuite) TestNewFromURL() {
	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{URL: s.server.URL},
	)
	s.Require().NoError(err)
	s.Require().NotNil(provider)

	// Test Get method
	imgData, headers, err := provider.Get(s.T().Context(), &options.ProcessingOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(imgData)
	s.Require().NotNil(headers)
	defer imgData.Close()

	// Verify image data
	reader := imgData.Reader()
	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal(s.testData, data)

	// Verify image format
	s.Equal(imagetype.JPEG, imgData.Format())
}

func (s *StaticProviderTestSuite) TestNewFromURLNotFound() {
	s.status = http.StatusNotFound
	s.data = []byte("Not Found")
	s.header.Set(httpheaders.ContentType, "text/plain")

	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{URL: s.server.URL},
	)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *StaticProviderTestSuite) TestNewFromURLInvalidImage() {
	s.data = []byte("not an image")

	provider, err := NewStaticProvider(
		context.Background(),
		&StaticConfig{URL: s.server.URL},
	)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *StaticProviderTestSuite) TestNewFromTriple() {
	testData2, err := os.ReadFile("../testdata/test1.png")
	s.Require().NoError(err)

	testData3, err := os.ReadFile("../testdata/test1.svg")
	s.Require().NoError(err)

	testData3B64 := base64.StdEncoding.EncodeToString(testData3)

	testCases := []struct {
		name   string
		cfg    *StaticConfig
		isNil  bool
		expect []byte
	}{
		{
			name: "All three (base64 should prefer)",
			cfg: &StaticConfig{
				Base64Data: testData3B64,
				Path:       "../testdata/test1.png",
				URL:        s.server.URL,
			},
			expect: testData3,
		},
		{
			name: "File path and URL (file should prefer)",
			cfg: &StaticConfig{
				Path: "../testdata/test1.png",
				URL:  s.server.URL,
			},
			expect: testData2,
		},
		{
			name: "Only URL",
			cfg: &StaticConfig{
				URL: s.server.URL,
			},
			expect: s.testData,
		},
		{
			name:  "No inputs",
			cfg:   &StaticConfig{},
			isNil: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			provider, err := NewStaticProvider(s.T().Context(), tc.cfg)
			s.Require().NoError(err)

			if tc.isNil {
				s.Nil(provider)
			} else {
				s.Require().NotNil(provider)
				s.Equal(tc.expect, s.readImageData(provider))
			}
		})
	}
}

func TestStaticProvider(t *testing.T) {
	suite.Run(t, new(StaticProviderTestSuite))
}
