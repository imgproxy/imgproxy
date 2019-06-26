package main

import (
	"context"
	"math"
	"runtime"

	"golang.org/x/sync/errgroup"
)

const msgSmartCropNotSupported = "Smart crop is not supported by used version of libvips"

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
	var scale float64

	srcW, srcH := float64(width), float64(height)

	if (po.Width == 0 || po.Width == width) && (po.Height == 0 || po.Height == height) {
		scale = 1
	} else {
		wr := float64(po.Width) / srcW
		hr := float64(po.Height) / srcH

		rt := po.Resize

		if rt == resizeAuto {
			srcD := width - height
			dstD := po.Width - po.Height

			if (srcD >= 0 && dstD >= 0) || (srcD < 0 && dstD < 0) {
				rt = resizeFill
			} else {
				rt = resizeFit
			}
		}

		if po.Width == 0 {
			scale = hr
		} else if po.Height == 0 {
			scale = wr
		} else if rt == resizeFit {
			scale = math.Min(wr, hr)
		} else {
			scale = math.Max(wr, hr)
		}
	}

	if !po.Enlarge && scale > 1 && imgtype != imageTypeSVG {
		scale = 1
	}

	scale = scale * po.Dpr

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

	offX, offY := int(gravity.X), int(gravity.Y)

	left = (width-cropWidth+1)/2 + offX
	top = (height-cropHeight+1)/2 + offY

	if gravity.Type == gravityNorth || gravity.Type == gravityNorthEast || gravity.Type == gravityNorthWest {
		top = 0 + offY
	}

	if gravity.Type == gravityEast || gravity.Type == gravityNorthEast || gravity.Type == gravitySouthEast {
		left = width - cropWidth - offX
	}

	if gravity.Type == gravitySouth || gravity.Type == gravitySouthEast || gravity.Type == gravitySouthWest {
		top = height - cropHeight - offY
	}

	if gravity.Type == gravityWest || gravity.Type == gravityNorthWest || gravity.Type == gravitySouthWest {
		left = 0 + offX
	}

	left = maxInt(0, minInt(left, width-cropWidth))
	top = maxInt(0, minInt(top, height-cropHeight))

	return
}

func cropImage(img *vipsImage, cropWidth, cropHeight int, gravity *gravityOptions) error {
	if cropWidth == 0 && cropHeight == 0 {
		return nil
	}

	imgWidth, imgHeight := img.Width(), img.Height()

	if cropWidth == 0 {
		cropWidth = imgWidth
	} else {
		cropWidth = minInt(cropWidth, imgWidth)
	}

	if cropHeight == 0 {
		cropHeight = imgHeight
	} else {
		cropHeight = minInt(cropHeight, imgHeight)
	}

	if cropWidth >= imgWidth && cropHeight >= imgHeight {
		return nil
	}

	if gravity.Type == gravitySmart {
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

	left, top := calcCrop(imgWidth, imgHeight, cropWidth, cropHeight, gravity)
	return img.Crop(left, top, cropWidth, cropHeight)
}

func scaleSize(size int, scale float64) int {
	if size == 0 {
		return 0
	}

	return roundToInt(float64(size) * scale)
}

func transformImage(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	var err error

	srcWidth, srcHeight, angle, flip := extractMeta(img)

	cropWidth, cropHeight := po.Crop.Width, po.Crop.Height

	cropGravity := po.Crop.Gravity
	if cropGravity.Type == gravityUnknown {
		cropGravity = po.Gravity
	}

	widthToScale, heightToScale := srcWidth, srcHeight

	if cropWidth > 0 {
		widthToScale = minInt(cropWidth, srcWidth)
	}
	if cropHeight > 0 {
		heightToScale = minInt(cropHeight, srcHeight)
	}

	scale := calcScale(widthToScale, heightToScale, po, imgtype)

	cropWidth = scaleSize(cropWidth, scale)
	cropHeight = scaleSize(cropHeight, scale)
	cropGravity.X = cropGravity.X * scale
	cropGravity.Y = cropGravity.Y * scale

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

		// Update scale after scale-on-load
		newWidth, newHeight, _, _ := extractMeta(img)

		widthToScale = scaleSize(widthToScale, float64(newWidth)/float64(srcWidth))
		heightToScale = scaleSize(heightToScale, float64(newHeight)/float64(srcHeight))

		scale = calcScale(widthToScale, heightToScale, po, imgtype)
	}

	if err = img.Rad2Float(); err != nil {
		return err
	}

	iccImported := false
	convertToLinear := conf.UseLinearColorspace && (scale != 1 || po.Dpr != 1)

	if convertToLinear || !img.IsSRGB() {
		if err = img.ImportColourProfile(true); err != nil {
			return err
		}
		iccImported = true
	}

	if convertToLinear {
		if err = img.LinearColourspace(); err != nil {
			return err
		}
	} else {
		if err = img.RgbColourspace(); err != nil {
			return err
		}
	}

	hasAlpha := img.HasAlpha()

	if scale != 1 {
		if err = img.Resize(scale, hasAlpha); err != nil {
			return err
		}
	}

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

	dprWidth := roundToInt(float64(po.Width) * po.Dpr)
	dprHeight := roundToInt(float64(po.Height) * po.Dpr)

	if cropGravity.Type == po.Gravity.Type && cropGravity.Type != gravityFocusPoint {
		if cropWidth == 0 {
			cropWidth = dprWidth
		} else if dprWidth > 0 {
			cropWidth = minInt(cropWidth, dprWidth)
		}

		if cropHeight == 0 {
			cropHeight = dprHeight
		} else if dprHeight > 0 {
			cropHeight = minInt(cropHeight, dprHeight)
		}

		sumGravity := gravityOptions{
			Type: cropGravity.Type,
			X:    cropGravity.X + po.Gravity.X,
			Y:    cropGravity.Y + po.Gravity.Y,
		}

		if err = cropImage(img, cropWidth, cropHeight, &sumGravity); err != nil {
			return err
		}
	} else {
		if err = cropImage(img, cropWidth, cropHeight, &cropGravity); err != nil {
			return err
		}
		if err = cropImage(img, dprWidth, dprHeight, &po.Gravity); err != nil {
			return err
		}
	}

	checkTimeout(ctx)

	if !iccImported {
		if err = img.ImportColourProfile(false); err != nil {
			return err
		}
	}

	if err = img.RgbColourspace(); err != nil {
		return err
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

	if po.Extend && (po.Width > img.Width() || po.Height > img.Height()) {
		if err = img.Embed(gravityCenter, po.Width, po.Height, 0, 0, po.Background); err != nil {
			return err
		}
	}

	checkTimeout(ctx)

	if po.Watermark.Enabled {
		if err = img.ApplyWatermark(&po.Watermark); err != nil {
			return err
		}
	}

	return img.RgbColourspace()
}

func transformAnimated(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	imgWidth := img.Width()

	frameHeight, err := img.GetInt("page-height")
	if err != nil {
		return err
	}

	framesCount := minInt(img.Height()/frameHeight, conf.MaxAnimationFrames)

	// Double check dimensions because animated image has many frames
	if err := checkDimensions(imgWidth, frameHeight*framesCount); err != nil {
		return err
	}

	// Vips 8.8+ supports n-pages and doesn't load the whole animated image on header access
	if nPages, _ := img.GetInt("n-pages"); nPages > 0 {
		scale := 1.0

		// Don't do scale on load if we need to crop
		if po.Crop.Width == 0 && po.Crop.Height == 0 {
			scale = calcScale(imgWidth, frameHeight, po, imgtype)
		}

		if nPages > framesCount || canScaleOnLoad(imgtype, scale) {
			logNotice("Animated scale on load")
			// Do some scale-on-load and load only the needed frames
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

	if po.Format == imageTypeUnknown {
		if po.PreferWebP && vipsTypeSupportSave[imageTypeWEBP] {
			po.Format = imageTypeWEBP
		} else if vipsTypeSupportSave[imgtype] && imgtype != imageTypeHEIC {
			po.Format = imgtype
		} else {
			po.Format = imageTypeJPEG
		}
	} else if po.EnforceWebP && vipsTypeSupportSave[imageTypeWEBP] {
		po.Format = imageTypeWEBP
	}

	if !vipsSupportSmartcrop {
		if po.Gravity.Type == gravitySmart {
			logWarning(msgSmartCropNotSupported)
			po.Gravity.Type = gravityCenter
		}
		if po.Crop.Gravity.Type == gravitySmart {
			logWarning(msgSmartCropNotSupported)
			po.Crop.Gravity.Type = gravityCenter
		}
	}

	if po.Resize == resizeCrop {
		logWarning("`crop` resizing type is deprecated and will be removed in future versions. Use `crop` processing option instead")

		po.Crop.Width, po.Crop.Height = po.Width, po.Height

		po.Resize = resizeFit
		po.Width, po.Height = 0, 0
	}

	animationSupport := conf.MaxAnimationFrames > 1 && vipsSupportAnimation(imgtype) && vipsSupportAnimation(po.Format)

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
