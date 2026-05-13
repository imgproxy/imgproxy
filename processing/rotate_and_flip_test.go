package processing_test

import (
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/stretchr/testify/suite"
)

type RotateAndFlipTestSuite struct {
	testSuite

	imgs []string
}

type rotateAndFlipTestCase struct {
	sourceFile string
	rotate     int
	flipH      bool
	flipV      bool
	autoRotate bool
}

func (c rotateAndFlipTestCase) ImagePath() string {
	return c.sourceFile
}

func (r rotateAndFlipTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	if r.rotate != 0 {
		opts.Add("rotate").Set(0, r.rotate)
	}

	if r.flipH {
		opts.Add("flip").Set(0, 1)
	}
	if r.flipV {
		opts.Add("flip").Set(1, 1)
	}

	if !r.autoRotate {
		opts.Add("auto_rotate").Set(0, 0)
	}

	return opts.String()
}

func (s *RotateAndFlipTestSuite) processImg(
	imgIndex int,
	opts rotateAndFlipTestCase,
) *testutil.ImageHash {
	opts.sourceFile = s.imgs[imgIndex]

	resultData := s.processImage(opts)
	defer resultData.Close()

	hash, err := testutil.NewImageHash(resultData.Reader(), testutil.HashTypeSHA256)
	s.Require().NoError(err)

	key := fmt.Sprintf(
		"img-%d_rotate-%v_flip-%v_%v_ar-%v",
		imgIndex,
		opts.rotate,
		opts.flipH,
		opts.flipV,
		opts.autoRotate,
	)

	s.ImageMatcher().ImageMatches(s.T(), resultData.Reader(), key, 0)

	return hash
}

func (s *RotateAndFlipTestSuite) collectRotationsFlips(imgIndex int, autoRotate bool) []*testutil.ImageHash {
	rotates := []int{0, 90, 180, 270}
	flips := []bool{false, true}

	hashes := make([]*testutil.ImageHash, 0, len(rotates)*len(flips)*len(flips))

	for _, rotate := range rotates {
		for _, flipH := range flips {
			for _, flipV := range flips {
				opts := rotateAndFlipTestCase{
					sourceFile: s.imgs[imgIndex],
					rotate:     rotate,
					flipH:      flipH,
					flipV:      flipV,
					autoRotate: autoRotate,
				}

				hashes = append(hashes, s.processImg(imgIndex, opts))
			}
		}
	}

	return hashes
}

func (s *RotateAndFlipTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	for i := range 8 {
		s.imgs = append(s.imgs, fmt.Sprintf("orientation-%d.png", i))
	}
}

func (s *RotateAndFlipTestSuite) TestOrientationAutoRotate() {
	opts := rotateAndFlipTestCase{autoRotate: false}

	// Test with auto_rotate:false - all outputs should be the same
	hashes := make([]*testutil.ImageHash, 0, len(s.imgs))
	for i := range s.imgs {
		hashes = append(hashes, s.processImg(i, opts))
	}

	for i := 1; i < len(hashes); i++ {
		d, err := hashes[0].Distance(hashes[i])

		s.Require().NoError(err)
		s.Require().Zero(d)
	}

	// Test with auto_rotate:true - each subsequent output should differ from the previous
	opts.autoRotate = true

	hashesAr := make([]*testutil.ImageHash, 0, len(s.imgs))
	for i := range s.imgs {
		hashesAr = append(hashesAr, s.processImg(i, opts))
	}

	for i := 1; i < len(hashesAr); i++ {
		d, err := hashesAr[i-1].Distance(hashesAr[i])
		s.Require().NoError(err)
		s.Require().NotZero(d)
	}
}

func (s *RotateAndFlipTestSuite) TestRotateFlip() {
	hashes := make([][]*testutil.ImageHash, 0, len(s.imgs))

	for i := range s.imgs {
		hashes = append(hashes, s.collectRotationsFlips(i, false))
	}

	// Ensure all hashes of img[0] and other imgs are equal
	for n := 1; n < len(hashes); n++ {
		for i := range hashes[0] {
			d, err := hashes[0][i].Distance(hashes[n][i])
			s.Require().NoError(err)
			s.Require().Zero(d)
		}
	}

	// Ensure that the next hash is different from the previous one for all imgs
	for n := range hashes {
		for i := range len(hashes[n]) - 1 {
			d, err := hashes[n][i].Distance(hashes[n][i+1])
			s.Require().NoError(err)
			s.Require().NotZero(d)
		}
	}
}

func (s *RotateAndFlipTestSuite) TestRotateFlipAutoRotate() {
	hashes := make([][]*testutil.ImageHash, 0, len(s.imgs))

	for i := range s.imgs {
		hashes = append(hashes, s.collectRotationsFlips(i, true))
	}

	// Ensure all hashes of img[0] and other imgs are NOT equal
	for n := 1; n < len(hashes); n++ {
		for i := range hashes[0] {
			d, err := hashes[0][i].Distance(hashes[n][i])
			s.Require().NoError(err)
			s.Require().NotZero(d)
		}
	}

	// Ensure that the next hash is different from the previous one for all imgs
	for n := range hashes {
		for i := range len(hashes[n]) - 1 {
			d, err := hashes[n][i].Distance(hashes[n][i+1])
			s.Require().NoError(err)
			s.Require().NotZero(d)
		}
	}
}

func TestRotateAndFlip(t *testing.T) {
	suite.Run(t, new(RotateAndFlipTestSuite))
}
