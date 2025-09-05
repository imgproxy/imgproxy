package processing

import (
	"context"
	"errors"
	"runtime"
	"slices"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/svg"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// The main processing pipeline (without finalization).
// Applied to non-animated images and individual frames of animated images.
var mainPipeline = pipeline{
	vectorGuardScale,
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

// The finalization pipeline.
// Applied right before saving the image.
var finalizePipeline = pipeline{
	colorspaceToResult,
	stripMetadata,
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
		return errors.New("no supported preferred formats specified")
	}

	config.PreferredFormats = filtered

	return nil
}

// Result holds the result of image processing.
type Result struct {
	OutData      imagedata.ImageData
	OriginWidth  int
	OriginHeight int
	ResultWidth  int
	ResultHeight int
}

// ProcessImage processes the image according to the provided processing options
// and returns a [Result] that includes the processed image data and dimensions.
//
// The provided processing options may be modified during processing.
func ProcessImage(
	ctx context.Context,
	imgdata imagedata.ImageData,
	po *options.ProcessingOptions,
	watermarkProvider auximageprovider.Provider,
	idf *imagedata.Factory,
) (*Result, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer vips.Cleanup()

	img := new(vips.Image)
	defer img.Clear()

	// Load a single page/frame of the image so we can analyze it
	// and decide how to process it further
	thumbnailLoaded, err := initialLoadImage(img, imgdata, po.EnforceThumbnail)
	if err != nil {
		return nil, err
	}

	// Let's check if we should skip standard processing
	if shouldSkipStandardProcessing(imgdata.Format(), po) {
		return skipStandardProcessing(img, imgdata, po)
	}

	// Check if we expect image to be processed as animated.
	// If MaxAnimationFrames is 1, we never process as animated since we can only
	// process a single frame.
	animated := po.SecurityOptions.MaxAnimationFrames > 1 &&
		img.IsAnimated()

	// Determine output format and check if it's supported.
	// The determined format is stored in po.Format.
	if err = determineOutputFormat(img, imgdata, po, animated); err != nil {
		return nil, err
	}

	// Now, as we know the output format, we know for sure if the image
	// should be processed as animated
	animated = animated && po.Format.SupportsAnimationSave()

	// Load required number of frames/pages for processing
	// and remove animation-related data if not animated.
	// Don't reload if we initially loaded a thumbnail.
	if !thumbnailLoaded {
		if err = reloadImageForProcessing(img, imgdata, po, animated); err != nil {
			return nil, err
		}
	}

	// Check image dimensions and number of frames for security reasons
	originWidth, originHeight, err := checkImageSize(img, imgdata.Format(), po.SecurityOptions)
	if err != nil {
		return nil, err
	}

	// Transform the image (resize, crop, etc)
	if err = transformImage(ctx, img, po, imgdata, animated, watermarkProvider); err != nil {
		return nil, err
	}

	// Finalize the image (colorspace conversion, metadata stripping, etc)
	if err = finalizePipeline.Run(ctx, img, po, imgdata, watermarkProvider); err != nil {
		return nil, err
	}

	outData, err := saveImage(ctx, img, po)
	if err != nil {
		return nil, err
	}

	resultWidth, resultHeight, _ := getImageSize(img)

	return &Result{
		OutData:      outData,
		OriginWidth:  originWidth,
		OriginHeight: originHeight,
		ResultWidth:  resultWidth,
		ResultHeight: resultHeight,
	}, nil
}

// initialLoadImage loads a single page/frame of the image.
// If the image format supports thumbnails and thumbnail loading is enforced,
// it tries to load the thumbnail first.
func initialLoadImage(
	img *vips.Image,
	imgdata imagedata.ImageData,
	enforceThumbnail bool,
) (bool, error) {
	if enforceThumbnail && imgdata.Format().SupportsThumbnail() {
		if err := img.LoadThumbnail(imgdata); err == nil {
			return true, nil
		} else {
			log.Debugf("Can't load thumbnail: %s", err)
		}
	}

	return false, img.Load(imgdata, 1, 1.0, 1)
}

// reloadImageForProcessing reloads the image for processing.
// For animated images, it loads all frames up to MaxAnimationFrames.
func reloadImageForProcessing(
	img *vips.Image,
	imgdata imagedata.ImageData,
	po *options.ProcessingOptions,
	asAnimated bool,
) error {
	// If we are going to process the image as animated, we need to load all frames
	// up to MaxAnimationFrames
	if asAnimated {
		frames := min(img.Pages(), po.SecurityOptions.MaxAnimationFrames)
		return img.Load(imgdata, 1, 1.0, frames)
	}

	// Otherwise, we just need to remove any animation-related data
	return img.RemoveAnimation()
}

// checkImageSize checks the image dimensions and number of frames against security options.
// It returns the image width, height and a security check error, if any.
func checkImageSize(
	img *vips.Image,
	imgtype imagetype.Type,
	secops security.Options,
) (int, int, error) {
	width, height, frames := getImageSize(img)

	if imgtype.IsVector() {
		// We don't check vector image dimensions as we can render it in any size
		return width, height, nil
	}

	err := security.CheckDimensions(width, height, frames, secops)

	return width, height, err
}

// getImageSize returns the width and height of the image, taking into account
// orientation and animation.
func getImageSize(img *vips.Image) (int, int, int) {
	width, height := img.Width(), img.Height()
	frames := 1

	if img.IsAnimated() {
		// Animated images contain multiple frames, and libvips loads them stacked vertically.
		// We want to return the size of a single frame
		height = img.PageHeight()
		frames = img.PagesLoaded()
	}

	// If the image is rotated by 90 or 270 degrees, we need to swap width and height
	orientation := img.Orientation()
	if orientation == 5 || orientation == 6 || orientation == 7 || orientation == 8 {
		width, height = height, width
	}

	return width, height, frames
}

// Returns true if image should not be processed as usual
func shouldSkipStandardProcessing(inFormat imagetype.Type, po *options.ProcessingOptions) bool {
	outFormat := po.Format
	skipProcessingFormatEnabled := slices.Contains(po.SkipProcessingFormats, inFormat)

	if inFormat == imagetype.SVG {
		isOutUnknown := outFormat == imagetype.Unknown

		switch {
		case outFormat == imagetype.SVG:
			return true
		case isOutUnknown && !config.AlwaysRasterizeSvg:
			return true
		case isOutUnknown && config.AlwaysRasterizeSvg && skipProcessingFormatEnabled:
			return true
		default:
			return false
		}
	} else {
		return skipProcessingFormatEnabled && (inFormat == outFormat || outFormat == imagetype.Unknown)
	}
}

// skipStandardProcessing skips standard image processing and returns the original image data.
//
// SVG images may be sanitized if configured to do so.
func skipStandardProcessing(
	img *vips.Image,
	imgdata imagedata.ImageData,
	po *options.ProcessingOptions,
) (*Result, error) {
	// Even if we skip standard processing, we still need to check image dimensions
	// to not send an image bomb to the client
	originWidth, originHeight, err := checkImageSize(img, imgdata.Format(), po.SecurityOptions)
	if err != nil {
		return nil, err
	}

	// Even in this case, SVG is an exception
	if imgdata.Format() == imagetype.SVG && config.SanitizeSvg {
		sanitized, err := svg.Sanitize(imgdata)
		if err != nil {
			return nil, err
		}

		return &Result{
			OutData:      sanitized,
			OriginWidth:  originWidth,
			OriginHeight: originHeight,
			ResultWidth:  originWidth,
			ResultHeight: originHeight,
		}, nil
	}

	// Return the original image
	return &Result{
		OutData:      imgdata,
		OriginWidth:  originWidth,
		OriginHeight: originHeight,
		ResultWidth:  originWidth,
		ResultHeight: originHeight,
	}, nil
}

// determineOutputFormat determines the output image format based on the processing options
// and image properties.
//
// It modifies the ProcessingOptions in place to set the output format.
func determineOutputFormat(
	img *vips.Image,
	imgdata imagedata.ImageData,
	po *options.ProcessingOptions,
	animated bool,
) error {
	// Check if the image may have transparency
	expectTransparency := !po.Flatten &&
		(img.HasAlpha() || po.Padding.Enabled || po.Extend.Enabled)

	switch {
	case po.Format == imagetype.SVG:
		// At this point we can't allow requested format to be SVG as we can't save SVGs
		return newSaveFormatError(po.Format)
	case po.Format == imagetype.Unknown:
		switch {
		case po.PreferJxl && !animated:
			po.Format = imagetype.JXL
		case po.PreferAvif && !animated:
			po.Format = imagetype.AVIF
		case po.PreferWebP:
			po.Format = imagetype.WEBP
		case isImageTypePreferred(imgdata.Format()):
			po.Format = imgdata.Format()
		default:
			po.Format = findPreferredFormat(animated, expectTransparency)
		}
	case po.EnforceJxl && !animated:
		po.Format = imagetype.JXL
	case po.EnforceAvif && !animated:
		po.Format = imagetype.AVIF
	case po.EnforceWebP:
		po.Format = imagetype.WEBP
	}

	if !vips.SupportsSave(po.Format) {
		return newSaveFormatError(po.Format)
	}

	return nil
}

// isImageTypePreferred checks if the given image type is in the list of preferred formats.
func isImageTypePreferred(imgtype imagetype.Type) bool {
	return slices.Contains(config.PreferredFormats, imgtype)
}

// isImageTypeCompatible checks if the given image type is compatible with the image properties.
func isImageTypeCompatible(imgtype imagetype.Type, animated, expectTransparency bool) bool {
	if animated && !imgtype.SupportsAnimationSave() {
		return false
	}

	if expectTransparency && !imgtype.SupportsAlpha() {
		return false
	}

	return true
}

// findPreferredFormat finds a suitable preferred format based on image's properties.
func findPreferredFormat(animated, expectTransparency bool) imagetype.Type {
	for _, t := range config.PreferredFormats {
		if isImageTypeCompatible(t, animated, expectTransparency) {
			return t
		}
	}

	return config.PreferredFormats[0]
}

func transformImage(
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
	imgdata imagedata.ImageData,
	asAnimated bool,
	watermark auximageprovider.Provider,
) error {
	if asAnimated {
		return transformAnimated(ctx, img, po, watermark)
	}

	return mainPipeline.Run(ctx, img, po, imgdata, watermark)
}

func transformAnimated(
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
	watermark auximageprovider.Provider,
) error {
	if po.Trim.Enabled {
		log.Warning("Trim is not supported for animated images")
		po.Trim.Enabled = false
	}

	imgWidth := img.Width()
	frameHeight := img.PageHeight()
	framesCount := img.PagesLoaded()

	// Get frame delays. We'll need to set them back later.
	// If we don't have delay info, we'll set a default delay later.
	delay, err := img.GetIntSliceDefault("delay", nil)
	if err != nil {
		return err
	}

	// Get loop count. We'll need to set it back later.
	// 0 means infinite looping.
	loop, err := img.GetIntDefault("loop", 0)
	if err != nil {
		return err
	}

	// Disable watermarking for individual frames.
	// It's more efficient to apply watermark to all frames at once after they are processed.
	watermarkEnabled := po.Watermark.Enabled
	po.Watermark.Enabled = false
	defer func() { po.Watermark.Enabled = watermarkEnabled }()

	// Make a slice to hold processed frames and ensure they are cleared on function exit
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

		// Extract an individual frame from the image.
		// Libvips loads animated images as a single image with frames stacked vertically.
		if err = img.Extract(frame, 0, i*frameHeight, imgWidth, frameHeight); err != nil {
			return err
		}

		frames = append(frames, frame)

		// Transform the frame using the main pipeline.
		// We don't provide imgdata here to prevent scale-on-load.
		// Let's skip passing watermark here since in would be applied later to all frames at once.
		if err = mainPipeline.Run(ctx, frame, po, nil, nil); err != nil {
			return err
		}

		// If the frame was scaled down, it's better to copy it to RAM
		// to speed up further processing.
		if r, _ := frame.GetIntDefault("imgproxy-scaled-down", 0); r == 1 {
			if err = frame.CopyMemory(); err != nil {
				return err
			}

			if err = server.CheckTimeout(ctx); err != nil {
				return err
			}
		}
	}

	// Join processed frames back into a single image.
	if err = img.Arrayjoin(frames); err != nil {
		return err
	}

	// Apply watermark to all frames at once if it was requested.
	// This is much more efficient than applying watermark to individual frames.
	if watermarkEnabled && watermark != nil {
		// Get DPR scale to apply watermark correctly on HiDPI images.
		// `imgproxy-dpr-scale` is set by the pipeline.
		dprScale, derr := img.GetDoubleDefault("imgproxy-dpr-scale", 1.0)
		if derr != nil {
			dprScale = 1.0
		}

		if err = applyWatermark(ctx, img, watermark, po, dprScale, framesCount); err != nil {
			return err
		}
	}

	if len(delay) == 0 {
		// if we don't have delay info, set it to 40ms for each frame (25 FPS).
		delay = make([]int, framesCount)
		for i := range delay {
			delay[i] = 40
		}
	} else if len(delay) > framesCount {
		// if we have more delay entries than frames, truncate it.
		delay = delay[:framesCount]
	}

	// Mark the image as animated so img.Strip() doesn't remove animation data.
	img.SetInt("imgproxy-is-animated", 1)
	// Set animation data back.
	img.SetInt("page-height", frames[0].Height())
	img.SetIntSlice("delay", delay)
	img.SetInt("loop", loop)
	img.SetInt("n-pages", img.Height()/frames[0].Height())

	return nil
}

func saveImage(
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
) (imagedata.ImageData, error) {
	// AVIF has a minimal dimension of 16 pixels.
	// If one of the dimensions is less, we need to switch to another format.
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

	// If we want and can fit the image into the specified number of bytes,
	// let's do it.
	if po.MaxBytes > 0 && po.Format.SupportsQuality() {
		return saveImageToFitBytes(ctx, po, img)
	}

	// Otherwise, just save the image with the specified quality.
	return img.Save(po.Format, po.GetQuality())
}
