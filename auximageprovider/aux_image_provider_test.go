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
	"github.com/imgproxy/imgproxy/v3/imagedownloader"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/transport"
)

type ImageProviderTestSuite struct {
	suite.Suite

	server      *httptest.Server
	factory     *Factory
	downloader  *imagedownloader.Downloader
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

	// Create downloader
	tr, err := transport.NewTransport()
	s.Require().NoError(err)

	fetcher, err := imagefetcher.NewFetcher(tr, imagefetcher.NewConfigFromEnv())
	s.Require().NoError(err)

	s.downloader = imagedownloader.NewDownloader(fetcher)
	s.factory = NewFactory(s.downloader)

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

		rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
		rw.WriteHeader(s.status)
		rw.Write(data)
	}))
}

func (s *ImageProviderTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *ImageProviderTestSuite) SetupTest() {
	s.status = http.StatusOK
	s.data = nil
	s.header = http.Header{}
	s.header.Set("Content-Type", "image/jpeg")
}

// Helper function to read data from ImageData
func (s *ImageProviderTestSuite) readImageData(provider AuxImageProvider) []byte {
	imgData, _, err := provider.Get(context.Background(), &options.ProcessingOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(imgData)
	defer imgData.Close()

	reader := imgData.Reader()
	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	return data
}

func (s *ImageProviderTestSuite) TestNewFromFile() {
	provider, err := s.factory.NewMemoryFromFile("../testdata/test1.jpg")
	s.Require().NoError(err)
	s.Require().NotNil(provider)

	// Test Get method
	imgData, headers, err := provider.Get(context.Background(), &options.ProcessingOptions{})
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

func (s *ImageProviderTestSuite) TestNewFromFileNonExistent() {
	provider, err := s.factory.NewMemoryFromFile("../testdata/non-existent.jpg")
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *ImageProviderTestSuite) TestNewFromBase64() {
	provider, err := s.factory.NewMemoryFromBase64(s.testDataB64)
	s.Require().NoError(err)
	s.Require().NotNil(provider)

	// Test Get method
	imgData, headers, err := provider.Get(context.Background(), &options.ProcessingOptions{})
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

func (s *ImageProviderTestSuite) TestNewFromBase64Invalid() {
	provider, err := s.factory.NewMemoryFromBase64("invalid-base64")
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *ImageProviderTestSuite) TestNewFromBase64InvalidImage() {
	invalidB64 := base64.StdEncoding.EncodeToString([]byte("not an image"))
	provider, err := s.factory.NewMemoryFromBase64(invalidB64)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *ImageProviderTestSuite) TestNewFromURL() {
	provider, err := s.factory.NewMemoryURL(context.Background(), s.server.URL)
	s.Require().NoError(err)
	s.Require().NotNil(provider)

	// Test Get method
	imgData, headers, err := provider.Get(context.Background(), &options.ProcessingOptions{})
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

func (s *ImageProviderTestSuite) TestNewFromURLNotFound() {
	s.status = http.StatusNotFound
	s.data = []byte("Not Found")
	s.header.Set("Content-Type", "text/plain")

	provider, err := s.factory.NewMemoryURL(context.Background(), s.server.URL)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *ImageProviderTestSuite) TestNewFromURLInvalidImage() {
	s.data = []byte("not an image")

	provider, err := s.factory.NewMemoryURL(context.Background(), s.server.URL)
	s.Require().Error(err)
	s.Require().Nil(provider)
}

func (s *ImageProviderTestSuite) TestNewFromTriple() {
	// Test with base64 (should prefer base64)
	provider, err := s.factory.NewMemoryTriple(s.testDataB64, "../testdata/test1.jpg", s.server.URL)
	s.Require().NoError(err)
	s.Require().NotNil(provider)
	s.Equal(s.testData, s.readImageData(provider))

	// Test with file path (no base64)
	provider, err = s.factory.NewMemoryTriple("", "../testdata/test1.jpg", s.server.URL)
	s.Require().NoError(err)
	s.Require().NotNil(provider)
	s.Equal(s.testData, s.readImageData(provider))

	// Test with URL (no base64 or file)
	provider, err = s.factory.NewMemoryTriple("", "", s.server.URL)
	s.Require().NoError(err)
	s.Require().NotNil(provider)
	s.Equal(s.testData, s.readImageData(provider))

	// Test with no inputs
	provider, err = s.factory.NewMemoryTriple("", "", "")
	s.Require().NoError(err)
	s.Nil(provider)
}

func TestImageProvider(t *testing.T) {
	suite.Run(t, new(ImageProviderTestSuite))
}
