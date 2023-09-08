package processing

import (
	"context"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/utils"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/mitchellh/copystructure"
	"math"

	log "github.com/sirupsen/logrus"
)

var blurPaddingPipeline = pipeline{
	trim,
	prepare,
	scaleOnLoad,
	importColorProfile,
	crop,
	scale,
	rotateAndFlip,
	cropToResult,
	applyFilters,
	extend,
	extendAspectRatio,
	fixSize,
	flatten,
	exportColorProfile,
	stripMetadata,
}

func padding(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Padding.Enabled {
		return nil
	}

	paddingTop := imath.ScaleToEven(po.Padding.Top, pctx.dprScale)
	paddingRight := imath.ScaleToEven(po.Padding.Right, pctx.dprScale)
	paddingBottom := imath.ScaleToEven(po.Padding.Bottom, pctx.dprScale)
	paddingLeft := imath.ScaleToEven(po.Padding.Left, pctx.dprScale)

	outputWidth := img.Width() + paddingLeft + paddingRight
	outputHeight := img.Height() + paddingTop + paddingBottom

	// Use blur effect for background. This uses a darkened, blurred, and
	// slightly blown up version of the image instead of a solid color.
	if po.Background.Effect == "blur" {
		// Make a copy of the image for embedding later
		centerImage := new(vips.Image)
		defer centerImage.Clear()

		// PIPeline version
		// set options
		srcWidth, srcHeight, _, _ := extractMeta(img, po.Rotate, po.AutoRotate)

		copiedPo, err := copystructure.Copy(po)
		if err != nil {
			log.Error("Error copying processing options")
		}
		centerPo, ok := copiedPo.(*options.ProcessingOptions)
		if !ok {
			log.Errorf("Error: Copied processing options is not of the expected type: %s", centerPo)
		}
		cropWidth := calcCropSize(srcWidth, po.Crop.Width)
		cropHeight := calcCropSize(srcHeight, po.Crop.Height)

		cropGravity := po.Crop.Gravity
		if cropGravity.Type == options.GravityUnknown {
			cropGravity = po.Gravity
		}

		widthToScale := utils.MinNonZeroInt(cropWidth, srcWidth)
		heightToScale := utils.MinNonZeroInt(cropHeight, srcHeight)

		scale := utils.OldCalcScale(widthToScale, heightToScale, po, imgdata.Type)
		// Note: original version did two crops, this is more difficult with the new pipeline, was it needed?
		//dprWidth := utils.ScaleInt(po.Width, po.Dpr)
		//dprHeight := utils.ScaleInt(po.Height, po.Dpr)

		if cropWidth > 0 {
			cropWidth = utils.MaxInt(1, utils.ScaleInt(cropWidth, scale))
		}
		if cropHeight > 0 {
			cropHeight = utils.MaxInt(1, utils.ScaleInt(cropHeight, scale))
		}
		if cropGravity.Type != options.GravityFocusPoint {
			cropGravity.X *= scale
			cropGravity.Y *= scale
		}
		centerPo.Crop.Width = float64(cropWidth)
		centerPo.Crop.Height = float64(cropHeight)
		centerPo.Crop.Gravity = cropGravity

		if err := centerImage.Load(imgdata, 1, 1.0, 1); err != nil {
			return err
		}

		if err := blurPaddingPipeline.Run(context.Background(), centerImage, centerPo, imgdata); err != nil {
			return err
		}

		// Resize to the image area and then center smart crop to trim off excess.
		outputScale := math.Max(float64(outputWidth)/float64(img.Width()), float64(outputHeight)/float64(img.Height()))
		// Pad out a little more that image size because the edge of the image repeating
		// looks bad.
		outputScale *= 1.1

		if err := img.OldResize(outputScale, false); err != nil {
			return err
		}

		if err := img.CenterFill(outputWidth, outputHeight); err != nil {
			return err
		}

		if err := img.ColorAdjust(0.6); err != nil {
			return err
		}

		if err := img.Blur(40); err != nil {
			return err
		}

		return img.EmbedImage(
			paddingLeft,
			paddingTop,
			centerImage,
		)
	} else {
		// Padding with color fill
		return img.Embed(
			img.Width()+paddingLeft+paddingRight,
			img.Height()+paddingTop+paddingBottom,
			paddingLeft,
			paddingTop,
		)
	}
}
