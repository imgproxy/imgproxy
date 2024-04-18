package etag

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

var (
	po = options.NewProcessingOptions()

	imgWithETag = imagedata.ImageData{
		Data:    []byte("Hello Test"),
		Headers: map[string]string{"ETag": `"loremipsumdolor"`},
	}
	imgWithoutETag = imagedata.ImageData{
		Data: []byte("Hello Test"),
	}

	etagReq  = `"yj0WO6sFU4GCciYUBWjzvvfqrBh869doeOC2Pp5EI1Y/RImxvcmVtaXBzdW1kb2xvciI"`
	etagData = `"yj0WO6sFU4GCciYUBWjzvvfqrBh869doeOC2Pp5EI1Y/DvyChhMNu_sFX7jrjoyrgQbnFwfoOVv7kzp_Fbs6hQBg"`
)

type EtagTestSuite struct {
	suite.Suite

	h Handler
}

func (s *EtagTestSuite) SetupSuite() {
	logrus.SetOutput(io.Discard)
}

func (s *EtagTestSuite) TeardownSuite() {
	logrus.SetOutput(os.Stdout)
}

func (s *EtagTestSuite) SetupTest() {
	s.h = Handler{}
	config.Reset()
}

func (s *EtagTestSuite) TestGenerateActualReq() {
	s.h.SetActualProcessingOptions(po)
	s.h.SetActualImageData(&imgWithETag)

	s.Require().Equal(etagReq, s.h.GenerateActualETag())
}

func (s *EtagTestSuite) TestGenerateActualData() {
	s.h.SetActualProcessingOptions(po)
	s.h.SetActualImageData(&imgWithoutETag)

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

	s.Require().True(s.h.SetActualProcessingOptions(po))
	s.Require().True(s.h.ProcessingOptionsMatch())
}

func (s *EtagTestSuite) TestProcessingOptionsCheckFailure() {
	i := strings.Index(etagReq, "/")
	wrongEtag := `"wrongpohash` + etagReq[i:]

	s.h.ParseExpectedETag(wrongEtag)

	s.Require().False(s.h.SetActualProcessingOptions(po))
	s.Require().False(s.h.ProcessingOptionsMatch())
}

func (s *EtagTestSuite) TestImageETagExpectedPresent() {
	s.h.ParseExpectedETag(etagReq)

	//nolint:testifylint // False-positive expected-actual
	s.Require().Equal(imgWithETag.Headers["ETag"], s.h.ImageEtagExpected())
}

func (s *EtagTestSuite) TestImageETagExpectedBlank() {
	s.h.ParseExpectedETag(etagData)

	s.Require().Empty(s.h.ImageEtagExpected())
}

func (s *EtagTestSuite) TestImageDataCheckDataToDataSuccess() {
	s.h.ParseExpectedETag(etagData)
	s.Require().True(s.h.SetActualImageData(&imgWithoutETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToDataFailure() {
	i := strings.Index(etagData, "/")
	wrongEtag := etagData[:i] + `/Dwrongimghash"`

	s.h.ParseExpectedETag(wrongEtag)
	s.Require().False(s.h.SetActualImageData(&imgWithoutETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToReqSuccess() {
	s.h.ParseExpectedETag(etagData)
	s.Require().True(s.h.SetActualImageData(&imgWithETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToReqFailure() {
	i := strings.Index(etagData, "/")
	wrongEtag := etagData[:i] + `/Dwrongimghash"`

	s.h.ParseExpectedETag(wrongEtag)
	s.Require().False(s.h.SetActualImageData(&imgWithETag))
}

func (s *EtagTestSuite) TestImageDataCheckReqToDataFailure() {
	s.h.ParseExpectedETag(etagReq)
	s.Require().False(s.h.SetActualImageData(&imgWithoutETag))
}

func (s *EtagTestSuite) TestETagBusterFailure() {
	config.ETagBuster = "busted"

	s.h.ParseExpectedETag(etagReq)
	s.Require().False(s.h.SetActualImageData(&imgWithoutETag))
}

func TestEtag(t *testing.T) {
	suite.Run(t, new(EtagTestSuite))
}
