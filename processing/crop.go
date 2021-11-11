package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func cropImage(img *vips.Image, cropWidth, cropHeight int, gravity *options.GravityOptions) error {
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
		if err := img.SmartCrop(cropWidth, cropHeight); err != nil {
			return err
		}
		// Applying additional modifications after smart crop causes SIGSEGV on Alpine
		// so we have to copy memory after it
		return img.CopyMemory()
	}

	left, top := calcPosition(imgWidth, imgHeight, cropWidth, cropHeight, gravity, false)
	return img.Crop(left, top, cropWidth, cropHeight)
}

func crop(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if err := cropImage(img, pctx.cropWidth, pctx.cropHeight, &pctx.cropGravity); err != nil {
		return err
	}

	// Crop image to the result size
	resultWidth := imath.Scale(po.Width, po.Dpr)
	resultHeight := imath.Scale(po.Height, po.Dpr)

	if po.ResizingType == options.ResizeFillDown {
		if resultWidth > img.Width() {
			resultHeight = imath.Scale(resultHeight, float64(img.Width())/float64(resultWidth))
			resultWidth = img.Width()
		}

		if resultHeight > img.Height() {
			resultWidth = imath.Scale(resultWidth, float64(img.Height())/float64(resultHeight))
			resultHeight = img.Height()
		}
	}

	return cropImage(img, resultWidth, resultHeight, &po.Gravity)
}
