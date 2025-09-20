package processing

import (
	"context"
	"math"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var watermarkPipeline = Pipeline{
	vectorGuardScale,
	prepare,
	scaleOnLoad,
	colorspaceToProcessing,
	scale,
	rotateAndFlip,
	padding,
}

func shouldReplicateWatermark(gt options.GravityType) bool {
	return gt == options.GravityReplicate
}

func prepareWatermark(
	ctx context.Context,
	runner *Runner,
	wm *vips.Image,
	wmData imagedata.ImageData,
	po options.Options,
	secops security.Options,
	imgWidth, imgHeight int,
	offsetScale float64,
) error {
	if err := wm.Load(wmData, 1, 1.0, 1); err != nil {
		return err
	}

	wmPo := options.New()
	wmPo[keys.ResizingType] = options.ResizeFit
	wmPo[keys.Dpr] = 1
	wmPo[keys.Enlarge] = true
	wmPo[keys.Format] = wmData.Format()

	if scale := options.GetFloat(po, keys.WatermarkScale, 0.0); scale > 0 {
		wmPo[keys.Width] = max(imath.ScaleToEven(imgWidth, scale), 1)
		wmPo[keys.Height] = max(imath.ScaleToEven(imgHeight, scale), 1)
	}

	shouldReplicate := shouldReplicateWatermark(
		options.Get(po, keys.WatermarkPosition, options.GravityCenter),
	)

	if shouldReplicate {
		offsetX := options.GetFloat(po, keys.WatermarkXOffset, 0.0)
		offsetY := options.GetFloat(po, keys.WatermarkYOffset, 0.0)

		var padX, padY int

		if math.Abs(offsetX) >= 1.0 {
			padX = imath.RoundToEven(offsetX * offsetScale)
		} else {
			padX = imath.ScaleToEven(imgWidth, offsetX)
		}

		if math.Abs(offsetY) >= 1.0 {
			padY = imath.RoundToEven(offsetY * offsetScale)
		} else {
			padY = imath.ScaleToEven(imgHeight, offsetY)
		}

		wmPo[keys.PaddingEnabled] = true
		wmPo[keys.PaddingLeft] = padX / 2
		wmPo[keys.PaddingRight] = padX - padX/2
		wmPo[keys.PaddingTop] = padY / 2
		wmPo[keys.PaddingBottom] = padY - padY/2
	}

	if err := runner.Run(watermarkPipeline, ctx, wm, wmPo, secops, wmData); err != nil {
		return err
	}

	// We need to copy the image to ensure that it is in memory since we will
	// close it after watermark processing is done.
	if err := wm.CopyMemory(); err != nil {
		return err
	}

	if shouldReplicate {
		if err := wm.Replicate(imgWidth, imgHeight, true); err != nil {
			return err
		}
	}

	// We don't want any headers to be copied from the watermark to the image
	return wm.StripAll()
}

func applyWatermark(
	ctx context.Context,
	runner *Runner,
	img *vips.Image,
	watermark auximageprovider.Provider,
	po options.Options,
	secops security.Options,
	offsetScale float64,
	framesCount int,
) error {
	if watermark == nil {
		return nil
	}

	wmData, _, err := watermark.Get(ctx, po)
	if err != nil {
		return err
	}
	if wmData == nil {
		return nil
	}
	defer wmData.Close()

	wm := new(vips.Image)
	defer wm.Clear()

	width := img.Width()
	height := img.Height()
	frameHeight := height / framesCount

	if err := prepareWatermark(
		ctx, runner, wm, wmData, po, secops, width, frameHeight, offsetScale,
	); err != nil {
		return err
	}

	if !img.ColourProfileImported() {
		if err := img.ImportColourProfile(); err != nil {
			return err
		}
	}

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	// TODO: Use runner config
	opacity := config.WatermarkOpacity * options.GetFloat(po, keys.WatermarkOpacity, 1.0)

	position := options.Get(po, keys.WatermarkPosition, options.GravityCenter)
	shouldReplicate := shouldReplicateWatermark(position)

	// If we replicated the watermark and need to apply it to an animated image,
	// it is faster to replicate the watermark to all the image and apply it single-pass
	if shouldReplicate && framesCount > 1 {
		if err := wm.Replicate(width, height, false); err != nil {
			return err
		}

		return img.ApplyWatermark(wm, 0, 0, opacity)
	}

	left, top := 0, 0
	wmWidth := wm.Width()
	wmHeight := wm.Height()

	if !shouldReplicate {
		gr := options.GravityOptions{
			Type: position,
			X:    options.GetFloat(po, keys.WatermarkXOffset, 0.0),
			Y:    options.GetFloat(po, keys.WatermarkYOffset, 0.0),
		}
		left, top = calcPosition(width, frameHeight, wmWidth, wmHeight, &gr, offsetScale, true)
	}

	if left >= width || top >= height || -left >= wmWidth || -top >= wmHeight {
		// Watermark is completely outside the image
		return nil
	}

	// if watermark is partially outside the image, it may partially be visible
	// on the next frame. We need to crop it vertically.
	// We don't care about horizontal overlap, as frames are stacked vertically
	if framesCount > 1 {
		cropTop := 0
		cropHeight := wmHeight

		if top < 0 {
			cropTop = -top
			cropHeight -= cropTop
			top = 0
		}

		if top+cropHeight > frameHeight {
			cropHeight = frameHeight - top
		}

		if cropTop > 0 || cropHeight < wmHeight {
			if err := wm.Crop(0, cropTop, wmWidth, cropHeight); err != nil {
				return err
			}
		}
	}

	for i := 0; i < framesCount; i++ {
		if err := img.ApplyWatermark(wm, left, top, opacity); err != nil {
			return err
		}
		top += frameHeight
	}

	return nil
}

func watermark(c *Context) error {
	if !options.Get(c.PO, keys.WatermarkEnabled, false) || c.WatermarkProvider == nil {
		return nil
	}

	return applyWatermark(
		c.Ctx, c.Runner(), c.Img, c.WatermarkProvider, c.PO, c.SecOps, c.DprScale, 1,
	)
}
