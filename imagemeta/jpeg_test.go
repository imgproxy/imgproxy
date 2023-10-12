package imagemeta

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JpegTestSuite struct {
	suite.Suite
}

func (s *JpegTestSuite) openFile(name string) *os.File {
	wd, err := os.Getwd()
	require.Nil(s.T(), err)
	path := filepath.Join(wd, "..", "testdata", name)
	f, err := os.Open(path)
	require.Nil(s.T(), err)
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
			require.Nil(s.T(), err)
			require.Equal(s.T(), expectedMeta, metadata)
		}()
	}
}

func TestJpeg(t *testing.T) {
	suite.Run(t, new(JpegTestSuite))
}
