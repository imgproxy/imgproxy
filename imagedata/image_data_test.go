package imagedata

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type ImageDataTestSuite struct {
	suite.Suite

	server *httptest.Server

	status int
	data   []byte
	header http.Header
	check  func(*http.Request)

	defaultData []byte
}

func (s *ImageDataTestSuite) SetupSuite() {
	config.Reset()
	config.ClientKeepAliveTimeout = 0

	Init()

	f, err := os.Open("../testdata/test1.jpg")
	s.Require().NoError(err)
	defer f.Close()

	data, err := io.ReadAll(f)
	s.Require().NoError(err)

	s.defaultData = data

	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if s.check != nil {
			s.check(r)
		}

		for k, vv := range s.header {
			for _, v := range vv {
				rw.Header().Add(k, v)
			}
		}

		data := s.data
		if data == nil {
			data = s.defaultData
		}

		rw.Header().Set("Content-Length", strconv.Itoa(len(data)))

		rw.WriteHeader(s.status)
		rw.Write(data)
	}))
}

func (s *ImageDataTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *ImageDataTestSuite) SetupTest() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true

	s.status = http.StatusOK
	s.data = nil
	s.check = nil

	s.header = http.Header{}
	s.header.Set("Content-Type", "image/jpeg")
}

func (s *ImageDataTestSuite) TestDownloadStatusOK() {
	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.defaultData), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func (s *ImageDataTestSuite) TestDownloadStatusPartialContent() {
	s.status = http.StatusPartialContent

	testCases := []struct {
		name         string
		contentRange string
		expectErr    bool
	}{
		{
			name:         "Full Content-Range",
			contentRange: fmt.Sprintf("bytes 0-%d/%d", len(s.defaultData)-1, len(s.defaultData)),
			expectErr:    false,
		},
		{
			name:         "Partial Content-Range, early end",
			contentRange: fmt.Sprintf("bytes 0-%d/%d", len(s.defaultData)-2, len(s.defaultData)),
			expectErr:    true,
		},
		{
			name:         "Partial Content-Range, late start",
			contentRange: fmt.Sprintf("bytes 1-%d/%d", len(s.defaultData)-1, len(s.defaultData)),
			expectErr:    true,
		},
		{
			name:         "Zero Content-Range",
			contentRange: "bytes 0-0/0",
			expectErr:    true,
		},
		{
			name:         "Invalid Content-Range",
			contentRange: "invalid",
			expectErr:    true,
		},
		{
			name:         "Unknown Content-Range range",
			contentRange: fmt.Sprintf("bytes */%d", len(s.defaultData)),
			expectErr:    true,
		},
		{
			name:         "Unknown Content-Range size, full range",
			contentRange: fmt.Sprintf("bytes 0-%d/*", len(s.defaultData)-1),
			expectErr:    false,
		},
		{
			name:         "Unknown Content-Range size, early end",
			contentRange: fmt.Sprintf("bytes 0-%d/*", len(s.defaultData)-2),
			expectErr:    true,
		},
		{
			name:         "Unknown Content-Range size, late start",
			contentRange: fmt.Sprintf("bytes 1-%d/*", len(s.defaultData)-1),
			expectErr:    true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.header.Set("Content-Range", tc.contentRange)

			imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Equal(404, ierrors.Wrap(err, 0).StatusCode())
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(imgdata)
				s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.defaultData), imgdata.Reader()))
				s.Require().Equal(imagetype.JPEG, imgdata.Format())
			}
		})
	}
}

func (s *ImageDataTestSuite) TestDownloadStatusNotFound() {
	s.status = http.StatusNotFound
	s.data = []byte("Not Found")
	s.header.Set("Content-Type", "text/plain")

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(404, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadStatusForbidden() {
	s.status = http.StatusForbidden
	s.data = []byte("Forbidden")
	s.header.Set("Content-Type", "text/plain")

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(404, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadStatusInternalServerError() {
	s.status = http.StatusInternalServerError
	s.data = []byte("Internal Server Error")
	s.header.Set("Content-Type", "text/plain")

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(500, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadUnreachable() {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	l.Close()

	serverURL := fmt.Sprintf("http://%s", l.Addr().String())

	imgdata, err := Download(context.Background(), serverURL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(500, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadInvalidImage() {
	s.data = []byte("invalid")

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(422, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadSourceAddressNotAllowed() {
	config.AllowLoopbackSourceAddresses = false

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(404, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadImageTooLarge() {
	config.MaxSrcResolution = 1

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(422, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadImageFileTooLarge() {
	config.MaxSrcFileSize = 1

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().Error(err)
	s.Require().Equal(422, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadGzip() {
	buf := new(bytes.Buffer)

	enc := gzip.NewWriter(buf)
	_, err := enc.Write(s.defaultData)
	s.Require().NoError(err)
	err = enc.Close()
	s.Require().NoError(err)

	s.data = buf.Bytes()
	s.header.Set("Content-Encoding", "gzip")

	imgdata, err := Download(context.Background(), s.server.URL, "Test image", DownloadOptions{}, security.DefaultOptions())

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.defaultData), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func (s *ImageDataTestSuite) TestFromFile() {
	imgdata, err := NewFromPath("../testdata/test1.jpg", security.DefaultOptions())

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.defaultData), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func (s *ImageDataTestSuite) TestFromBase64() {
	b64 := base64.StdEncoding.EncodeToString(s.defaultData)

	imgdata, err := NewFromBase64(b64, security.DefaultOptions())

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.defaultData), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func TestImageData(t *testing.T) {
	suite.Run(t, new(ImageDataTestSuite))
}
