package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"runtime"

	"github.com/imgproxy/imgproxy/v2/imagemeta"
)

const msgSmartCropNotSupported = "Smart crop is not supported by used version of libvips"

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
	case imageTypeJPEG, imageTypeWEBP, imageTypeHEIC, imageTypeTIFF:
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

	return wm.Embed(imgWidth, imgHeight, left, top, rgbColor{0, 0, 0})
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

func transformImage(ctx context.Context, img *vipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	var (
		err     error
		trimmed bool
	)

	if po.Trim.Enabled {
		if err = img.Trim(po.Trim.Threshold, po.Trim.Smart, po.Trim.Color, po.Trim.EqualHor, po.Trim.EqualVer); err != nil {
			return err
		}
		trimmed = true
	}

	srcWidth, srcHeight, angle, flip := extractMeta(img)
	cropWidth, cropHeight := po.Crop.Width, po.Crop.Height

	cropGravity := po.Crop.Gravity
	if cropGravity.Type == gravityUnknown {
		cropGravity = po.Gravity
	}

	widthToScale := minNonZeroInt(cropWidth, srcWidth)
	heightToScale := minNonZeroInt(cropHeight, srcHeight)

	scale := calcScale(widthToScale, heightToScale, po, imgtype)

	cropWidth = scaleInt(cropWidth, scale)
	cropHeight = scaleInt(cropHeight, scale)
	if cropGravity.Type != gravityFocusPoint {
		cropGravity.X *= scale
		cropGravity.Y *= scale
	}

	if !trimmed && scale != 1 && data != nil && canScaleOnLoad(imgtype, scale) {
		if imgtype == imageTypeWEBP || imgtype == imageTypeSVG {
			// Do some scale-on-load
			if err = img.Load(data, imgtype, 1, scale, 1); err != nil {
				return err
			}
		} else if imgtype == imageTypeJPEG {
			// Do some shrink-on-load
			if shrink := calcJpegShink(scale, imgtype); shrink != 1 {
				if err = img.Load(data, imgtype, shrink, 1.0, 1); err != nil {
					return err
				}
			}
		}

		// Update scale after scale-on-load
		newWidth, newHeight, _, _ := extractMeta(img)

		widthToScale = scaleInt(widthToScale, float64(newWidth)/float64(srcWidth))
		heightToScale = scaleInt(heightToScale, float64(newHeight)/float64(srcHeight))

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

	if err = img.CopyMemory(); err != nil {
		return err
	}

	checkTimeout(ctx)

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

	checkTimeout(ctx)

	dprWidth := scaleInt(po.Width, po.Dpr)
	dprHeight := scaleInt(po.Height, po.Dpr)

	if err = cropImage(img, cropWidth, cropHeight, &cropGravity); err != nil {
		return err
	}
	if err = cropImage(img, dprWidth, dprHeight, &po.Gravity); err != nil {
		return err
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

	if err = img.CopyMemory(); err != nil {
		return err
	}

	checkTimeout(ctx)

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

	if po.Extend.Enabled && (po.Width > img.Width() || po.Height > img.Height()) {
		offX, offY := calcPosition(po.Width, po.Height, img.Width(), img.Height(), &po.Extend.Gravity, false)
		if err = img.Embed(po.Width, po.Height, offX, offY, po.Background); err != nil {
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
		); err != nil {
			return err
		}
	}

	checkTimeout(ctx)

	if po.Watermark.Enabled && watermark != nil {
		if err = applyWatermark(img, watermark, &po.Watermark, 1); err != nil {
			return err
		}
	}

	return img.RgbColourspace()
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
	if nPages, _ := img.GetInt("n-pages"); nPages > 0 {
		scale := 1.0

		// Don't do scale on load if we need to crop
		if po.Crop.Width == 0 && po.Crop.Height == 0 {
			scale = calcScale(imgWidth, frameHeight, po, imgtype)
		}

		if nPages > framesCount || canScaleOnLoad(imgtype, scale) {
			logNotice("Animated scale on load")
			// Do some scale-on-load and load only the needed frames
			if err = img.Load(data, imgtype, 1, scale, framesCount); err != nil {
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

	watermarkEnabled := po.Watermark.Enabled
	po.Watermark.Enabled = false
	defer func() { po.Watermark.Enabled = watermarkEnabled }()

	var errg panicGroup

	frames := make([]*vipsImage, framesCount)
	defer func() {
		// Ensure all frames are processed before freeing
		errg.Wait()

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

		frame.CopyMemory()

		frames[i] = frame

		errg.Go(func() error {
			return transformImage(ctx, frame, nil, po, imgtype)
		})
	}

	if err = errg.Wait(); err != nil {
		return err
	}

	checkTimeout(ctx)

	if err = img.Arrayjoin(frames); err != nil {
		return err
	}

	if watermarkEnabled && watermark != nil {
		if err = applyWatermark(img, watermark, &po.Watermark, framesCount); err != nil {
			return err
		}
	}

	img.SetInt("page-height", frames[0].Height())
	img.SetInt("gif-delay", delay)
	img.SetInt("gif-loop", loop)
	img.SetInt("n-pages", framesCount)

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

func saveImageToFitBytes(po *processingOptions, img *vipsImage) ([]byte, context.CancelFunc, error) {
	var diff float64
	quality := po.Quality

	img.CopyMemory()

	for {
		result, cancel, err := img.Save(po.Format, quality, conf.StripMetadata)
		if len(result) <= po.MaxBytes || quality <= 10 || err != nil {
			return result, cancel, err
		}
		cancel()

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

	if po.Format == imageTypeUnknown {
		switch {
		case po.PreferWebP && imageTypeSaveSupport(imageTypeWEBP):
			po.Format = imageTypeWEBP
		case imageTypeSaveSupport(imgdata.Type) && imageTypeGoodForWeb(imgdata.Type):
			po.Format = imgdata.Type
		default:
			po.Format = imageTypeJPEG
		}
	} else if po.EnforceWebP && imageTypeSaveSupport(imageTypeWEBP) {
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

		po.Crop.Width, po.Crop.Height = po.Width, po.Height

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

	checkTimeout(ctx)

	if po.Format == imageTypeGIF {
		if err := img.CastUchar(); err != nil {
			return nil, func() {}, err
		}
		checkTimeout(ctx)
	}

	if po.MaxBytes > 0 && canFitToBytes(po.Format) {
		return saveImageToFitBytes(po, img)
	}

	return img.Save(po.Format, po.Quality, conf.StripMetadata)
}
