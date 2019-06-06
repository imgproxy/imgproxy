package main

import (
	"context"
	"math"
	"runtime"

	"golang.org/x/sync/errgroup"
)

func extractMeta(img *vipsImage) (int, int, int, bool) {
	width := img.Width()
	height := img.Height()

	angle := vipsAngleD0
	flip := false

	orientation := img.Orientation()

	if orientation >= 5 && orientation <= 8 {
		width, height = height, width
	}
	if orientation == 3 || orientation == 4 {
		angle = vipsAngleD180
	}
	if orientation == 5 || orientation == 6 {
		angle = vipsAngleD90
	}
	if orientation == 7 || orientation == 8 {
		angle = vipsAngleD270
	}
	if orientation == 2 || orientation == 4 || orientation == 5 || orientation == 7 {
		flip = true
	}

	return width, height, angle, flip
}

func calcScale(width, height int, po *processingOptions, imgtype imageType) float64 {
	// If we're going only to crop, we need only to scale down to DPR.
	// Scaling up while cropping is not optimal on this stage, we'll do it later if needed.
	if po.Resize == resizeCrop {
		if po.Dpr < 1 {
			return po.Dpr
		}
		return 1
	}

	var scale float64

	srcW, srcH := float64(width), float64(height)

	if (po.Width == 0 || po.Width == width) && (po.Height == 0 || po.Height == height) {
		scale = 1
	} else {
		wr := float64(po.Width) / srcW
		hr := float64(po.Height) / srcH

		if po.Width == 0 {
			scale = hr
		} else if po.Height == 0 {
			scale = wr
		} else if po.Resize == resizeFit {
			scale = math.Min(wr, hr)
		} else {
			scale = math.Max(wr, hr)
		}
	}

	scale = scale * po.Dpr

	if !po.Enlarge && scale > 1 && imgtype != imageTypeSVG {
		return 1
	}

	if srcW*scale < 1 {
		scale = 1 / srcW
	}

	if srcH*scale < 1 {
		scale = 1 / srcH
	}

	return scale
}

func canScaleOnLoad(imgtype imageType, scale float64) bool {
	if imgtype == imageTypeSVG {
		return true
	}

	if conf.DisableShrinkOnLoad || scale >= 1 {
		return false
	}

	return imgtype == imageTypeJPEG || imgtype == imageTypeWEBP
}

func calcJpegShink(scale float64, imgtype imageType) int {
	shrink := int(1.0 / scale)

	switch {
	case shrink >= 8:
		return 8
	case shrink >= 4:
		return 4
	case shrink >= 2:
		return 2
	}

	return 1
}

func calcCrop(width, height, cropWidth, cropHeight int, gravity *gravityOptions) (left, top int) {
	if gravity.Type == gravityFocusPoint {
		pointX := int(float64(width) * gravity.X)
		pointY := int(float64(height) * gravity.Y)

		left = maxInt(0, minInt(pointX-cropWidth/2, width-cropWidth))
		top = maxInt(0, minInt(pointY-cropHeight/2, height-cropHeight))

		return
	}

	left = (width - cropWidth + 1) / 2
	top = (height - cropHeight + 1) / 2

	if gravity.Type == gravityNorth || gravity.Type == gravityNorthEast || gravity.Type == gravityNorthWest {
		top = 0
	}

	if gravity.Type == gravityEast || gravity.Type == gravityNorthEast || gravity.Type == gravitySouthEast {
		left = width - cropWidth
	}

	if gravity.Type == gravitySouth || gravity.Type == gravitySouthEast || gravity.Type == gravitySouthWest {
		top = height - cropHeight
	}

	if gravity.Type == gravityWest || gravity.Type == gravityNorthWest || gravity.Type == gravitySouthWest {
		left = 0
	}

	return
}

func transformImage(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	var err error

	imgWidth, imgHeight, angle, flip := extractMeta(img)

	hasAlpha := img.HasAlpha()

	scale := calcScale(imgWidth, imgHeight, po, imgtype)

	if scale != 1 && data != nil && canScaleOnLoad(imgtype, scale) {
		if imgtype == imageTypeWEBP || imgtype == imageTypeSVG {
			// Do some scale-on-load
			if err := img.Load(data, imgtype, 1, scale, 1); err != nil {
				return err
			}
		} else if imgtype == imageTypeJPEG {
			// Do some shrink-on-load
			if shrink := calcJpegShink(scale, imgtype); shrink != 1 {
				if err := img.Load(data, imgtype, shrink, 1.0, 1); err != nil {
					return err
				}
			}
		}

		// Update actual image size ans scale after scale-on-load
		imgWidth, imgHeight, _, _ = extractMeta(img)
		scale = calcScale(imgWidth, imgHeight, po, imgtype)
	}

	if err = img.Rad2Float(); err != nil {
		return err
	}

	convertToLinear := conf.UseLinearColorspace && (scale != 1 || po.Dpr != 1)

	if convertToLinear {
		if err = img.ImportColourProfile(true); err != nil {
			return err
		}

		if err = img.LinearColourspace(); err != nil {
			return err
		}
	}

	if scale != 1 {
		if err = img.Resize(scale, hasAlpha); err != nil {
			return err
		}
	}

	// Update actual image size after resize
	imgWidth, imgHeight, _, _ = extractMeta(img)

	checkTimeout(ctx)

	if angle != vipsAngleD0 || flip {
		if err = img.CopyMemory(); err != nil {
			return err
		}

		if angle != vipsAngleD0 {
			if err = img.Rotate(angle); err != nil {
				return err
			}
		}

		if flip {
			if err = img.Flip(); err != nil {
				return err
			}
		}
	}

	checkTimeout(ctx)

	cropW, cropH := po.Width, po.Height

	if po.Dpr < 1 || (po.Dpr > 1 && po.Resize != resizeCrop) {
		cropW = int(float64(cropW) * po.Dpr)
		cropH = int(float64(cropH) * po.Dpr)
	}

	if cropW == 0 {
		cropW = imgWidth
	} else {
		cropW = minInt(cropW, imgWidth)
	}

	if cropH == 0 {
		cropH = imgHeight
	} else {
		cropH = minInt(cropH, imgHeight)
	}

	if cropW < imgWidth || cropH < imgHeight {
		if po.Gravity.Type == gravitySmart {
			if err = img.CopyMemory(); err != nil {
				return err
			}
			if err = img.SmartCrop(cropW, cropH); err != nil {
				return err
			}
			// Applying additional modifications after smart crop causes SIGSEGV on Alpine
			// so we have to copy memory after it
			if err = img.CopyMemory(); err != nil {
				return err
			}
		} else {
			left, top := calcCrop(imgWidth, imgHeight, cropW, cropH, &po.Gravity)
			if err = img.Crop(left, top, cropW, cropH); err != nil {
				return err
			}
		}

		checkTimeout(ctx)
	}

	if po.Enlarge && po.Resize == resizeCrop && po.Dpr > 1 {
		// We didn't enlarge the image before, because is wasn't optimal. Now it's time to do it
		if err = img.Resize(po.Dpr, hasAlpha); err != nil {
			return err
		}
		if err = img.CopyMemory(); err != nil {
			return err
		}
	}

	if convertToLinear {
		if err = img.FixColourspace(); err != nil {
			return err
		}
	} else {
		if err = img.ImportColourProfile(false); err != nil {
			return err
		}
	}

	if po.Expand && (po.Width > img.Width() || po.Height > img.Height()) {
		if err = img.EnsureAlpha(); err != nil {
			return err
		}

		hasAlpha = true

		if err = img.Embed(gravityCenter, po.Width, po.Height, 0, 0); err != nil {
			return err
		}
	}

	if hasAlpha && (po.Flatten || po.Format == imageTypeJPEG) {
		if err = img.Flatten(po.Background); err != nil {
			return err
		}
	}

	if po.Blur > 0 {
		if err = img.Blur(po.Blur); err != nil {
			return err
		}
	}

	if po.Sharpen > 0 {
		if err = img.Sharpen(po.Sharpen); err != nil {
			return err
		}
	}

	checkTimeout(ctx)

	if po.Watermark.Enabled {
		if err = img.ApplyWatermark(&po.Watermark); err != nil {
			return err
		}
	}

	return img.FixColourspace()
}

func transformAnimated(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	imgWidth := img.Width()

	frameHeight, err := img.GetInt("page-height")
	if err != nil {
		return err
	}

	framesCount := minInt(img.Height()/frameHeight, conf.MaxGifFrames)

	// Double check dimensions because animated image has many frames
	if err := checkDimensions(imgWidth, frameHeight*framesCount); err != nil {
		return err
	}

	// Vips 8.8+ supports n-pages and doesn't load the whole animated image on header access
	if nPages, _ := img.GetInt("n-pages"); nPages > 0 {
		scale := calcScale(imgWidth, frameHeight, po, imgtype)

		if nPages > framesCount || canScaleOnLoad(imgtype, scale) {
			// Do some scale-on-load
			if err := img.Load(data, imgtype, 1, scale, framesCount); err != nil {
				return err
			}
		}

		imgWidth = img.Width()

		frameHeight, err = img.GetInt("page-height")
		if err != nil {
			return err
		}
	}

	delay, err := img.GetInt("gif-delay")
	if err != nil {
		return err
	}

	loop, err := img.GetInt("gif-loop")
	if err != nil {
		return err
	}

	frames := make([]*vipsImage, framesCount)
	defer func() {
		for _, frame := range frames {
			frame.Clear()
		}
	}()

	var errg errgroup.Group

	for i := 0; i < framesCount; i++ {
		ind := i
		errg.Go(func() error {
			frame := new(vipsImage)

			if err := img.Extract(frame, 0, ind*frameHeight, imgWidth, frameHeight); err != nil {
				return err
			}

			if err := transformImage(ctx, frame, nil, po, imgtype); err != nil {
				return err
			}

			frames[ind] = frame

			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		return err
	}

	checkTimeout(ctx)

	if err := img.Arrayjoin(frames); err != nil {
		return err
	}

	img.SetInt("page-height", frames[0].Height())
	img.SetInt("gif-delay", delay)
	img.SetInt("gif-loop", loop)
	img.SetInt("n-pages", framesCount)

	return nil
}

func processImage(ctx context.Context) ([]byte, context.CancelFunc, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if newRelicEnabled {
		newRelicCancel := startNewRelicSegment(ctx, "Processing image")
		defer newRelicCancel()
	}

	if prometheusEnabled {
		defer startPrometheusDuration(prometheusProcessingDuration)()
	}

	defer vipsCleanup()

	po := getProcessingOptions(ctx)
	data := getImageData(ctx).Bytes()
	imgtype := getImageType(ctx)

	if po.Gravity.Type == gravitySmart && !vipsSupportSmartcrop {
		return nil, func() {}, errSmartCropNotSupported
	}

	if po.Format == imageTypeUnknown {
		if vipsTypeSupportSave[imgtype] {
			po.Format = imgtype
		} else {
			po.Format = imageTypeJPEG
		}
	}

	animationSupport := conf.MaxGifFrames > 1 && vipsSupportAnimation(imgtype) && vipsSupportAnimation(po.Format)

	pages := 1
	if animationSupport {
		pages = -1
	}

	img := new(vipsImage)
	defer img.Clear()

	if err := img.Load(data, imgtype, 1, 1.0, pages); err != nil {
		return nil, func() {}, err
	}

	if animationSupport && img.IsAnimated() {
		if err := transformAnimated(ctx, img, data, po, imgtype); err != nil {
			return nil, func() {}, err
		}
	} else {
		if err := transformImage(ctx, img, data, po, imgtype); err != nil {
			return nil, func() {}, err
		}
	}

	checkTimeout(ctx)

	if po.Format == imageTypeGIF {
		if err := img.CastUchar(); err != nil {
			return nil, func() {}, err
		}
		checkTimeout(ctx)
	}

	return img.Save(po.Format, po.Quality)
}
