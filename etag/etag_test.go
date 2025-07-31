package etag

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

const (
	etagReq  = `"yj0WO6sFU4GCciYUBWjzvvfqrBh869doeOC2Pp5EI1Y/RImxvcmVtaXBzdW1kb2xvciI"`
	etagData = `"yj0WO6sFU4GCciYUBWjzvvfqrBh869doeOC2Pp5EI1Y/D3t8wWhX4piqDCV4ZMEZsKvOaIO6onhKjbf9f-ZfYUV0"`
)

type EtagTestSuite struct {
	suite.Suite

	po             *options.ProcessingOptions
	imgWithETag    *imagedata.ImageData
	imgWithoutETag *imagedata.ImageData

	h Handler
}

func (s *EtagTestSuite) SetupSuite() {
	logrus.SetOutput(io.Discard)
	s.po = options.NewProcessingOptions()

	d, err := os.ReadFile("../testdata/test1.jpg")
	s.Require().NoError(err)

	imgWithETag, err := imagedata.NewFromBytes(d, http.Header{"ETag": []string{`"loremipsumdolor"`}})
	s.Require().NoError(err)

	imgWithoutETag, err := imagedata.NewFromBytes(d, make(http.Header))
	s.Require().NoError(err)

	s.imgWithETag = imgWithETag
	s.imgWithoutETag = imgWithoutETag
}

func (s *EtagTestSuite) TeardownSuite() {
	logrus.SetOutput(os.Stdout)
}

func (s *EtagTestSuite) SetupTest() {
	s.h = Handler{}
	config.Reset()
}

func (s *EtagTestSuite) TestGenerateActualReq() {
	s.h.SetActualProcessingOptions(s.po)
	s.h.SetActualImageData(s.imgWithETag)

	s.Require().Equal(etagReq, s.h.GenerateActualETag())
}

func (s *EtagTestSuite) TestGenerateActualData() {
	s.h.SetActualProcessingOptions(s.po)
	s.h.SetActualImageData(s.imgWithoutETag)

	s.Require().Equal(etagData, s.h.GenerateActualETag())
}

func (s *EtagTestSuite) TestGenerateExpectedReq() {
	s.h.ParseExpectedETag(etagReq)
	s.Require().Equal(etagReq, s.h.GenerateExpectedETag())
}

func (s *EtagTestSuite) TestGenerateExpectedData() {
	s.h.ParseExpectedETag(etagData)
	s.Require().Equal(etagData, s.h.GenerateExpectedETag())
}

func (s *EtagTestSuite) TestProcessingOptionsCheckSuccess() {
	s.h.ParseExpectedETag(etagReq)

	s.Require().True(s.h.SetActualProcessingOptions(s.po))
	s.Require().True(s.h.ProcessingOptionsMatch())
}

func (s *EtagTestSuite) TestProcessingOptionsCheckFailure() {
	i := strings.Index(etagReq, "/")
	wrongEtag := `"wrongpohash` + etagReq[i:]

	s.h.ParseExpectedETag(wrongEtag)

	s.Require().False(s.h.SetActualProcessingOptions(s.po))
	s.Require().False(s.h.ProcessingOptionsMatch())
}

func (s *EtagTestSuite) TestImageETagExpectedPresent() {
	s.h.ParseExpectedETag(etagReq)

	//nolint:testifylint // False-positive expected-actual
	s.Require().Equal(s.imgWithETag.Headers["ETag"], s.h.ImageEtagExpected())
}

func (s *EtagTestSuite) TestImageETagExpectedBlank() {
	s.h.ParseExpectedETag(etagData)

	s.Require().Empty(s.h.ImageEtagExpected())
}

func (s *EtagTestSuite) TestImageDataCheckDataToDataSuccess() {
	s.h.ParseExpectedETag(etagData)
	s.Require().True(s.h.SetActualImageData(s.imgWithoutETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToDataFailure() {
	i := strings.Index(etagData, "/")
	wrongEtag := etagData[:i] + `/Dwrongimghash"`

	s.h.ParseExpectedETag(wrongEtag)
	s.Require().False(s.h.SetActualImageData(s.imgWithoutETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToReqSuccess() {
	s.h.ParseExpectedETag(etagData)
	s.Require().True(s.h.SetActualImageData(s.imgWithETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToReqFailure() {
	i := strings.Index(etagData, "/")
	wrongEtag := etagData[:i] + `/Dwrongimghash"`

	s.h.ParseExpectedETag(wrongEtag)
	s.Require().False(s.h.SetActualImageData(s.imgWithETag))
}

func (s *EtagTestSuite) TestImageDataCheckReqToDataFailure() {
	s.h.ParseExpectedETag(etagReq)
	s.Require().False(s.h.SetActualImageData(s.imgWithoutETag))
}

func (s *EtagTestSuite) TestETagBusterFailure() {
	config.ETagBuster = "busted"

	s.h.ParseExpectedETag(etagReq)
	s.Require().False(s.h.SetActualImageData(s.imgWithoutETag))
}

func TestEtag(t *testing.T) {
	suite.Run(t, new(EtagTestSuite))
}
