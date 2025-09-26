package processing

import (
	"context"
	"math"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/security"
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

func shouldReplicateWatermark(gt GravityType) bool {
	return gt == GravityReplicate
}

func (p *Processor) prepareWatermark(
	ctx context.Context,
	wm *vips.Image,
	wmData imagedata.ImageData,
	po ProcessingOptions,
	secops security.Options,
	imgWidth, imgHeight int,
	offsetScale float64,
	framesCount int,
) error {
	if err := wm.Load(wmData, 1.0, 0, 1); err != nil {
		return err
	}

	wmPo := p.NewProcessingOptions(options.New())
	wmPo.Set(keys.ResizingType, ResizeFit)
	wmPo.Set(keys.Dpr, 1)
	wmPo.Set(keys.Enlarge, true)
	wmPo.Set(keys.Format, wmData.Format())

	if scale := po.WatermarkScale(); scale > 0 {
		wmPo.Set(keys.Width, max(imath.ScaleToEven(imgWidth, scale), 1))
		wmPo.Set(keys.Height, max(imath.ScaleToEven(imgHeight, scale), 1))
	}

	shouldReplicate := shouldReplicateWatermark(po.WatermarkPosition())

	if shouldReplicate {
		offsetX := po.WatermarkXOffset()
		offsetY := po.WatermarkYOffset()

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

		wmPo.Set(keys.PaddingLeft, padX/2)
		wmPo.Set(keys.PaddingRight, padX-padX/2)
		wmPo.Set(keys.PaddingTop, padY/2)
		wmPo.Set(keys.PaddingBottom, padY-padY/2)
	}

	if err := p.watermarkPipeline().Run(ctx, wm, wmPo, secops, wmData); err != nil {
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

func (p *Processor) applyWatermark(
	ctx context.Context,
	img *vips.Image,
	po ProcessingOptions,
	secops security.Options,
	offsetScale float64,
	framesCount int,
) error {
	if p.watermarkProvider == nil {
		return nil
	}

	wmData, _, err := p.watermarkProvider.Get(ctx, po.Options)
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

	if err := p.prepareWatermark(
		ctx, wm, wmData, po, secops, width, frameHeight, offsetScale, framesCount,
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

	opacity := po.WatermarkOpacity() * p.config.WatermarkOpacity

	position := po.WatermarkPosition()
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
		gr := GravityOptions{
			Type: position,
			X:    po.WatermarkXOffset(),
			Y:    po.WatermarkYOffset(),
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

func (p *Processor) watermark(c *Context) error {
	if p.watermarkProvider == nil || c.PO.WatermarkOpacity() == 0 {
		return nil
	}

	return p.applyWatermark(c.Ctx, c.Img, c.PO, c.SecOps, c.DprScale, 1)
}
