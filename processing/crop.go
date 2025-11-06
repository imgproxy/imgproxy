package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
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

func crop(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	width, height := pctx.cropWidth, pctx.cropHeight

	// Since we crop before rotating and flipping,
	// we need to adjust gravity options accordingly.
	// After rotation and flipping, we'll get the same result
	// as if we cropped with the original gravity options after
	// rotation and flipping.
	//
	// During rotation/flipping, we first apply the EXIF orientation,
	// then the user-specified operations.
	// So here we apply the adjustments in the reverse order.
	opts := pctx.cropGravity
	opts.RotateAndFlip(po.Rotate, po.Flip.Horizontal, po.Flip.Vertical)
	opts.RotateAndFlip(pctx.angle, pctx.flip, false)

	// If the final image is rotated by 90 or 270 degrees,
	// we need to swap width and height for cropping.
	// After rotation, we'll get the originally intended dimensions.
	if (pctx.angle+po.Rotate)%180 == 90 {
		width, height = height, width
	}

	// Since we crop before scaling, we shouldn't consider DPR
	return cropImage(img, width, height, &opts, 1.0)
}

func cropToResult(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	return cropImage(img, pctx.resultCropWidth, pctx.resultCropHeight, &po.Gravity, pctx.dprScale)
}
