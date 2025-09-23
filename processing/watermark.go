package processing

import (
	"context"
	"math"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// watermarkPipeline constructs the watermark processing pipeline.
// This pipeline is applied to the watermark image.
func (p *Processor) watermarkPipeline() Pipeline {
	return Pipeline{
		p.vectorGuardScale,
		p.scaleOnLoad,
		p.colorspaceToProcessing,
		p.scale,
		p.rotateAndFlip,
		p.padding,
	}
}

func (p *Processor) prepareWatermark(
	ctx context.Context,
	wm *vips.Image,
	wmData imagedata.ImageData,
	po *options.ProcessingOptions,
	imgWidth, imgHeight int,
	offsetScale float64,
	framesCount int,
) error {
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

	if err := p.watermarkPipeline().Run(ctx, wm, wmPo, wmData); err != nil {
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

func (p *Processor) applyWatermark(
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
	offsetScale float64,
	framesCount int,
) error {
	if p.watermarkProvider == nil {
		return nil
	}

	wmData, _, err := p.watermarkProvider.Get(ctx, po)
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

	if err := p.prepareWatermark(ctx, wm, wmData, po, width, frameHeight, offsetScale, framesCount); err != nil {
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

func (p *Processor) watermark(c *Context) error {
	if !c.PO.Watermark.Enabled || c.WatermarkProvider == nil {
		return nil
	}

	return p.applyWatermark(c.Ctx, c.Img, c.PO, c.DprScale, 1)
}
