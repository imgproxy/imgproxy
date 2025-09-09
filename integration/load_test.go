package integration

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/corona10/goimagehash"
	"github.com/imgproxy/imgproxy/v3"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

const (
	similarityThreshold = 5 // Distance between images to be considered similar
)

type LoadTestSuite struct {
	Suite
	testData       *testutil.TestDataProvider
	testImagesPath string

	server *TestServer
}

// SetupSuite starts imgproxy instance server
func (s *LoadTestSuite) SetupSuite() {
	s.testData = testutil.NewTestDataProvider(s.T())
	s.testImagesPath = s.testData.Path("test-images")

	c, err := imgproxy.LoadConfigFromEnv(nil)
	s.Require().NoError(err)

	c.Fetcher.Transport.Local.Root = s.testImagesPath
	config.MaxAnimationFrames = 999
	config.DevelopmentErrorsMode = true

	// In this test we start the single imgproxy server for all test cases
	s.server = s.StartImgproxy(c)
}

// TearDownSuite stops imgproxy instance server
func (s *LoadTestSuite) TearDownSuite() {
	s.server.Shutdown()
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
		basePath := filepath.Base(path)

		// Replace the extension with .png
		referencePath := strings.TrimSuffix(basePath, filepath.Ext(basePath)) + ".png"

		// Construct the full path to the reference image (integration/ folder)
		referencePath = filepath.Join(s.testImagesPath, "integration", folder, referencePath)

		// Construct the source URL for imgproxy (no processing)
		sourceUrl := fmt.Sprintf("insecure/plain/local:///%s/%s@png", folder, basePath)

		imgproxyImageBytes := s.fetchImage(sourceUrl)
		imgproxyImage, err := png.Decode(bytes.NewReader(imgproxyImageBytes))
		s.Require().NoError(err, "Failed to decode PNG image from imgproxy for %s", basePath)

		referenceFile, err := os.Open(referencePath)
		s.Require().NoError(err)
		defer referenceFile.Close()

		referenceImage, err := png.Decode(referenceFile)
		s.Require().NoError(err, "Failed to decode PNG reference image for %s", referencePath)

		hash1, err := goimagehash.DifferenceHash(imgproxyImage)
		s.Require().NoError(err)

		hash2, err := goimagehash.DifferenceHash(referenceImage)
		s.Require().NoError(err)

		distance, err := hash1.Distance(hash2)
		s.Require().NoError(err)

		s.Require().LessOrEqual(distance, similarityThreshold,
			"Image %s differs from reference image %s by %d, which is greater than the allowed threshold of %d",
			basePath, referencePath, distance, similarityThreshold)

		return nil
	})

	s.Require().NoError(err)
}

// fetchImage fetches an image from the imgproxy server
func (s *LoadTestSuite) fetchImage(path string) []byte {
	url := fmt.Sprintf("http://%s/%s", s.server.Addr, path)

	resp, err := http.Get(url)
	s.Require().NoError(err, "Failed to fetch image from %s", url)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "Expected status code 200 OK, got %d, url: %s", resp.StatusCode, url)

	bytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "Failed to read response body from %s", url)

	return bytes
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
