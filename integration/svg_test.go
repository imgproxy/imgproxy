package integration

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

const (
	// svgoTestSuiteURL represents the SVG test suite archive URL
	// Source: https://svg.github.io/svgo-test-suite/svgo-test-suite.tar.gz
	svgoTestSuiteURL = "https://svg.github.io/svgo-test-suite/svgo-test-suite.tar.gz"
)

var (
	ignore = []string{
		// vips: corrupted file
		"W3C_SVG_11_TestSuite/svg/struct-frag-01-t.svg", // bad dimensions
		"W3C_SVG_11_TestSuite/svg/struct-frag-04-t.svg", // failed to load

		// glib: encoding 'UTF-16' doesn't match auto-detected 'UTF-8'
		"oxygen-icons-5.116.0/scalable/apps/kalarm.svg",
	}
)

// SvgTestSuite is a test suite for testing SVG processing
type SvgTestSuite struct {
	Suite

	imagesDir string
}

func (s *SvgTestSuite) SetupTest() {
	path := os.Getenv("TEST_SVG_PATH")
	if path == "" {
		s.T().Skipf(
			"Use TEST_SVG_PATH=/path/to/svgs to run SVG test suite. Download and extract %s to that folder. Skipping.",
			svgoTestSuiteURL,
		)
	}

	var err error

	s.imagesDir, err = filepath.Abs(path)
	s.Require().NoError(err)

	if _, err := os.Stat(s.imagesDir); err != nil {
		s.T().Skipf(
			"SVG test suite directory `%s` is empty or does not exist. Download and extract %s somewhere. Skipping.",
			s.imagesDir,
			svgoTestSuiteURL,
		)

		return
	}

	s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = true
	s.Config().Fetcher.Transport.Local.Root = s.imagesDir

	// Enable SVG sanitization
	s.Config().Processing.Svg.Sanitize = true
}

func (s *SvgTestSuite) SetupSubTest() {
	// We use t.Run() a lot, so we need to reset lazy objects at the beginning of each subtest
	s.ResetLazyObjects()
}

// TestSvgSuite tests the SVG test suite with minification and sanitization enabled
func (s *SvgTestSuite) TestSvgSuite() {
	// Walk through all SVG files in the directory
	err := filepath.Walk(s.imagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-SVG files and directories
		if info.IsDir() || filepath.Base(path) == ".DS_Store" {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".svg" {
			return nil
		}

		relPath, _ := filepath.Rel(s.imagesDir, path)
		if slices.Contains(ignore, relPath) {
			s.T().Logf("Ignoring SVG: %s\n", relPath)
			return nil
		}

		s.T().Logf("Processing SVG: %s\n", relPath)

		// Calculate hash of original SVG
		sourceHash, err := testutil.NewImageHashFromPath(path, testutil.HashTypeSHA256)
		s.Require().NoError(err)

		// Construct the source URL for imgproxy
		sourceUrl := fmt.Sprintf("/insecure/plain/local:///%s", relPath)

		// Read processed SVG from imgproxy
		resp := s.GET(sourceUrl)
		defer resp.Body.Close()

		s.Require().Equal(
			http.StatusOK,
			resp.StatusCode,
			"expected status code 200 OK, got %d, url: %s",
			resp.StatusCode,
			sourceUrl,
		)

		// Calculate hash of processed SVG
		processedHash, err := testutil.NewImageHashFromReader(resp.Body, testutil.HashTypeSHA256)
		s.Require().NoError(err)

		// Compare hashes
		distance, err := sourceHash.Distance(processedHash)
		s.Require().NoError(err)

		s.Require().Equal(0, distance, "SVG hashes are too different for %s: distance %d", relPath, distance)

		return nil
	})

	s.Require().NoError(err, "Failed to process SVG test suite")
}

func TestSvgSuite(t *testing.T) {
	suite.Run(t, new(SvgTestSuite))
}
