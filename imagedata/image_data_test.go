package imagedata

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type ImageDataTestSuite struct {
	testutil.LazySuite

	fetcherCfg testutil.LazyObj[*fetcher.Config]
	factory    testutil.LazyObj[*Factory]
	testServer testutil.LazyTestServer

	data []byte
}

func (s *ImageDataTestSuite) SetupSuite() {
	s.data = testutil.NewTestDataProvider(s.T).Read("test1.jpg")

	s.fetcherCfg, _ = testutil.NewLazySuiteObj(
		s,
		func() (*fetcher.Config, error) {
			c := fetcher.NewDefaultConfig()
			c.Transport.HTTP.AllowLoopbackSourceAddresses = true
			c.Transport.HTTP.ClientKeepAliveTimeout = 0

			return &c, nil
		},
	)

	s.factory, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Factory, error) {
			fetcher, err := fetcher.New(s.fetcherCfg())
			if err != nil {
				return nil, err
			}

			return NewFactory(fetcher, nil), nil
		},
	)

	s.testServer, _ = testutil.NewLazySuiteTestServer(
		s,
		func(srv *testutil.TestServer) error {
			// Default headers and body for 200 OK response
			srv.SetHeaders(
				httpheaders.ContentType, "image/jpeg",
				httpheaders.ContentLength, strconv.Itoa(len(s.data)),
			).SetBody(s.data)

			return nil
		},
	)
}

func (s *ImageDataTestSuite) SetupSubTest() {
	// We use t.Run() a lot, so we need to reset lazy objects at the beginning of each subtest
	s.ResetLazyObjects()
}

func (s *ImageDataTestSuite) TestDownloadStatusOK() {
	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.data), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func (s *ImageDataTestSuite) TestDownloadStatusPartialContent() {
	testCases := []struct {
		name         string
		contentRange string
		expectErr    bool
	}{
		{
			name:         "Full Content-Range",
			contentRange: fmt.Sprintf("bytes 0-%d/%d", len(s.data)-1, len(s.data)),
			expectErr:    false,
		},
		{
			name:         "Partial Content-Range, early end",
			contentRange: fmt.Sprintf("bytes 0-%d/%d", len(s.data)-2, len(s.data)),
			expectErr:    true,
		},
		{
			name:         "Partial Content-Range, late start",
			contentRange: fmt.Sprintf("bytes 1-%d/%d", len(s.data)-1, len(s.data)),
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
			contentRange: fmt.Sprintf("bytes */%d", len(s.data)),
			expectErr:    true,
		},
		{
			name:         "Unknown Content-Range size, full range",
			contentRange: fmt.Sprintf("bytes 0-%d/*", len(s.data)-1),
			expectErr:    false,
		},
		{
			name:         "Unknown Content-Range size, early end",
			contentRange: fmt.Sprintf("bytes 0-%d/*", len(s.data)-2),
			expectErr:    true,
		},
		{
			name:         "Unknown Content-Range size, late start",
			contentRange: fmt.Sprintf("bytes 1-%d/*", len(s.data)-1),
			expectErr:    true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.testServer().
				SetHeaders(httpheaders.ContentRange, tc.contentRange).
				SetStatusCode(http.StatusPartialContent)

			imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Equal(http.StatusNotFound, ierrors.Wrap(err, 0).StatusCode())
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(imgdata)
				s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.data), imgdata.Reader()))
				s.Require().Equal(imagetype.JPEG, imgdata.Format())
			}
		})
	}
}

func (s *ImageDataTestSuite) TestDownloadStatusNotFound() {
	s.testServer().
		SetStatusCode(http.StatusNotFound).
		SetBody([]byte("Not Found")).
		SetHeaders(httpheaders.ContentType, "text/plain")

	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().Error(err)
	s.Require().Equal(404, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadStatusForbidden() {
	s.testServer().
		SetStatusCode(http.StatusForbidden).
		SetBody([]byte("Forbidden")).
		SetHeaders(httpheaders.ContentType, "text/plain")

	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().Error(err)
	s.Require().Equal(403, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadStatusInternalServerError() {
	s.testServer().
		SetStatusCode(http.StatusInternalServerError).
		SetBody([]byte("Internal Server Error")).
		SetHeaders(httpheaders.ContentType, "text/plain")

	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().Error(err)
	s.Require().Equal(502, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadUnreachable() {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	l.Close()

	serverURL := fmt.Sprintf("http://%s", l.Addr().String())

	imgdata, _, err := s.factory().DownloadSync(context.Background(), serverURL, "Test image", DownloadOptions{})

	s.Require().Error(err)
	s.Require().Equal(500, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadInvalidImage() {
	s.testServer().SetBody([]byte("invalid"))

	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().Error(err)
	s.Require().Equal(http.StatusUnprocessableEntity, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadSourceAddressNotAllowed() {
	s.fetcherCfg().Transport.HTTP.AllowLoopbackSourceAddresses = false

	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().Error(err)
	s.Require().Equal(404, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadImageFileTooLarge() {
	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{
		MaxSrcFileSize: 1,
	})

	s.Require().Error(err)
	s.Require().Equal(422, ierrors.Wrap(err, 0).StatusCode())
	s.Require().Nil(imgdata)
}

func (s *ImageDataTestSuite) TestDownloadGzip() {
	buf := new(bytes.Buffer)

	enc := gzip.NewWriter(buf)
	_, err := enc.Write(s.data)
	s.Require().NoError(err)
	err = enc.Close()
	s.Require().NoError(err)

	s.testServer().
		SetBody(buf.Bytes()).
		SetHeaders(
			httpheaders.ContentEncoding, "gzip",
			httpheaders.ContentLength, strconv.Itoa(buf.Len()), // Update Content-Length
		)

	imgdata, _, err := s.factory().DownloadSync(context.Background(), s.testServer().URL(), "Test image", DownloadOptions{})

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.data), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func (s *ImageDataTestSuite) TestFromFile() {
	imgdata, err := s.factory().NewFromPath("../testdata/test1.jpg")

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.data), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func (s *ImageDataTestSuite) TestFromBase64() {
	b64 := base64.StdEncoding.EncodeToString(s.data)

	imgdata, err := s.factory().NewFromBase64(b64)

	s.Require().NoError(err)
	s.Require().NotNil(imgdata)
	s.Require().True(testutil.ReadersEqual(s.T(), bytes.NewReader(s.data), imgdata.Reader()))
	s.Require().Equal(imagetype.JPEG, imgdata.Format())
}

func TestImageData(t *testing.T) {
	suite.Run(t, new(ImageDataTestSuite))
}
