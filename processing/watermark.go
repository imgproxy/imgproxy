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
	padding,
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

	if opts.Replicate {
		po.Padding.Enabled = true
		po.Padding.Left = int(opts.Gravity.X / 2)
		po.Padding.Right = int(opts.Gravity.X) - po.Padding.Left
		po.Padding.Top = int(opts.Gravity.Y / 2)
		po.Padding.Bottom = int(opts.Gravity.Y) - po.Padding.Top
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

func watermark(_ *pipelineContext, img *vips.Image, po *options.ProcessingOptions, _ *imagedata.ImageData) error {
	if !po.Watermark.Enabled {
		return nil
	}

	var wm = imagedata.Watermark
	if &po.Watermark.Url != nil && len(po.Watermark.Url) > 0 {
		if !imagedata.CachedWatermark.Contains(po.Watermark.Url) {
			imgData, err := imagedata.Download(po.Watermark.Url, "url-based watermark", nil, nil)
			if err != nil {
				return err
			}
			imagedata.CachedWatermark.Add(po.Watermark.Url, imgData)
		}

		wmu, ok := imagedata.CachedWatermark.Get(po.Watermark.Url)
		if ok {
			wm = wmu.(*imagedata.ImageData)
		} else {
			// 워터마크 url이 존재하는데 이미지를 불러오지 못했다면 nil을 설정한다.
			wm = nil
		}
	}

	if wm == nil {
		return nil
	}

	return applyWatermark(img, wm, &po.Watermark, 1)
}
