package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func cropImage(img *vips.Image, cropWidth, cropHeight int, gravity *options.GravityOptions, offsetScale float64) error {
	if cropWidth == 0 && cropHeight == 0 {
		return nil
	}

	imgWidth, imgHeight := img.Width(), img.Height()

	cropWidth = imath.MinNonZero(cropWidth, imgWidth)
	cropHeight = imath.MinNonZero(cropHeight, imgHeight)

	if cropWidth >= imgWidth && cropHeight >= imgHeight {
		return nil
	}

	if gravity.Type == options.GravitySmart {
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

	opts := c.CropGravity
	opts.RotateAndFlip(c.Angle, c.Flip)
	opts.RotateAndFlip(c.PO.Rotate, false)

	if (c.Angle+c.PO.Rotate)%180 == 90 {
		width, height = height, width
	}

	// Since we crop before scaling, we shouldn't consider DPR
	return cropImage(c.Img, width, height, &opts, 1.0)
}

func (p *Processor) cropToResult(c *Context) error {
	return cropImage(c.Img, c.ResultCropWidth, c.ResultCropHeight, &c.PO.Gravity, c.DprScale)
}
