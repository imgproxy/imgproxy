package processing

import (
	"context"
	"math"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var watermarkPipeline = pipeline{
	vectorGuardScale,
	prepare,
	scaleOnLoad,
	colorspaceToProcessing,
	scale,
	rotateAndFlip,
	padding,
}

func prepareWatermark(wm *vips.Image, wmData imagedata.ImageData, po *options.ProcessingOptions, config *Config, imgWidth, imgHeight int, offsetScale float64, framesCount int) error {
	if err := wm.Load(wmData, 1, 1.0, 1); err != nil {
		return err
	}

	opts := po.Watermark

	wmPo := po.Default()
	wmPo.ResizingType = options.ResizeFit
	wmPo.Dpr = 1
	wmPo.Enlarge = true
	wmPo.Format = wmData.Format()

	if opts.Scale > 0 {
		wmPo.Width = max(imath.ScaleToEven(imgWidth, opts.Scale), 1)
		wmPo.Height = max(imath.ScaleToEven(imgHeight, opts.Scale), 1)
	}

	if opts.ShouldReplicate() {
		var offX, offY int

		if math.Abs(opts.Position.X) >= 1.0 {
			offX = imath.RoundToEven(opts.Position.X * offsetScale)
		} else {
			offX = imath.ScaleToEven(imgWidth, opts.Position.X)
		}

		if math.Abs(opts.Position.Y) >= 1.0 {
			offY = imath.RoundToEven(opts.Position.Y * offsetScale)
		} else {
			offY = imath.ScaleToEven(imgHeight, opts.Position.Y)
		}

		wmPo.Padding.Enabled = true
		wmPo.Padding.Left = offX / 2
		wmPo.Padding.Right = offX - wmPo.Padding.Left
		wmPo.Padding.Top = offY / 2
		wmPo.Padding.Bottom = offY - wmPo.Padding.Top
	}

	if err := watermarkPipeline.Run(context.Background(), wm, wmPo, wmData, nil, config); err != nil {
		return err
	}

	// We need to copy the image to ensure that it is in memory since we will
	// close it after watermark processing is done.
	if err := wm.CopyMemory(); err != nil {
		return err
	}

	if opts.ShouldReplicate() {
		if err := wm.Replicate(imgWidth, imgHeight, true); err != nil {
			return err
		}
	}

	// We don't want any headers to be copied from the watermark to the image
	return wm.StripAll()
}

func applyWatermark(
	ctx context.Context,
	img *vips.Image,
	watermark auximageprovider.Provider,
	po *options.ProcessingOptions,
	offsetScale float64,
	framesCount int,
	config *Config,
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

	opts := po.Watermark

	wm := new(vips.Image)
	defer wm.Clear()

	width := img.Width()
	height := img.Height()
	frameHeight := height / framesCount

	if err := prepareWatermark(wm, wmData, po, config, width, frameHeight, offsetScale, framesCount); err != nil {
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

	opacity := opts.Opacity * config.WatermarkOpacity

	// If we replicated the watermark and need to apply it to an animated image,
	// it is faster to replicate the watermark to all the image and apply it single-pass
	if opts.ShouldReplicate() && framesCount > 1 {
		if err := wm.Replicate(width, height, false); err != nil {
			return err
		}

		return img.ApplyWatermark(wm, 0, 0, opacity)
	}

	left, top := 0, 0
	wmWidth := wm.Width()
	wmHeight := wm.Height()

	if !opts.ShouldReplicate() {
		left, top = calcPosition(width, frameHeight, wmWidth, wmHeight, &opts.Position, offsetScale, true)
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

func watermark(
	pctx *pipelineContext,
	img *vips.Image,
	po *options.ProcessingOptions,
	imgdata imagedata.ImageData,
) error {
	if !po.Watermark.Enabled || pctx.watermarkProvider == nil {
		return nil
	}

	return applyWatermark(pctx.ctx, img, pctx.watermarkProvider, po, pctx.dprScale, 1, pctx.config)
}
