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

	opts := pctx.cropGravity
	opts.RotateAndFlip(pctx.angle, pctx.flip)
	opts.RotateAndFlip(po.Rotate, false)

	if (pctx.angle+po.Rotate)%180 == 90 {
		width, height = height, width
	}

	// Since we crop before scaling, we shouldn't consider DPR
	return cropImage(img, width, height, &opts, 1.0)
}

func cropToResult(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	return cropImage(img, pctx.resultCropWidth, pctx.resultCropHeight, &po.Gravity, pctx.dprScale)
}
