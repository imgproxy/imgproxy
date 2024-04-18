package imagemeta

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type JpegTestSuite struct {
	suite.Suite
}

func (s *JpegTestSuite) openFile(name string) *os.File {
	wd, err := os.Getwd()
	s.Require().NoError(err)
	path := filepath.Join(wd, "..", "testdata", name)
	f, err := os.Open(path)
	s.Require().NoError(err)
	return f
}

func (s *JpegTestSuite) TestDecodeJpegMeta() {
	files := []string{
		"test1.jpg",
		"test1.arith.jpg",
	}

	expectedMeta := &meta{
		format: imagetype.JPEG,
		width:  10,
		height: 10,
	}

	for _, file := range files {
		func() {
			f := s.openFile(file)
			defer f.Close()

			metadata, err := DecodeJpegMeta(f)
			s.Require().NoError(err)
			s.Require().Equal(expectedMeta, metadata)
		}()
	}
}

func TestJpeg(t *testing.T) {
	suite.Run(t, new(JpegTestSuite))
}
