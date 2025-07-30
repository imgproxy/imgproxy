package processing

import (
	"context"
	"math"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var watermarkPipeline = pipeline{
	prepare,
	scaleOnLoad,
	importColorProfile,
	scale,
	rotateAndFlip,
	padding,
}

func prepareWatermark(wm *vips.Image, wmData imagedatanew.ImageData, opts *options.WatermarkOptions, imgWidth, imgHeight int, offsetScale float64, framesCount int) error {
	if err := wm.Load(imagedata.From(wmData), 1, 1.0, 1); err != nil {
		return err
	}

	po := options.NewProcessingOptions()
	po.ResizingType = options.ResizeFit
	po.Dpr = 1
	po.Enlarge = true
	po.Format = wmData.Format()

	if opts.Scale > 0 {
		po.Width = imath.Max(imath.ScaleToEven(imgWidth, opts.Scale), 1)
		po.Height = imath.Max(imath.ScaleToEven(imgHeight, opts.Scale), 1)
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

		po.Padding.Enabled = true
		po.Padding.Left = offX / 2
		po.Padding.Right = offX - po.Padding.Left
		po.Padding.Top = offY / 2
		po.Padding.Bottom = offY - po.Padding.Top
	}

	if err := watermarkPipeline.Run(context.Background(), wm, po, wmData); err != nil {
		return err
	}

	if opts.ShouldReplicate() || framesCount > 1 {
		// We need to copy image if we're going to replicate.
		// Replication requires image to be read several times, and this requires
		// random access to pixels
		if err := wm.CopyMemory(); err != nil {
			return err
		}
	}

	if opts.ShouldReplicate() {
		if err := wm.Replicate(imgWidth, imgHeight, true); err != nil {
			return err
		}
	}

	// We don't want any headers to be copied from the watermark to the image
	return wm.StripAll()
}

func applyWatermark(img *vips.Image, wmData imagedatanew.ImageData, opts *options.WatermarkOptions, offsetScale float64, framesCount int) error {
	wm := new(vips.Image)
	defer wm.Clear()

	width := img.Width()
	height := img.Height()
	frameHeight := height / framesCount

	if err := prepareWatermark(wm, wmData, opts, width, frameHeight, offsetScale, framesCount); err != nil {
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

func watermark(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata imagedatanew.ImageData) error {
	if !po.Watermark.Enabled || Watermark == nil {
		return nil
	}

	wm, _, err := Watermark.Get(pctx.ctx, po)
	if err != nil {
		return err
	}

	return applyWatermark(img, wm, &po.Watermark, pctx.dprScale, 1)
}
