package processing

import (
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type RotateAndFlipTestSuite struct {
	testSuite

	imgs []imagedata.ImageData
}

func (s *RotateAndFlipTestSuite) processImg(imgIndex int, o *options.Options) *testutil.ImageHash {
	result, err := s.Processor().ProcessImage(s.T().Context(), s.imgs[imgIndex], o)
	s.Require().NoError(err)
	defer result.OutData.Close()

	hash, err := testutil.NewImageHashFromReader(result.OutData.Reader(), testutil.HashTypeSHA256)
	s.Require().NoError(err)

	key := fmt.Sprintf(
		"img-%d_rotate-%v_flip-%v_%v_ar-%v",
		imgIndex,
		o.GetInt(keys.Rotate, 0),
		o.GetBool(keys.FlipHorizontal, false),
		o.GetBool(keys.FlipVertical, false),
		o.GetBool(keys.AutoRotate, false),
	)

	s.ImageMatcher.ImageMatches(s.T(), result.OutData.Reader(), key, 0)

	return hash
}

func (s *RotateAndFlipTestSuite) collectRotationsFlips(imgIndex int, o *options.Options) []*testutil.ImageHash {
	rotates := []int{0, 90, 180, 270}
	flips := []bool{false, true}

	var hashes []*testutil.ImageHash

	for _, rotate := range rotates {
		for _, flipH := range flips {
			for _, flipV := range flips {
				o.Set(keys.Rotate, rotate)

				if flipH {
					o.Set(keys.FlipHorizontal, true)
				} else {
					o.Delete(keys.FlipHorizontal)
				}

				if flipV {
					o.Set(keys.FlipVertical, true)
				} else {
					o.Delete(keys.FlipVertical)
				}

				hashes = append(hashes, s.processImg(imgIndex, o))
			}
		}
	}

	return hashes
}

func (s *RotateAndFlipTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	for i := range 8 {
		img, err := s.ImageDataFactory().NewFromPath(
			s.TestData.Path(fmt.Sprintf("orientation-%d.png", i)),
		)
		s.Require().NoError(err)
		s.imgs = append(s.imgs, img)
	}
}

func (s *RotateAndFlipTestSuite) TestOrientationAutoRotate() {
	o := options.New()
	o.Set(keys.AutoRotate, false)

	// Test with auto_rotate:false - all outputs should be the same
	hashes := []*testutil.ImageHash{}
	for i := range s.imgs {
		hashes = append(hashes, s.processImg(i, o))
	}

	for i := 1; i < len(hashes); i++ {
		d, err := hashes[0].Distance(hashes[i])

		s.Require().NoError(err)
		s.Require().Zero(d)
	}

	// Test with auto_rotate:true - each subsequent output should differ from the previous
	o.Set(keys.AutoRotate, true)

	hashesAr := []*testutil.ImageHash{}
	for i := range s.imgs {
		hashesAr = append(hashesAr, s.processImg(i, o))
	}

	for i := 1; i < len(hashesAr); i++ {
		d, err := hashesAr[i-1].Distance(hashesAr[i])
		s.Require().NoError(err)
		s.Require().NotZero(d)
	}
}

//nolint:dupl
func (s *RotateAndFlipTestSuite) TestRotateFlip() {
	o := options.New()
	o.Set(keys.AutoRotate, false)

	hashes := [][]*testutil.ImageHash{}

	for i := range s.imgs {
		hashes = append(hashes, s.collectRotationsFlips(i, o))
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

//nolint:dupl
func (s *RotateAndFlipTestSuite) TestRotateFlipAutoRotate() {
	o := options.New()
	o.Set(keys.AutoRotate, true)

	hashes := [][]*testutil.ImageHash{}

	for i := range s.imgs {
		hashes = append(hashes, s.collectRotationsFlips(i, o))
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
