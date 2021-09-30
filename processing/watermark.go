package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
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
	exportColorProfile,
	finalize,
}

func prepareWatermark(wm *vips.Image, wmData *imagedata.ImageData, opts *options.WatermarkOptions, imgWidth, imgHeight int) error {
	if err := wm.Load(wmData, 1, 1.0, 1); err != nil {
		return err
	}

	po := options.NewProcessingOptions()
	po.ResizingType = options.ResizeFit
	po.Dpr = 1
	po.Enlarge = true
	po.Format = wmData.Type

	if opts.Scale > 0 {
		po.Width = imath.Max(imath.Scale(imgWidth, opts.Scale), 1)
		po.Height = imath.Max(imath.Scale(imgHeight, opts.Scale), 1)
	}

	if err := watermarkPipeline.Run(context.Background(), wm, po, wmData); err != nil {
		return err
	}

	if opts.Replicate {
		return wm.Replicate(imgWidth, imgHeight)
	}

	left, top := calcPosition(imgWidth, imgHeight, wm.Width(), wm.Height(), &opts.Gravity, true)

	return wm.Embed(imgWidth, imgHeight, left, top)
}

func applyWatermark(img *vips.Image, wmData *imagedata.ImageData, opts *options.WatermarkOptions, framesCount int) error {
	if err := img.RgbColourspace(); err != nil {
		return err
	}

	if err := img.CopyMemory(); err != nil {
		return err
	}

	wm := new(vips.Image)
	defer wm.Clear()

	width := img.Width()
	height := img.Height()

	if err := prepareWatermark(wm, wmData, opts, width, height/framesCount); err != nil {
		return err
	}

	if framesCount > 1 {
		if err := wm.Replicate(width, height); err != nil {
			return err
		}
	}

	opacity := opts.Opacity * config.WatermarkOpacity

	return img.ApplyWatermark(wm, opacity)
}

func watermark(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Watermark.Enabled || imagedata.Watermark == nil {
		return nil
	}

	return applyWatermark(img, imagedata.Watermark, &po.Watermark, 1)
}
