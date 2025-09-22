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
	maxDistance = 0 // maximum image distance
)

type LoadTestSuite struct {
	Suite

	matcher           *testutil.ImageHashMatcher
	testImagesPath    string
	saveTmpImagesPath string
}

func (s *LoadTestSuite) SetupTest() {
	s.testImagesPath = s.TestData.Path("test-images")
	s.saveTmpImagesPath = os.Getenv("TEST_SAVE_TMP_IMAGES")
	s.matcher = testutil.NewImageHashMatcher(s.TestData)

	s.Config().Security.DefaultOptions.MaxAnimationFrames = 999
	s.Config().Server.DevelopmentErrorsMode = true
	s.Config().Fetcher.Transport.Local.Root = s.testImagesPath
}

// testLoadFolder fetches images iterates over images in the specified folder,
// runs imgproxy on each image, and compares the result with the reference image
// which is expected to be in the `integration` folder with the same name
// but with `.png` extension.
func (s *LoadTestSuite) testLoadFolder(folder string) {
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
		sourceUrl := fmt.Sprintf("/insecure/plain/local:///%s/%s@bmp", folder, baseName)

		// Read source image from imgproxy
		resp := s.GET(sourceUrl)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK, got %d, path: %s", resp.StatusCode, path)

		// Match image to precalculated hash
		s.matcher.ImageMatches(s.T(), resp.Body, baseName, maxDistance)

		return nil
	})

	s.Require().NoError(err)
}

// TestLoadSaveToPng ensures that our load pipeline works,
// including standard and custom loaders. For each source image
// in the folder, it does the passthrough request through imgproxy:
// no processing, just convert format of the source file to png.
// Then, it compares the result with the reference image.
func (s *LoadTestSuite) TestLoadSaveToPng() {
	testCases := []struct {
		name       string
		imageType  imagetype.Type
		folderName string
	}{
		{"GIF", imagetype.GIF, "gif"},
		{"JPEG", imagetype.JPEG, "jpg"},
		{"HEIC", imagetype.HEIC, "heif"},
		{"JXL", imagetype.JXL, "jxl"},
		{"SVG", imagetype.SVG, "svg"},
		{"TIFF", imagetype.TIFF, "tiff"},
		{"WEBP", imagetype.WEBP, "webp"},
		{"BMP", imagetype.BMP, "bmp"},
		{"ICO", imagetype.ICO, "ico"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if vips.SupportsLoad(tc.imageType) {
				s.testLoadFolder(tc.folderName)
			} else {
				t.Skipf("%s format not supported by VIPS", tc.name)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(LoadTestSuite))
}
