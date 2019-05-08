package main

/*
#cgo pkg-config: vips
#cgo LDFLAGS: -s -w
#cgo CFLAGS: -O3
#include "vips.h"
*/
import "C"

import (
	"context"
	"errors"
	"math"
	"os"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sync/errgroup"
)

var (
	vipsSupportSmartcrop bool
	vipsTypeSupportLoad  = make(map[imageType]bool)
	vipsTypeSupportSave  = make(map[imageType]bool)

	watermark *C.VipsImage

	errSmartCropNotSupported = errors.New("Smart crop is not supported by used version of libvips")
)

type cConfig struct {
	JpegProgressive       C.int
	PngInterlaced         C.int
	PngQuantize           C.int
	PngQuantizationColors C.int
	WatermarkOpacity      C.double
}

var cConf cConfig

var cstrings = make(map[string]*C.char)

func initVips() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		logFatal("unable to start vips!")
	}

	// Disable libvips cache. Since processing pipeline is fine tuned, we won't get much profit from it.
	// Enabled cache can cause SIGSEGV on Musl-based systems like Alpine.
	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)

	C.vips_concurrency_set(1)

	if len(os.Getenv("IMGPROXY_VIPS_LEAK_CHECK")) > 0 {
		C.vips_leak_set(C.gboolean(1))
	}

	if len(os.Getenv("IMGPROXY_VIPS_CACHE_TRACE")) > 0 {
		C.vips_cache_set_trace(C.gboolean(1))
	}

	vipsSupportSmartcrop = C.vips_support_smartcrop() == 1

	if int(C.vips_type_find_load_go(C.int(imageTypeJPEG))) != 0 {
		vipsTypeSupportLoad[imageTypeJPEG] = true
	}
	if int(C.vips_type_find_load_go(C.int(imageTypePNG))) != 0 {
		vipsTypeSupportLoad[imageTypePNG] = true
	}
	if int(C.vips_type_find_load_go(C.int(imageTypeWEBP))) != 0 {
		vipsTypeSupportLoad[imageTypeWEBP] = true
	}
	if int(C.vips_type_find_load_go(C.int(imageTypeGIF))) != 0 {
		vipsTypeSupportLoad[imageTypeGIF] = true
	}
	if int(C.vips_type_find_load_go(C.int(imageTypeSVG))) != 0 {
		vipsTypeSupportLoad[imageTypeSVG] = true
	}

	// we load ICO with github.com/mat/besticon/ico and send decoded data to vips
	vipsTypeSupportLoad[imageTypeICO] = true

	if int(C.vips_type_find_save_go(C.int(imageTypeJPEG))) != 0 {
		vipsTypeSupportSave[imageTypeJPEG] = true
	}
	if int(C.vips_type_find_save_go(C.int(imageTypePNG))) != 0 {
		vipsTypeSupportSave[imageTypePNG] = true
	}
	if int(C.vips_type_find_save_go(C.int(imageTypeWEBP))) != 0 {
		vipsTypeSupportSave[imageTypeWEBP] = true
	}
	if int(C.vips_type_find_save_go(C.int(imageTypeGIF))) != 0 {
		vipsTypeSupportSave[imageTypeGIF] = true
	}
	if int(C.vips_type_find_save_go(C.int(imageTypeICO))) != 0 {
		vipsTypeSupportSave[imageTypeICO] = true
	}

	if conf.JpegProgressive {
		cConf.JpegProgressive = C.int(1)
	}

	if conf.PngInterlaced {
		cConf.PngInterlaced = C.int(1)
	}

	if conf.PngQuantize {
		cConf.PngQuantize = C.int(1)
	}

	cConf.PngQuantizationColors = C.int(conf.PngQuantizationColors)

	cConf.WatermarkOpacity = C.double(conf.WatermarkOpacity)

	if err := vipsPrepareWatermark(); err != nil {
		logFatal(err.Error())
	}

	collectVipsMetrics()
}

func shutdownVips() {
	C.clear_image(&watermark)
	C.vips_shutdown()
}

func collectVipsMetrics() {
	if prometheusEnabled {
		go func() {
			for range time.Tick(5 * time.Second) {
				prometheusVipsMemory.Set(float64(C.vips_tracked_get_mem()))
				prometheusVipsMaxMemory.Set(float64(C.vips_tracked_get_mem_highwater()))
				prometheusVipsAllocs.Set(float64(C.vips_tracked_get_allocs()))
			}
		}()
	}
}

func cachedCString(str string) *C.char {
	if cstr, ok := cstrings[str]; ok {
		return cstr
	}

	cstr := C.CString(str)
	cstrings[str] = cstr

	return cstr
}

func extractMeta(img *C.VipsImage) (int, int, int, bool) {
	width := int(img.Xsize)
	height := int(img.Ysize)

	angle := C.VIPS_ANGLE_D0
	flip := false

	orientation := C.vips_get_exif_orientation(img)

	if orientation >= 5 && orientation <= 8 {
		width, height = height, width
	}
	if orientation == 3 || orientation == 4 {
		angle = C.VIPS_ANGLE_D180
	}
	if orientation == 5 || orientation == 6 {
		angle = C.VIPS_ANGLE_D90
	}
	if orientation == 7 || orientation == 8 {
		angle = C.VIPS_ANGLE_D270
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

func calcShink(scale float64, imgtype imageType) int {
	switch imgtype {
	case imageTypeWEBP:
		return int(1.0 / scale)
	case imageTypeJPEG:
		shrink := int(1.0 / scale)

		switch {
		case shrink >= 8:
			return 8
		case shrink >= 4:
			return 4
		case shrink >= 2:
			return 2
		}
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

func transformImage(ctx context.Context, img **C.VipsImage, data []byte, po *processingOptions, imgtype imageType) error {
	var err error

	imgWidth, imgHeight, angle, flip := extractMeta(*img)

	hasAlpha := vipsImageHasAlpha(*img)

	scale := calcScale(imgWidth, imgHeight, po, imgtype)

	if scale != 1 && data != nil {
		if imgtype == imageTypeSVG {
			// Load SVG with desired scale
			if tmp, err := vipsLoadImage(data, imgtype, 1, scale, false); err == nil {
				C.swap_and_clear(img, tmp)
			} else {
				return err
			}

			scale = 1
		} else if !conf.DisableShrinkOnLoad && scale < 1.0 {
			// Do some shrink-on-load
			if shrink := calcShink(scale, imgtype); shrink != 1 {
				if tmp, err := vipsLoadImage(data, imgtype, shrink, 1.0, false); err == nil {
					C.swap_and_clear(img, tmp)
				} else {
					return err
				}

				scale = scale * float64(shrink)
			}
		}
	}

	if err = vipsRad2Float(img); err != nil {
		return err
	}

	convertToLinear := conf.UseLinearColorspace && (scale != 1 || po.Dpr != 1)

	if convertToLinear {
		if err = vipsImportColourProfile(img, true); err != nil {
			return err
		}

		if err = vipsLinearColourspace(img); err != nil {
			return err
		}
	}

	if scale != 1 {
		if err = vipsResize(img, scale, hasAlpha); err != nil {
			return err
		}
	}

	// Update actual image size after resize
	imgWidth, imgHeight, _, _ = extractMeta(*img)

	checkTimeout(ctx)

	if angle != C.VIPS_ANGLE_D0 || flip {
		if err = vipsImageCopyMemory(img); err != nil {
			return err
		}

		if angle != C.VIPS_ANGLE_D0 {
			if err = vipsRotate(img, angle); err != nil {
				return err
			}
		}

		if flip {
			if err = vipsFlip(img); err != nil {
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
			if err = vipsImageCopyMemory(img); err != nil {
				return err
			}
			if err = vipsSmartCrop(img, cropW, cropH); err != nil {
				return err
			}
			// Applying additional modifications after smart crop causes SIGSEGV on Alpine
			// so we have to copy memory after it
			if err = vipsImageCopyMemory(img); err != nil {
				return err
			}
		} else {
			left, top := calcCrop(imgWidth, imgHeight, cropW, cropH, &po.Gravity)
			if err = vipsCrop(img, left, top, cropW, cropH); err != nil {
				return err
			}
		}

		checkTimeout(ctx)
	}

	if po.Enlarge && po.Resize == resizeCrop && po.Dpr > 1 {
		// We didn't enlarge the image before, because is wasn't optimal. Now it's time to do it
		if err = vipsResize(img, po.Dpr, hasAlpha); err != nil {
			return err
		}
		if err = vipsImageCopyMemory(img); err != nil {
			return err
		}
	}

	if convertToLinear {
		if err = vipsFixColourspace(img); err != nil {
			return err
		}
	} else {
		if err = vipsImportColourProfile(img, false); err != nil {
			return err
		}
	}

	if po.Expand && (po.Width > int((*img).Xsize) || po.Height > int((*img).Ysize)) {
		if err = vipsEnsureAlpha(img); err != nil {
			return err
		}

		hasAlpha = true

		if err = vipsEmbed(img, gravityCenter, C.int(po.Width), C.int(po.Height), 0, 0); err != nil {
			return err
		}
	}

	if hasAlpha && (po.Flatten || po.Format == imageTypeJPEG) {
		if err = vipsFlatten(img, po.Background); err != nil {
			return err
		}
	}

	if po.Blur > 0 {
		if err = vipsBlur(img, po.Blur); err != nil {
			return err
		}
	}

	if po.Sharpen > 0 {
		if err = vipsSharpen(img, po.Sharpen); err != nil {
			return err
		}
	}

	checkTimeout(ctx)

	if po.Watermark.Enabled {
		if err = vipsApplyWatermark(img, &po.Watermark); err != nil {
			return err
		}
	}

	return vipsFixColourspace(img)
}

func transformGif(ctx context.Context, img **C.VipsImage, po *processingOptions) error {
	imgWidth := int((*img).Xsize)
	imgHeight := int((*img).Ysize)

	// Double check dimensions because gif may have many frames
	if err := checkDimensions(imgWidth, imgHeight); err != nil {
		return err
	}

	frameHeight, err := vipsGetInt(*img, "page-height")
	if err != nil {
		return err
	}

	delay, err := vipsGetInt(*img, "gif-delay")
	if err != nil {
		return err
	}

	loop, err := vipsGetInt(*img, "gif-loop")
	if err != nil {
		return err
	}

	framesCount := minInt(imgHeight/frameHeight, conf.MaxGifFrames)

	frames := make([]*C.VipsImage, framesCount)
	defer func() {
		for _, frame := range frames {
			C.clear_image(&frame)
		}
	}()

	var errg errgroup.Group

	for i := 0; i < framesCount; i++ {
		ind := i
		errg.Go(func() error {
			var frame *C.VipsImage

			if err := vipsExtract(*img, &frame, 0, ind*frameHeight, imgWidth, frameHeight); err != nil {
				return err
			}

			if err := transformImage(ctx, &frame, nil, po, imageTypeGIF); err != nil {
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

	if err := vipsArrayjoin(frames, img); err != nil {
		return err
	}

	vipsSetInt(*img, "page-height", int(frames[0].Ysize))
	vipsSetInt(*img, "gif-delay", delay)
	vipsSetInt(*img, "gif-loop", loop)

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

	defer C.vips_cleanup()

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

	img, err := vipsLoadImage(data, imgtype, 1, 1.0, po.Format == imageTypeGIF)
	if err != nil {
		return nil, func() {}, err
	}
	defer C.clear_image(&img)

	if imgtype == imageTypeGIF && po.Format == imageTypeGIF && vipsIsAnimatedGif(img) {
		if err := transformGif(ctx, &img, po); err != nil {
			return nil, func() {}, err
		}
	} else {
		if err := transformImage(ctx, &img, data, po, imgtype); err != nil {
			return nil, func() {}, err
		}
	}

	checkTimeout(ctx)

	if po.Format == imageTypeGIF {
		if err := vipsCastUchar(&img); err != nil {
			return nil, func() {}, err
		}
		checkTimeout(ctx)
	}

	return vipsSaveImage(img, po.Format, po.Quality)
}

func vipsPrepareWatermark() error {
	data, imgtype, cancel, err := watermarkData()
	defer cancel()

	if err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	watermark, err = vipsLoadImage(data, imgtype, 1, 1.0, false)
	if err != nil {
		return err
	}

	var tmp *C.VipsImage

	if C.vips_apply_opacity(watermark, &tmp, C.double(conf.WatermarkOpacity)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&watermark, tmp)

	if tmp = C.vips_image_copy_memory(watermark); tmp == nil {
		return vipsError()
	}
	C.swap_and_clear(&watermark, tmp)

	return nil
}

func vipsLoadImage(data []byte, imgtype imageType, shrink int, svgScale float64, allPages bool) (*C.VipsImage, error) {
	var img *C.VipsImage

	err := C.int(0)

	switch imgtype {
	case imageTypeJPEG:
		err = C.vips_jpegload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(shrink), &img)
	case imageTypePNG:
		err = C.vips_pngload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &img)
	case imageTypeWEBP:
		err = C.vips_webpload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(shrink), &img)
	case imageTypeGIF:
		pages := C.int(1)
		if allPages {
			pages = -1
		}

		err = C.vips_gifload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), pages, &img)
	case imageTypeSVG:
		err = C.vips_svgload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.double(svgScale), &img)
	case imageTypeICO:
		rawData, width, height, icoErr := icoData(data)
		if icoErr != nil {
			return nil, icoErr
		}

		img = C.vips_image_new_from_memory_copy(unsafe.Pointer(&rawData[0]), C.size_t(width*height*4), C.int(width), C.int(height), 4, C.VIPS_FORMAT_UCHAR)
	}
	if err != 0 {
		return nil, vipsError()
	}

	return img, nil
}

func vipsSaveImage(img *C.VipsImage, imgtype imageType, quality int) ([]byte, context.CancelFunc, error) {
	var ptr unsafe.Pointer

	cancel := func() {
		C.g_free_go(&ptr)
	}

	err := C.int(0)

	imgsize := C.size_t(0)

	switch imgtype {
	case imageTypeJPEG:
		err = C.vips_jpegsave_go(img, &ptr, &imgsize, C.int(quality), cConf.JpegProgressive)
	case imageTypePNG:
		err = C.vips_pngsave_go(img, &ptr, &imgsize, cConf.PngInterlaced, cConf.PngQuantize, cConf.PngQuantizationColors)
	case imageTypeWEBP:
		err = C.vips_webpsave_go(img, &ptr, &imgsize, C.int(quality))
	case imageTypeGIF:
		err = C.vips_gifsave_go(img, &ptr, &imgsize)
	case imageTypeICO:
		err = C.vips_icosave_go(img, &ptr, &imgsize)
	}
	if err != 0 {
		C.g_free_go(&ptr)
		return nil, cancel, vipsError()
	}

	const maxBufSize = ^uint32(0)

	b := (*[maxBufSize]byte)(ptr)[:int(imgsize):int(imgsize)]

	return b, cancel, nil
}

func vipsArrayjoin(in []*C.VipsImage, out **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_arrayjoin_go(&in[0], &tmp, C.int(len(in))) != 0 {
		return vipsError()
	}

	C.swap_and_clear(out, tmp)
	return nil
}

func vipsIsAnimatedGif(img *C.VipsImage) bool {
	return C.vips_is_animated_gif(img) > 0
}

func vipsImageHasAlpha(img *C.VipsImage) bool {
	return C.vips_image_hasalpha_go(img) > 0
}

func vipsGetInt(img *C.VipsImage, name string) (int, error) {
	var i C.int

	if C.vips_image_get_int(img, cachedCString(name), &i) != 0 {
		return 0, vipsError()
	}
	return int(i), nil
}

func vipsSetInt(img *C.VipsImage, name string, value int) {
	C.vips_image_set_int(img, cachedCString(name), C.int(value))
}

func vipsCastUchar(img **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_image_get_format(*img) != C.VIPS_FORMAT_UCHAR {
		if C.vips_cast_go(*img, &tmp, C.VIPS_FORMAT_UCHAR) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsRad2Float(img **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_image_get_coding(*img) == C.VIPS_CODING_RAD {
		if C.vips_rad2float_go(*img, &tmp) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsResize(img **C.VipsImage, scale float64, hasAlpa bool) error {
	var tmp *C.VipsImage

	if hasAlpa {
		if C.vips_resize_with_premultiply(*img, &tmp, C.double(scale)) != 0 {
			return vipsError()
		}
	} else {
		if C.vips_resize_go(*img, &tmp, C.double(scale)) != 0 {
			return vipsError()
		}
	}

	C.swap_and_clear(img, tmp)

	return nil
}

func vipsRotate(img **C.VipsImage, angle int) error {
	var tmp *C.VipsImage

	if C.vips_rot_go(*img, &tmp, C.VipsAngle(angle)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsFlip(img **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_flip_horizontal_go(*img, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsCrop(img **C.VipsImage, left, top, width, height int) error {
	var tmp *C.VipsImage

	if C.vips_extract_area_go(*img, &tmp, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsExtract(in *C.VipsImage, out **C.VipsImage, left, top, width, height int) error {
	if C.vips_extract_area_go(in, out, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}
	return nil
}

func vipsSmartCrop(img **C.VipsImage, width, height int) error {
	var tmp *C.VipsImage

	if C.vips_smartcrop_go(*img, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsEnsureAlpha(img **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_ensure_alpha(*img, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsFlatten(img **C.VipsImage, bg rgbColor) error {
	var tmp *C.VipsImage

	if C.vips_flatten_go(*img, &tmp, C.double(bg.R), C.double(bg.G), C.double(bg.B)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsBlur(img **C.VipsImage, sigma float32) error {
	var tmp *C.VipsImage

	if C.vips_gaussblur_go(*img, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsSharpen(img **C.VipsImage, sigma float32) error {
	var tmp *C.VipsImage

	if C.vips_sharpen_go(*img, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsImportColourProfile(img **C.VipsImage, evenSRGB bool) error {
	var tmp *C.VipsImage

	if C.vips_need_icc_import(*img) > 0 && (evenSRGB || C.vips_icc_is_srgb_iec61966(*img) == 0) {
		profile, err := cmykProfilePath()
		if err != nil {
			return err
		}

		if C.vips_icc_import_go(*img, &tmp, cachedCString(profile)) == 0 {
			C.swap_and_clear(img, tmp)
		} else {
			logWarning("Can't import ICC profile: %s", vipsError())
		}
	}

	return nil
}

func vipsLinearColourspace(img **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_image_guess_interpretation(*img) != C.VIPS_INTERPRETATION_scRGB {
		if C.vips_colourspace_go(*img, &tmp, C.VIPS_INTERPRETATION_scRGB) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsFixColourspace(img **C.VipsImage) error {
	var tmp *C.VipsImage

	if C.vips_image_guess_interpretation(*img) != C.VIPS_INTERPRETATION_sRGB {
		if C.vips_colourspace_go(*img, &tmp, C.VIPS_INTERPRETATION_sRGB) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsImageCopyMemory(img **C.VipsImage) error {
	var tmp *C.VipsImage
	if tmp = C.vips_image_copy_memory(*img); tmp == nil {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)
	return nil
}

func vipsReplicate(img **C.VipsImage, width, height C.int) error {
	var tmp *C.VipsImage

	if C.vips_replicate_go(*img, &tmp, width, height) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsEmbed(img **C.VipsImage, gravity gravityType, width, height C.int, offX, offY C.int) error {
	wmWidth := (*img).Xsize
	wmHeight := (*img).Ysize

	left := (width-wmWidth+1)/2 + offX
	top := (height-wmHeight+1)/2 + offY

	if gravity == gravityNorth || gravity == gravityNorthEast || gravity == gravityNorthWest {
		top = offY
	}

	if gravity == gravityEast || gravity == gravityNorthEast || gravity == gravitySouthEast {
		left = width - wmWidth - offX
	}

	if gravity == gravitySouth || gravity == gravitySouthEast || gravity == gravitySouthWest {
		top = height - wmHeight - offY
	}

	if gravity == gravityWest || gravity == gravityNorthWest || gravity == gravitySouthWest {
		left = offX
	}

	if left > width {
		left = width - wmWidth
	} else if left < -wmWidth {
		left = 0
	}

	if top > height {
		top = height - wmHeight
	} else if top < -wmHeight {
		top = 0
	}

	var tmp *C.VipsImage
	if C.vips_embed_go(*img, &tmp, left, top, width, height) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsResizeWatermark(width, height int) (wm *C.VipsImage, err error) {
	wmW := float64(watermark.Xsize)
	wmH := float64(watermark.Ysize)

	wr := float64(width) / wmW
	hr := float64(height) / wmH

	scale := math.Min(wr, hr)

	if wmW*scale < 1 {
		scale = 1 / wmW
	}

	if wmH*scale < 1 {
		scale = 1 / wmH
	}

	if C.vips_resize_with_premultiply(watermark, &wm, C.double(scale)) != 0 {
		err = vipsError()
	}

	return
}

func vipsApplyWatermark(img **C.VipsImage, opts *watermarkOptions) error {
	if watermark == nil {
		return nil
	}

	var wm, tmp *C.VipsImage
	defer C.clear_image(&wm)

	var err error

	imgW := (*img).Xsize
	imgH := (*img).Ysize

	if opts.Scale == 0 {
		if C.vips_copy_go(watermark, &wm) != 0 {
			return vipsError()
		}
	} else {
		wmW := maxInt(int(float64(imgW)*opts.Scale), 1)
		wmH := maxInt(int(float64(imgH)*opts.Scale), 1)

		if wm, err = vipsResizeWatermark(wmW, wmH); err != nil {
			return err
		}
	}

	if opts.Replicate {
		if err = vipsReplicate(&wm, imgW, imgH); err != nil {
			return err
		}
	} else {
		if err = vipsEmbed(&wm, opts.Gravity, imgW, imgH, C.int(opts.OffsetX), C.int(opts.OffsetY)); err != nil {
			return err
		}
	}

	if C.vips_apply_watermark(*img, wm, &tmp, C.double(opts.Opacity)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsError() error {
	return newUnexpectedError(C.GoString(C.vips_error_buffer()), 1)
}
