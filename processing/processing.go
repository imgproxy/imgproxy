package processing

import (
	"context"
	"errors"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var mainPipeline = pipeline{
	trim,
	prepare,
	scaleOnLoad,
	colorspaceToProcessing,
	crop,
	scale,
	rotateAndFlip,
	cropToResult,
	applyFilters,
	extend,
	extendAspectRatio,
	padding,
	fixSize,
	flatten,
	watermark,
}

var finalizePipeline = pipeline{
	colorspaceToResult,
	stripMetadata,
}

func isImageTypePreferred(imgtype imagetype.Type) bool {
	for _, t := range config.PreferredFormats {
		if imgtype == t {
			return true
		}
	}

	return false
}

func findBestFormat(srcType imagetype.Type, animated, expectAlpha bool) imagetype.Type {
	for _, t := range config.PreferredFormats {
		if animated && !t.SupportsAnimationSave() {
			continue
		}

		if expectAlpha && !t.SupportsAlpha() {
			continue
		}

		return t
	}

	return config.PreferredFormats[0]
}

func ValidatePreferredFormats() error {
	filtered := config.PreferredFormats[:0]

	for _, t := range config.PreferredFormats {
		if !vips.SupportsSave(t) {
			log.Warnf("%s can't be a preferred format as it's saving is not supported", t)
		} else {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return errors.New("No supported preferred formats specified")
	}

	config.PreferredFormats = filtered

	return nil
}

func getImageSize(img *vips.Image) (int, int) {
	width, height := img.Width(), img.Height()

	if img.IsAnimated() {
		// Animated images contain multiple frames, and libvips loads them stacked vertically.
		// We want to return the size of a single frame
		height = img.PageHeight()
	}

	// If the image is rotated by 90 or 270 degrees, we need to swap width and height
	orientation := img.Orientation()
	if orientation == 5 || orientation == 6 || orientation == 7 || orientation == 8 {
		width, height = height, width
	}

	return width, height
}

func transformAnimated(ctx context.Context, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if po.Trim.Enabled {
		log.Warning("Trim is not supported for animated images")
		po.Trim.Enabled = false
	}

	imgWidth := img.Width()
	framesCount := imath.Min(img.Pages(), po.SecurityOptions.MaxAnimationFrames)

	frameHeight, err := img.GetInt("page-height")
	if err != nil {
		return err
	}

	// Double check dimensions because animated image has many frames
	if err = security.CheckDimensions(imgWidth, frameHeight, framesCount, po.SecurityOptions); err != nil {
		return err
	}

	if img.Pages() > framesCount {
		// Load only the needed frames
		if err = img.Load(imgdata, 1, 1.0, framesCount); err != nil {
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

	watermarkEnabled := po.Watermark.Enabled
	po.Watermark.Enabled = false
	defer func() { po.Watermark.Enabled = watermarkEnabled }()

	frames := make([]*vips.Image, 0, framesCount)
	defer func() {
		for _, frame := range frames {
			if frame != nil {
				frame.Clear()
			}
		}
	}()

	for i := 0; i < framesCount; i++ {
		frame := new(vips.Image)

		if err = img.Extract(frame, 0, i*frameHeight, imgWidth, frameHeight); err != nil {
			return err
		}

		frames = append(frames, frame)

		if err = mainPipeline.Run(ctx, frame, po, nil); err != nil {
			return err
		}

		if r, _ := frame.GetIntDefault("imgproxy-scaled-down", 0); r == 1 {
			if err = frame.CopyMemory(); err != nil {
				return err
			}

			if err = router.CheckTimeout(ctx); err != nil {
				return err
			}
		}
	}

	if err = img.Arrayjoin(frames); err != nil {
		return err
	}

	if watermarkEnabled && imagedata.Watermark != nil {
		dprScale, derr := img.GetDoubleDefault("imgproxy-dpr-scale", 1.0)
		if derr != nil {
			dprScale = 1.0
		}

		if err = applyWatermark(img, imagedata.Watermark, &po.Watermark, dprScale, framesCount); err != nil {
			return err
		}
	}

	if err = img.CastUchar(); err != nil {
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

	img.SetInt("imgproxy-is-animated", 1)
	img.SetInt("page-height", frames[0].Height())
	img.SetIntSlice("delay", delay)
	img.SetInt("loop", loop)
	img.SetInt("n-pages", img.Height()/frames[0].Height())

	return nil
}

func saveImageToFitBytes(ctx context.Context, po *options.ProcessingOptions, img *vips.Image) (*imagedata.ImageData, error) {
	var diff float64
	quality := po.GetQuality()

	if err := img.CopyMemory(); err != nil {
		return nil, err
	}

	for {
		imgdata, err := img.Save(po.Format, quality)
		if err != nil || len(imgdata.Data) <= po.MaxBytes || quality <= 10 {
			return imgdata, err
		}
		imgdata.Close()

		if err := router.CheckTimeout(ctx); err != nil {
			return nil, err
		}

		delta := float64(len(imgdata.Data)) / float64(po.MaxBytes)
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

func ProcessImage(ctx context.Context, imgdata *imagedata.ImageData, po *options.ProcessingOptions) (*imagedata.ImageData, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer vips.Cleanup()

	animationSupport :=
		po.SecurityOptions.MaxAnimationFrames > 1 &&
			imgdata.Type.SupportsAnimationLoad() &&
			(po.Format == imagetype.Unknown || po.Format.SupportsAnimationSave())

	pages := 1
	if animationSupport {
		pages = -1
	}

	img := new(vips.Image)
	defer img.Clear()

	if po.EnforceThumbnail && imgdata.Type.SupportsThumbnail() {
		if err := img.LoadThumbnail(imgdata); err != nil {
			log.Debugf("Can't load thumbnail: %s", err)
			// Failed to load thumbnail, rollback to the full image
			if err := img.Load(imgdata, 1, 1.0, pages); err != nil {
				return nil, err
			}
		}
	} else {
		if err := img.Load(imgdata, 1, 1.0, pages); err != nil {
			return nil, err
		}
	}

	originWidth, originHeight := getImageSize(img)

	animated := img.IsAnimated()

	if po.Page >= 0 && animated {
		if po.Page >= img.Pages() {
			po.Page = img.Pages() - 1
		}

		frameHeight := img.PageHeight()
		imgWidth := img.Width()

		frame := new(vips.Image)
		if err := img.Extract(frame, 0, po.Page*frameHeight, imgWidth, frameHeight); err != nil {
			return nil, err
		}

		// Swap extracted frame into `img` so deferred Clear() will free it,
		// and clear the temporary holder which now contains the original image.
		img.Swap(frame)
		frame.Clear()
		animated = false
	}

	expectAlpha := !po.Flatten && (img.HasAlpha() || po.Padding.Enabled || po.Extend.Enabled)

	switch {
	case po.Format == imagetype.Unknown:
		switch {
		case po.PreferJxl && !animated:
			po.Format = imagetype.JXL
		case po.PreferAvif && !animated:
			po.Format = imagetype.AVIF
		case po.PreferWebP:
			po.Format = imagetype.WEBP
		case isImageTypePreferred(imgdata.Type):
			po.Format = imgdata.Type
		default:
			po.Format = findBestFormat(imgdata.Type, animated, expectAlpha)
		}
	case po.EnforceJxl && !animated:
		po.Format = imagetype.JXL
	case po.EnforceAvif && !animated:
		po.Format = imagetype.AVIF
	case po.EnforceWebP:
		po.Format = imagetype.WEBP
	}

	if !vips.SupportsSave(po.Format) {
		return nil, newSaveFormatError(po.Format)
	}

	if po.Format.SupportsAnimationSave() && animated {
		if err := transformAnimated(ctx, img, po, imgdata); err != nil {
			return nil, err
		}
	} else {
		if animated {
			// We loaded animated image but the resulting format doesn't support
			// animations, so we need to reload image as not animated
			if err := img.Load(imgdata, 1, 1.0, 1); err != nil {
				return nil, err
			}
		}

		if err := mainPipeline.Run(ctx, img, po, imgdata); err != nil {
			return nil, err
		}
	}

	if err := finalizePipeline.Run(ctx, img, po, imgdata); err != nil {
		return nil, err
	}

	if po.Format == imagetype.AVIF && (img.Width() < 16 || img.Height() < 16) {
		if img.HasAlpha() {
			po.Format = imagetype.PNG
		} else {
			po.Format = imagetype.JPEG
		}

		log.Warningf(
			"Minimal dimension of AVIF is 16, current image size is %dx%d. Image will be saved as %s",
			img.Width(), img.Height(), po.Format,
		)
	}

	var (
		outData *imagedata.ImageData
		err     error
	)

	if po.MaxBytes > 0 && po.Format.SupportsQuality() {
		outData, err = saveImageToFitBytes(ctx, po, img)
	} else {
		outData, err = img.Save(po.Format, po.GetQuality())
	}

	if err == nil {
		if outData.Headers == nil {
			outData.Headers = make(map[string]string)
		}
		resultWidth, resultHeight := getImageSize(img)
		outData.Headers["X-Origin-Width"] = strconv.Itoa(originWidth)
		outData.Headers["X-Origin-Height"] = strconv.Itoa(originHeight)
		outData.Headers["X-Result-Width"] = strconv.Itoa(resultWidth)
		outData.Headers["X-Result-Height"] = strconv.Itoa(resultHeight)
	}

	return outData, err
}
