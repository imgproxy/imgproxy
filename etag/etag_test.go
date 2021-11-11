package etag

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

	etagReq  string
	etagData string
)

type EtagTestSuite struct {
	suite.Suite

	h Handler
}

func (s *EtagTestSuite) SetupSuite() {
	logrus.SetOutput(ioutil.Discard)

	s.h.SetActualProcessingOptions(po)
	s.h.SetActualImageData(&imgWithETag)
	etagReq = s.h.GenerateActualETag()

	s.h.SetActualImageData(&imgWithoutETag)
	etagData = s.h.GenerateActualETag()
}

func (s *EtagTestSuite) TeardownSuite() {
	logrus.SetOutput(os.Stdout)
}

func (s *EtagTestSuite) SetupTest() {
	s.h = Handler{}
}

func (s *EtagTestSuite) TestGenerateActualReq() {
	s.h.SetActualProcessingOptions(po)
	s.h.SetActualImageData(&imgWithETag)

	assert.Equal(s.T(), etagReq, s.h.GenerateActualETag())
}

func (s *EtagTestSuite) TestGenerateActualData() {
	s.h.SetActualProcessingOptions(po)
	s.h.SetActualImageData(&imgWithoutETag)

	assert.Equal(s.T(), etagData, s.h.GenerateActualETag())
}

func (s *EtagTestSuite) TestGenerateExpectedReq() {
	s.h.ParseExpectedETag(etagReq)
	assert.Equal(s.T(), etagReq, s.h.GenerateExpectedETag())
}

func (s *EtagTestSuite) TestGenerateExpectedData() {
	s.h.ParseExpectedETag(etagData)
	assert.Equal(s.T(), etagData, s.h.GenerateExpectedETag())
}

func (s *EtagTestSuite) TestProcessingOptionsCheckSuccess() {
	s.h.ParseExpectedETag(etagReq)

	assert.True(s.T(), s.h.SetActualProcessingOptions(po))
	assert.True(s.T(), s.h.ProcessingOptionsMatch())
}

func (s *EtagTestSuite) TestProcessingOptionsCheckFailure() {
	i := strings.Index(etagReq, "/")
	wrongEtag := `"wrongpohash` + etagReq[i:]

	s.h.ParseExpectedETag(wrongEtag)

	assert.False(s.T(), s.h.SetActualProcessingOptions(po))
	assert.False(s.T(), s.h.ProcessingOptionsMatch())
}

func (s *EtagTestSuite) TestImageETagExpectedPresent() {
	s.h.ParseExpectedETag(etagReq)

	assert.Equal(s.T(), imgWithETag.Headers["ETag"], s.h.ImageEtagExpected())
}

func (s *EtagTestSuite) TestImageETagExpectedBlank() {
	s.h.ParseExpectedETag(etagData)

	assert.Empty(s.T(), s.h.ImageEtagExpected())
}

func (s *EtagTestSuite) TestImageDataCheckDataToDataSuccess() {
	s.h.ParseExpectedETag(etagData)
	assert.True(s.T(), s.h.SetActualImageData(&imgWithoutETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToDataFailure() {
	i := strings.Index(etagData, "/")
	wrongEtag := etagData[:i] + `/Dwrongimghash"`

	s.h.ParseExpectedETag(wrongEtag)
	assert.False(s.T(), s.h.SetActualImageData(&imgWithoutETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToReqSuccess() {
	s.h.ParseExpectedETag(etagData)
	assert.True(s.T(), s.h.SetActualImageData(&imgWithETag))
}

func (s *EtagTestSuite) TestImageDataCheckDataToReqFailure() {
	i := strings.Index(etagData, "/")
	wrongEtag := etagData[:i] + `/Dwrongimghash"`

	s.h.ParseExpectedETag(wrongEtag)
	assert.False(s.T(), s.h.SetActualImageData(&imgWithETag))
}

func (s *EtagTestSuite) TestImageDataCheckReqToDataFailure() {
	s.h.ParseExpectedETag(etagReq)
	assert.False(s.T(), s.h.SetActualImageData(&imgWithoutETag))
}

func TestEtag(t *testing.T) {
	suite.Run(t, new(EtagTestSuite))
}
