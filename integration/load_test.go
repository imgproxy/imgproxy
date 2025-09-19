package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
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

	testImagesPath string
}

func (s *LoadTestSuite) SetupTest() {
	s.testImagesPath = s.TestData.Path("test-images")

	config.MaxAnimationFrames = 999
	config.DevelopmentErrorsMode = true

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
		basePath := filepath.Base(path)

		// Replace the extension with .png
		referencePath := strings.TrimSuffix(basePath, filepath.Ext(basePath)) + ".png"

		// Construct the full path to the reference image (integration/ folder)
		referencePath = filepath.Join(s.testImagesPath, "integration", folder, referencePath)

		// Construct the source URL for imgproxy (no processing)
		sourceUrl := fmt.Sprintf("/insecure/plain/local:///%s/%s@png", folder, basePath)

		imgproxyImageData := s.fetchImage(sourceUrl)
		var imgproxyImage vips.Image
		s.Require().NoError(imgproxyImage.Load(imgproxyImageData, 1, 1.0, 1))

		hash1, err := testutil.ImageHash(unsafe.Pointer(imgproxyImage.VipsImage))
		s.Require().NoError(err)

		referenceFile, err := os.Open(referencePath)
		s.Require().NoError(err)
		defer referenceFile.Close()

		referenceImageData, err := s.Imgproxy().ImageDataFactory().NewFromPath(referencePath)
		s.Require().NoError(err)

		var referenceImage vips.Image
		s.Require().NoError(referenceImage.Load(referenceImageData, 1, 1.0, 1))

		hash2, err := testutil.ImageHash(unsafe.Pointer(referenceImage.VipsImage))
		s.Require().NoError(err)

		distance, err := hash1.Distance(hash2)
		s.Require().NoError(err)

		imgproxyImageData.Close()
		referenceImageData.Close()
		imgproxyImage.Clear()
		referenceImage.Clear()

		s.Require().LessOrEqual(distance, similarityThreshold,
			"Image %s differs from reference image %s by %d, which is greater than the allowed threshold of %d",
			basePath, referencePath, distance, similarityThreshold)

		return nil
	})

	s.Require().NoError(err)
}

// fetchImage fetches an image from the imgproxy server
func (s *LoadTestSuite) fetchImage(path string) imagedata.ImageData {
	resp := s.GET(path)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "Expected status code 200 OK, got %d, path: %s", resp.StatusCode, path)

	bytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "Failed to read response body from %s", path)

	d, err := s.Imgproxy().ImageDataFactory().NewFromBytes(bytes)
	s.Require().NoError(err, "Failed to load image from bytes for %s", path)

	return d
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
