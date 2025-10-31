package integration

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

const (
	// maximum Hamming distance between the source and destination hash
	// NOTE: 2 here is because HEIC works a bit differently over the platforms, investigate
	maxDistance = 2
)

// formats to test
var formats = []imagetype.Type{
	imagetype.GIF,
	imagetype.JPEG,
	imagetype.HEIC,
	imagetype.JXL,
	imagetype.SVG,
	imagetype.TIFF,
	imagetype.WEBP,
	imagetype.BMP,
	imagetype.ICO,
}

type MatrixTestSuite struct {
	Suite

	matcher        *testutil.ImageHashMatcher
	testImagesPath string
}

func (s *MatrixTestSuite) SetupTest() {
	s.testImagesPath = s.TestData.Path("test-images")
	s.matcher = testutil.NewImageHashMatcher(s.TestData)

	s.Config().Security.MaxAnimationFrames = 999
	s.Config().Server.DevelopmentErrorsMode = true
	s.Config().Fetcher.Transport.Local.Root = s.testImagesPath
}

// testLoadFolder fetches images iterates over images in the specified folder,
// runs imgproxy on each image, and compares the result with the reference image
// which is expected to be in the `integration` folder with the same name
// but with `.png` extension.
func (s *MatrixTestSuite) testFormat(source, target imagetype.Type) {
	folder := source.String()

	// TODO: rename the folders in test-images repo
	if folder == "heic" {
		folder = "heif"
	}

	if folder == "jpeg" {
		folder = "jpg"
	}

	walkPath := path.Join(s.testImagesPath, folder)

	// Iterate over the files in the source folder
	err := filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {
		s.Require().NoError(err)

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// get the base name of the file (8-bpp.png)
		baseName := filepath.Base(path)

		// Construct the source URL for imgproxy (no processing)
		sourceUrl := fmt.Sprintf("/insecure/plain/local:///%s/%s@%s", folder, baseName, target.String())

		// Read source image from imgproxy
		resp := s.GET(sourceUrl)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK, got %d, url: %s", resp.StatusCode, sourceUrl)

		// Match image to precalculated hash
		s.matcher.ImageMatches(s.T(), resp.Body, baseName, maxDistance)

		return nil
	})

	s.Require().NoError(err)
}

func (s *MatrixTestSuite) TestMatrix() {
	for _, source := range formats {
		for _, target := range formats {
			s.Run(fmt.Sprintf("%s/%s", source.String(), target.String()), func() {
				if !source.IsVector() && target.IsVector() {
					// we can not vectorize a raster image
					s.T().Logf("Skipping %s -> %s conversion: we can not vectorize raster image", source.String(), target.String())
					return
				}

				if !vips.SupportsLoad(source) {
					s.T().Logf("Skipping %s -> %s conversion: source format not supported by VIPS", source.String(), target.String())
					return
				}

				if !vips.SupportsSave(target) {
					s.T().Logf("Skipping %s -> %s conversion: target format not supported by VIPS", source.String(), target.String())
					return
				}

				s.testFormat(source, target)
			})
		}
	}
}

func TestMatrix(t *testing.T) {
	suite.Run(t, new(MatrixTestSuite))
}
