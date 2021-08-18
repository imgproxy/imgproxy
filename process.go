package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"runtime"

	"github.com/imgproxy/imgproxy/v2/imagemeta"
)

const (
	msgSmartCropNotSupported = "Smart crop is not supported by used version of libvips"

	// https://chromium.googlesource.com/webm/libwebp/+/refs/heads/master/src/webp/encode.h#529
	webpMaxDimension = 16383.0
)

var errConvertingNonSvgToSvg = newError(422, "Converting non-SVG images to SVG is not supported", "Converting non-SVG images to SVG is not supported")

func imageTypeLoadSupport(imgtype imageType) bool {
	return imgtype == imageTypeSVG ||
		imgtype == imageTypeICO ||
		vipsTypeSupportLoad[imgtype]
}

func imageTypeSaveSupport(imgtype imageType) bool {
	return imgtype == imageTypeSVG || vipsTypeSupportSave[imgtype]
}

func imageTypeGoodForWeb(imgtype imageType) bool {
	return imgtype != imageTypeTIFF &&
		imgtype != imageTypeBMP
}

func canSwitchFormat(src, dst, want imageType) bool {
	return imageTypeSaveSupport(want) &&
		(!vipsSupportAnimation(src) ||
			(dst != imageTypeUnknown && !vipsSupportAnimation(dst)) ||
			vipsSupportAnimation(want))
}

func extractMeta(img *vipsImage, baseAngle int, useOrientation bool) (int, int, int, bool) {
	width := img.Width()
	height := img.Height()

	angle := 0
	flip := false

	if useOrientation {
		orientation := img.Orientation()

		if orientation == 3 || orientation == 4 {
			angle = 180
		}
		if orientation == 5 || orientation == 6 {
			angle = 90
		}
		if orientation == 7 || orientation == 8 {
			angle = 270
		}
		if orientation == 2 || orientation == 4 || orientation == 5 || orientation == 7 {
			flip = true
		}
	}

	if (angle+baseAngle)%180 != 0 {
		width, height = height, width
	}

	return width, height, angle, flip
}

func calcScale(width, height int, po *processingOptions, imgtype imageType) float64 {
	var shrink float64

	srcW, srcH := float64(width), float64(height)
	dstW, dstH := float64(po.Width), float64(po.Height)

	if po.Width == 0 {
		dstW = srcW
	}

	if po.Height == 0 {
		dstH = srcH
	}

	if dstW == srcW && dstH == srcH {
		shrink = 1
	} else {
		wshrink := srcW / dstW
		hshrink := srcH / dstH

		rt := po.ResizingType

		if rt == resizeAuto {
			srcD := width - height
			dstD := po.Width - po.Height

			if (srcD >= 0 && dstD >= 0) || (srcD < 0 && dstD < 0) {
				rt = resizeFill
			} else {
				rt = resizeFit
			}
		}

		switch {
		case po.Width == 0:
			shrink = hshrink
		case po.Height == 0:
			shrink = wshrink
		case rt == resizeFit:
			shrink = math.Max(wshrink, hshrink)
		default:
			shrink = math.Min(wshrink, hshrink)
		}
	}

	if !po.Enlarge && shrink < 1 && imgtype != imageTypeSVG {
		shrink = 1
	}

	shrink /= po.Dpr

	if shrink > srcW {
		shrink = srcW
	}

	if shrink > srcH {
		shrink = srcH
	}

	return 1.0 / shrink
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

func canFitToBytes(imgtype imageType) bool {
	switch imgtype {
	case imageTypeJPEG, imageTypeWEBP, imageTypeAVIF, imageTypeTIFF:
		return true
	default:
		return false
	}
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

func calcCropSize(orig int, crop float64) int {
	switch {
	case crop == 0.0:
		return 0
	case crop >= 1.0:
		return int(crop)
	default:
		return maxInt(1, scaleInt(orig, crop))
	}
}

func calcPosition(width, height, innerWidth, innerHeight int, gravity *gravityOptions, allowOverflow bool) (left, top int) {
	if gravity.Type == gravityFocusPoint {
		pointX := scaleInt(width, gravity.X)
		pointY := scaleInt(height, gravity.Y)

		left = pointX - innerWidth/2
		top = pointY - innerHeight/2
	} else {
		offX, offY := int(gravity.X), int(gravity.Y)

		left = (width-innerWidth+1)/2 + offX
		top = (height-innerHeight+1)/2 + offY

		if gravity.Type == gravityNorth || gravity.Type == gravityNorthEast || gravity.Type == gravityNorthWest {
			top = 0 + offY
		}

		if gravity.Type == gravityEast || gravity.Type == gravityNorthEast || gravity.Type == gravitySouthEast {
			left = width - innerWidth - offX
		}

		if gravity.Type == gravitySouth || gravity.Type == gravitySouthEast || gravity.Type == gravitySouthWest {
			top = height - innerHeight - offY
		}

		if gravity.Type == gravityWest || gravity.Type == gravityNorthWest || gravity.Type == gravitySouthWest {
			left = 0 + offX
		}
	}

	var minX, maxX, minY, maxY int

	if allowOverflow {
		minX, maxX = -innerWidth+1, width-1
		minY, maxY = -innerHeight+1, height-1
	} else {
		minX, maxX = 0, width-innerWidth
		minY, maxY = 0, height-innerHeight
	}

	left = maxInt(minX, minInt(left, maxX))
	top = maxInt(minY, minInt(top, maxY))

	return
}

func cropImage(img *vipsImage, cropWidth, cropHeight int, gravity *gravityOptions) error {
	if cropWidth == 0 && cropHeight == 0 {
		return nil
	}

	imgWidth, imgHeight := img.Width(), img.Height()

	cropWidth = minNonZeroInt(cropWidth, imgWidth)
	cropHeight = minNonZeroInt(cropHeight, imgHeight)

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

	left, top := calcPosition(imgWidth, imgHeight, cropWidth, cropHeight, gravity, false)
	return img.Crop(left, top, cropWidth, cropHeight)
}

func prepareWatermark(wm *vipsImage, wmData *imageData, opts *watermarkOptions, imgWidth, imgHeight int) error {
	if err := wm.Load(wmData.Data, wmData.Type, 1, 1.0, 1); err != nil {
		return err
	}

	po := newProcessingOptions()
	po.ResizingType = resizeFit
	po.Dpr = 1
	po.Enlarge = true
	po.Format = wmData.Type

	if opts.Scale > 0 {
		po.Width = maxInt(scaleInt(imgWidth, opts.Scale), 1)
		po.Height = maxInt(scaleInt(imgHeight, opts.Scale), 1)
	}

	if err := transformImage(context.Background(), wm, wmData.Data, po, wmData.Type); err != nil {
		return err
	}

	if err := wm.EnsureAlpha(); err != nil {
		return nil
	}

	if opts.Replicate {
		return wm.Replicate(imgWidth, imgHeight)
	}

	left, top := calcPosition(imgWidth, imgHeight, wm.Width(), wm.Height(), &opts.Gravity, true)

	return wm.Embed(imgWidth, imgHeight, left, top, rgbColor{0, 0, 0}, true)
}

func applyWatermark(img *vipsImage, wmData *imageData, opts *watermarkOptions, framesCount int) error {
	if err := img.RgbColourspace(); err != nil {
		return err
	}

	if err := img.CopyMemory(); err != nil {
		return err
	}

	wm := new(vipsImage)
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

	opacity := opts.Opacity * conf.WatermarkOpacity

	return img.ApplyWatermark(wm, opacity)
}

func copyMemoryAndCheckTimeout(ctx context.Context, img *vipsImage) error {
	err := img.CopyMemory()
	checkTimeout(ctx)
	return err
}

func transformImage(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	var (
		err     error
		trimmed bool
	)

	if po.Trim.Enabled {
		if err = img.Trim(po.Trim.Threshold, po.Trim.Smart, po.Trim.Color, po.Trim.EqualHor, po.Trim.EqualVer); err != nil {
			return err
		}
		if err = copyMemoryAndCheckTimeout(ctx, img); err != nil {
			return err
		}
		trimmed = true
	}

	srcWidth, srcHeight, angle, flip := extractMeta(img, po.Rotate, po.AutoRotate)

	cropWidth := calcCropSize(srcWidth, po.Crop.Width)
	cropHeight := calcCropSize(srcHeight, po.Crop.Height)

	cropGravity := po.Crop.Gravity
	if cropGravity.Type == gravityUnknown {
		cropGravity = po.Gravity
	}

	widthToScale := minNonZeroInt(cropWidth, srcWidth)
	heightToScale := minNonZeroInt(cropHeight, srcHeight)

	scale := calcScale(widthToScale, heightToScale, po, imgtype)

	if cropWidth > 0 {
		cropWidth = maxInt(1, scaleInt(cropWidth, scale))
	}
	if cropHeight > 0 {
		cropHeight = maxInt(1, scaleInt(cropHeight, scale))
	}
	if cropGravity.Type != gravityFocusPoint {
		cropGravity.X *= scale
		cropGravity.Y *= scale
	}

	if !trimmed && scale != 1 && data != nil && canScaleOnLoad(imgtype, scale) {
		jpegShrink := calcJpegShink(scale, imgtype)

		if imgtype != imageTypeJPEG || jpegShrink != 1 {
			// Do some scale-on-load
			if err = img.Load(data, imgtype, jpegShrink, scale, 1); err != nil {
				return err
			}
		}

		// Update scale after scale-on-load
		newWidth, newHeight, _, _ := extractMeta(img, po.Rotate, po.AutoRotate)
		if srcWidth > srcHeight {
			scale = float64(srcWidth) * scale / float64(newWidth)
		} else {
			scale = float64(srcHeight) * scale / float64(newHeight)
		}
		if srcWidth == scaleInt(srcWidth, scale) && srcHeight == scaleInt(srcHeight, scale) {
			scale = 1.0
		}
	}

	if err = img.Rad2Float(); err != nil {
		return err
	}

	iccImported := false
	convertToLinear := conf.UseLinearColorspace && scale != 1

	if convertToLinear {
		if err = img.ImportColourProfile(); err != nil {
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

	if err = copyMemoryAndCheckTimeout(ctx, img); err != nil {
		return err
	}

	if err = img.Rotate(angle); err != nil {
		return err
	}

	if flip {
		if err = img.Flip(); err != nil {
			return err
		}
	}

	if err = img.Rotate(po.Rotate); err != nil {
		return err
	}

	dprWidth := scaleInt(po.Width, po.Dpr)
	dprHeight := scaleInt(po.Height, po.Dpr)

	if err = cropImage(img, cropWidth, cropHeight, &cropGravity); err != nil {
		return err
	}
	if err = cropImage(img, dprWidth, dprHeight, &po.Gravity); err != nil {
		return err
	}

	if po.Format == imageTypeWEBP {
		webpLimitShrink := float64(maxInt(img.Width(), img.Height())) / webpMaxDimension

		if webpLimitShrink > 1.0 {
			if err = img.Resize(1.0/webpLimitShrink, hasAlpha); err != nil {
				return err
			}
			logWarning("WebP dimension size is limited to %d. The image is rescaled to %dx%d", int(webpMaxDimension), img.Width(), img.Height())

			if err = copyMemoryAndCheckTimeout(ctx, img); err != nil {
				return err
			}
		}
	}

	keepProfile := !po.StripColorProfile && po.Format.SupportsColourProfile()

	if iccImported {
		if keepProfile {
			// We imported ICC profile and want to keep it,
			// so we need to export it
			if err = img.ExportColourProfile(); err != nil {
				return err
			}
		} else {
			// We imported ICC profile but don't want to keep it,
			// so we need to export image to sRGB for maximum compatibility
			if err = img.ExportColourProfileToSRGB(); err != nil {
				return err
			}
		}
	} else if !keepProfile {
		// We don't import ICC profile and don't want to keep it,
		// so we need to transform it to sRGB for maximum compatibility
		if err = img.TransformColourProfile(); err != nil {
			return err
		}
	}

	if err = img.RgbColourspace(); err != nil {
		return err
	}

	if !keepProfile {
		if err = img.RemoveColourProfile(); err != nil {
			return err
		}
	}

	transparentBg := po.Format.SupportsAlpha() && !po.Flatten

	if hasAlpha && !transparentBg {
		if err = img.Flatten(po.Background); err != nil {
			return err
		}
	}

	if err = copyMemoryAndCheckTimeout(ctx, img); err != nil {
		return err
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

	if err = copyMemoryAndCheckTimeout(ctx, img); err != nil {
		return err
	}

	if po.Extend.Enabled && (dprWidth > img.Width() || dprHeight > img.Height()) {
		offX, offY := calcPosition(dprWidth, dprHeight, img.Width(), img.Height(), &po.Extend.Gravity, false)
		if err = img.Embed(dprWidth, dprHeight, offX, offY, po.Background, transparentBg); err != nil {
			return err
		}
	}

	if po.Padding.Enabled {
		paddingTop := scaleInt(po.Padding.Top, po.Dpr)
		paddingRight := scaleInt(po.Padding.Right, po.Dpr)
		paddingBottom := scaleInt(po.Padding.Bottom, po.Dpr)
		paddingLeft := scaleInt(po.Padding.Left, po.Dpr)
		if err = img.Embed(
			img.Width()+paddingLeft+paddingRight,
			img.Height()+paddingTop+paddingBottom,
			paddingLeft,
			paddingTop,
			po.Background,
			transparentBg,
		); err != nil {
			return err
		}
	}

	if po.Watermark.Enabled && watermark != nil {
		if err = applyWatermark(img, watermark, &po.Watermark, 1); err != nil {
			return err
		}
	}

	if err = img.RgbColourspace(); err != nil {
		return err
	}

	if err := img.CastUchar(); err != nil {
		return err
	}

	if po.StripMetadata {
		if err := img.Strip(); err != nil {
			return err
		}
	}

	return copyMemoryAndCheckTimeout(ctx, img)
}

func transformAnimated(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	if po.Trim.Enabled {
		logWarning("Trim is not supported for animated images")
		po.Trim.Enabled = false
	}

	imgWidth := img.Width()

	frameHeight, err := img.GetInt("page-height")
	if err != nil {
		return err
	}

	framesCount := minInt(img.Height()/frameHeight, conf.MaxAnimationFrames)

	// Double check dimensions because animated image has many frames
	if err = checkDimensions(imgWidth, frameHeight*framesCount); err != nil {
		return err
	}

	// Vips 8.8+ supports n-pages and doesn't load the whole animated image on header access
	if nPages, _ := img.GetIntDefault("n-pages", 0); nPages > framesCount {
		// Load only the needed frames
		if err = img.Load(data, imgtype, 1, 1.0, framesCount); err != nil {
			return err
		}
	}

	delay, err := img.GetIntSliceDefault("delay", nil)
	if err != nil {
		return err
	}

	loop, err := img.GetIntDefault("loop", 0)
	if err != nil {
		return err
	}

	// Legacy fields
	// TODO: remove this in major update
	gifLoop, err := img.GetIntDefault("gif-loop", -1)
	if err != nil {
		return err
	}
	gifDelay, err := img.GetIntDefault("gif-delay", -1)
	if err != nil {
		return err
	}

	watermarkEnabled := po.Watermark.Enabled
	po.Watermark.Enabled = false
	defer func() { po.Watermark.Enabled = watermarkEnabled }()

	frames := make([]*vipsImage, framesCount)
	defer func() {
		for _, frame := range frames {
			if frame != nil {
				frame.Clear()
			}
		}
	}()

	for i := 0; i < framesCount; i++ {
		frame := new(vipsImage)

		if err = img.Extract(frame, 0, i*frameHeight, imgWidth, frameHeight); err != nil {
			return err
		}

		frames[i] = frame

		if err = transformImage(ctx, frame, nil, po, imgtype); err != nil {
			return err
		}

		if err = copyMemoryAndCheckTimeout(ctx, frame); err != nil {
			return err
		}
	}

	if err = img.Arrayjoin(frames); err != nil {
		return err
	}

	if watermarkEnabled && watermark != nil {
		if err = applyWatermark(img, watermark, &po.Watermark, framesCount); err != nil {
			return err
		}
	}

	if err = img.CastUchar(); err != nil {
		return err
	}

	if err = copyMemoryAndCheckTimeout(ctx, img); err != nil {
		return err
	}

	if len(delay) == 0 {
		delay = make([]int, framesCount)
		for i := range delay {
			delay[i] = 40
		}
	} else if len(delay) > framesCount {
		delay = delay[:framesCount]
	}

	img.SetInt("page-height", frames[0].Height())
	img.SetIntSlice("delay", delay)
	img.SetInt("loop", loop)
	img.SetInt("n-pages", framesCount)

	// Legacy fields
	// TODO: remove this in major update
	if gifLoop >= 0 {
		img.SetInt("gif-loop", gifLoop)
	}
	if gifDelay >= 0 {
		img.SetInt("gif-delay", gifDelay)
	}

	return nil
}

func getIcoData(imgdata *imageData) (*imageData, error) {
	icoMeta, err := imagemeta.DecodeIcoMeta(bytes.NewReader(imgdata.Data))
	if err != nil {
		return nil, err
	}

	offset := icoMeta.BestImageOffset()
	size := icoMeta.BestImageSize()

	data := imgdata.Data[offset : offset+size]

	var format string

	meta, err := imagemeta.DecodeMeta(bytes.NewReader(data))
	if err != nil {
		// Looks like it's BMP with an incomplete header
		if d, err := imagemeta.FixBmpHeader(data); err == nil {
			format = "bmp"
			data = d
		} else {
			return nil, err
		}
	} else {
		format = meta.Format()
	}

	if imgtype, ok := imageTypes[format]; ok && vipsTypeSupportLoad[imgtype] {
		return &imageData{
			Data: data,
			Type: imgtype,
		}, nil
	}

	return nil, fmt.Errorf("Can't load %s from ICO", meta.Format())
}

func saveImageToFitBytes(ctx context.Context, po *processingOptions, img *vipsImage) ([]byte, context.CancelFunc, error) {
	var diff float64
	quality := po.getQuality()

	for {
		result, cancel, err := img.Save(po.Format, quality)
		if len(result) <= po.MaxBytes || quality <= 10 || err != nil {
			return result, cancel, err
		}
		cancel()

		checkTimeout(ctx)

		delta := float64(len(result)) / float64(po.MaxBytes)
		switch {
		case delta > 3:
			diff = 0.25
		case delta > 1.5:
			diff = 0.5
		default:
			diff = 0.75
		}
		quality = int(float64(quality) * diff)
	}
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
	imgdata := getImageData(ctx)

	switch {
	case po.Format == imageTypeUnknown:
		switch {
		case po.PreferAvif && canSwitchFormat(imgdata.Type, imageTypeUnknown, imageTypeAVIF):
			po.Format = imageTypeAVIF
		case po.PreferWebP && canSwitchFormat(imgdata.Type, imageTypeUnknown, imageTypeWEBP):
			po.Format = imageTypeWEBP
		case imageTypeSaveSupport(imgdata.Type) && imageTypeGoodForWeb(imgdata.Type):
			po.Format = imgdata.Type
		default:
			po.Format = imageTypeJPEG
		}
	case po.EnforceAvif && canSwitchFormat(imgdata.Type, po.Format, imageTypeAVIF):
		po.Format = imageTypeAVIF
	case po.EnforceWebP && canSwitchFormat(imgdata.Type, po.Format, imageTypeWEBP):
		po.Format = imageTypeWEBP
	}

	if po.Format == imageTypeSVG {
		if imgdata.Type != imageTypeSVG {
			return []byte{}, func() {}, errConvertingNonSvgToSvg
		}

		return imgdata.Data, func() {}, nil
	}

	if imgdata.Type == imageTypeSVG && !vipsTypeSupportLoad[imageTypeSVG] {
		return []byte{}, func() {}, errSourceImageTypeNotSupported
	}

	if imgdata.Type == imageTypeICO {
		icodata, err := getIcoData(imgdata)
		if err != nil {
			return nil, func() {}, err
		}

		imgdata = icodata
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

	if po.ResizingType == resizeCrop {
		logWarning("`crop` resizing type is deprecated and will be removed in future versions. Use `crop` processing option instead")

		po.Crop.Width, po.Crop.Height = float64(po.Width), float64(po.Height)

		po.ResizingType = resizeFit
		po.Width, po.Height = 0, 0
	}

	animationSupport := conf.MaxAnimationFrames > 1 && vipsSupportAnimation(imgdata.Type) && vipsSupportAnimation(po.Format)

	pages := 1
	if animationSupport {
		pages = -1
	}

	img := new(vipsImage)
	defer img.Clear()

	if err := img.Load(imgdata.Data, imgdata.Type, 1, 1.0, pages); err != nil {
		return nil, func() {}, err
	}

	if animationSupport && img.IsAnimated() {
		if err := transformAnimated(ctx, img, imgdata.Data, po, imgdata.Type); err != nil {
			return nil, func() {}, err
		}
	} else {
		if err := transformImage(ctx, img, imgdata.Data, po, imgdata.Type); err != nil {
			return nil, func() {}, err
		}
	}

	if err := copyMemoryAndCheckTimeout(ctx, img); err != nil {
		return nil, func() {}, err
	}

	if po.MaxBytes > 0 && canFitToBytes(po.Format) {
		return saveImageToFitBytes(ctx, po, img)
	}

	return img.Save(po.Format, po.getQuality())
}
