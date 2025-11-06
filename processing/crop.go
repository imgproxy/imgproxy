package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func cropImage(img *vips.Image, cropWidth, cropHeight int, gravity *GravityOptions, offsetScale float64) error {
	if cropWidth == 0 && cropHeight == 0 {
		return nil
	}

	imgWidth, imgHeight := img.Width(), img.Height()

	cropWidth = imath.MinNonZero(cropWidth, imgWidth)
	cropHeight = imath.MinNonZero(cropHeight, imgHeight)

	if cropWidth >= imgWidth && cropHeight >= imgHeight {
		return nil
	}

	if gravity.Type == GravitySmart {
		if err := img.CopyMemory(); err != nil {
			return err
		}
		return img.SmartCrop(cropWidth, cropHeight)
	}

	left, top := calcPosition(imgWidth, imgHeight, cropWidth, cropHeight, gravity, offsetScale, false)
	return img.Crop(left, top, cropWidth, cropHeight)
}

func (p *Processor) crop(c *Context) error {
	width, height := c.CropWidth, c.CropHeight
	rotateAngle := c.PO.Rotate()
	flipX := c.PO.FlipHorizontal()
	flipY := c.PO.FlipVertical()

	// Since we crop before rotating and flipping,
	// we need to adjust gravity options accordingly.
	// After rotation and flipping, we'll get the same result
	// as if we cropped with the original gravity options after
	// rotation and flipping.
	//
	// During rotation/flipping, we first apply the EXIF orientation,
	// then the user-specified operations.
	// So here we apply the adjustments in the reverse order.
	opts := c.CropGravity
	opts.RotateAndFlip(rotateAngle, flipX, flipY)
	opts.RotateAndFlip(c.Angle, c.Flip, false)

	// If the final image is rotated by 90 or 270 degrees,
	// we need to swap width and height for cropping.
	// After rotation, we'll get the originally intended dimensions.
	if (c.Angle+rotateAngle)%180 == 90 {
		width, height = height, width
	}

	// Since we crop before scaling, we shouldn't consider DPR
	return cropImage(c.Img, width, height, &opts, 1.0)
}

func (p *Processor) cropToResult(c *Context) error {
	gravity := c.PO.Gravity()
	return cropImage(c.Img, c.ResultCropWidth, c.ResultCropHeight, &gravity, c.DprScale)
}
