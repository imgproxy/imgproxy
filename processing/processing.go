package processing

import (
	"context"
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v2/config"
	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/imagetype"
	"github.com/imgproxy/imgproxy/v2/imath"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/router"
	"github.com/imgproxy/imgproxy/v2/security"
	"github.com/imgproxy/imgproxy/v2/vips"
)

var mainPipeline = pipeline{
	trim,
	prepare,
	scaleOnLoad,
	importColorProfile,
	scale,
	rotateAndFlip,
	crop,
	fixWebpSize,
	exportColorProfile,
	applyFilters,
	extend,
	padding,
	flatten,
	watermark,
	finalize,
}

func imageTypeGoodForWeb(imgtype imagetype.Type) bool {
	return imgtype != imagetype.TIFF &&
		imgtype != imagetype.BMP
}

// src  - the source image format
// dst  - what the user specified
// want - what we want switch to
func canSwitchFormat(src, dst, want imagetype.Type) bool {
	// If the format we want is not supported, we can't switch to it anyway
	return vips.SupportsSave(want) &&
		// if src format does't support animation, we can switch to whatever we want
		(!src.SupportsAnimation() ||
			// if user specified the format and it doesn't support animation, we can switch to whatever we want
			(dst != imagetype.Unknown && !dst.SupportsAnimation()) ||
			// if the format we want supports animation, we can switch in any case
			want.SupportsAnimation())
}

func canFitToBytes(imgtype imagetype.Type) bool {
	switch imgtype {
	case imagetype.JPEG, imagetype.WEBP, imagetype.AVIF, imagetype.TIFF:
		return true
	default:
		return false
	}
}

func transformAnimated(ctx context.Context, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if po.Trim.Enabled {
		log.Warning("Trim is not supported for animated images")
		po.Trim.Enabled = false
	}

	imgWidth := img.Width()

	frameHeight, err := img.GetInt("page-height")
	if err != nil {
		return err
	}

	framesCount := imath.Min(img.Height()/frameHeight, config.MaxAnimationFrames)

	// Double check dimensions because animated image has many frames
	if err = security.CheckDimensions(imgWidth, frameHeight*framesCount); err != nil {
		return err
	}

	// Vips 8.8+ supports n-pages and doesn't load the whole animated image on header access
	if nPages, _ := img.GetIntDefault("n-pages", 0); nPages > framesCount {
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

	frames := make([]*vips.Image, framesCount)
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

		frames[i] = frame

		if err = mainPipeline.Run(ctx, frame, po, nil); err != nil {
			return err
		}
	}

	if err = img.Arrayjoin(frames); err != nil {
		return err
	}

	if watermarkEnabled && imagedata.Watermark != nil {
		if err = applyWatermark(img, imagedata.Watermark, &po.Watermark, framesCount); err != nil {
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

func saveImageToFitBytes(ctx context.Context, po *options.ProcessingOptions, img *vips.Image) (*imagedata.ImageData, error) {
	var diff float64
	quality := po.GetQuality()

	for {
		imgdata, err := img.Save(po.Format, quality)
		if len(imgdata.Data) <= po.MaxBytes || quality <= 10 || err != nil {
			return imgdata, err
		}
		imgdata.Close()

		router.CheckTimeout(ctx)

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

	switch {
	case po.Format == imagetype.Unknown:
		switch {
		case po.PreferAvif && canSwitchFormat(imgdata.Type, imagetype.Unknown, imagetype.AVIF):
			po.Format = imagetype.AVIF
		case po.PreferWebP && canSwitchFormat(imgdata.Type, imagetype.Unknown, imagetype.WEBP):
			po.Format = imagetype.WEBP
		case vips.SupportsSave(imgdata.Type) && imageTypeGoodForWeb(imgdata.Type):
			po.Format = imgdata.Type
		default:
			po.Format = imagetype.JPEG
		}
	case po.EnforceAvif && canSwitchFormat(imgdata.Type, po.Format, imagetype.AVIF):
		po.Format = imagetype.AVIF
	case po.EnforceWebP && canSwitchFormat(imgdata.Type, po.Format, imagetype.WEBP):
		po.Format = imagetype.WEBP
	}

	if !vips.SupportsSave(po.Format) {
		return nil, fmt.Errorf("Can't save %s, probably not supported by your libvips", po.Format)
	}

	animationSupport := config.MaxAnimationFrames > 1 && imgdata.Type.SupportsAnimation() && po.Format.SupportsAnimation()

	pages := 1
	if animationSupport {
		pages = -1
	}

	img := new(vips.Image)
	defer img.Clear()

	if err := img.Load(imgdata, 1, 1.0, pages); err != nil {
		return nil, err
	}

	if animationSupport && img.IsAnimated() {
		if err := transformAnimated(ctx, img, po, imgdata); err != nil {
			return nil, err
		}
	} else {
		if err := mainPipeline.Run(ctx, img, po, imgdata); err != nil {
			return nil, err
		}
	}

	if err := copyMemoryAndCheckTimeout(ctx, img); err != nil {
		return nil, err
	}

	if po.MaxBytes > 0 && canFitToBytes(po.Format) {
		return saveImageToFitBytes(ctx, po, img)
	}

	return img.Save(po.Format, po.GetQuality())
}
