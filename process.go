package main

/*
#cgo pkg-config: vips
#cgo LDFLAGS: -s -w
#include "vips.h"
*/
import "C"

import (
	"context"
	"errors"
	"log"
	"math"
	"os"
	"runtime"
	"unsafe"
)

var (
	vipsSupportSmartcrop bool
	vipsTypeSupportLoad  = make(map[imageType]bool)
	vipsTypeSupportSave  = make(map[imageType]bool)

	watermark *C.struct__VipsImage

	errSmartCropNotSupported = errors.New("Smart crop is not supported by used version of libvips")
)

type cConfig struct {
	Quality          C.int
	JpegProgressive  C.int
	PngInterlaced    C.int
	WatermarkOpacity C.double
}

var cConf cConfig

func initVips() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		log.Fatalln("unable to start vips!")
	}

	// Disable libvips cache. Since processing pipeline is fine tuned, we won't get much profit from it.
	// Enabled cache can cause SIGSEGV on Musl-based systems like Alpine.
	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)

	if len(os.Getenv("IMGPROXY_VIPS_LEAK_CHECK")) > 0 {
		C.vips_leak_set(C.gboolean(1))
	}

	if len(os.Getenv("IMGPROXY_VIPS_CACHE_TRACE")) > 0 {
		C.vips_cache_set_trace(C.gboolean(1))
	}

	vipsSupportSmartcrop = C.vips_support_smartcrop() == 1

	if int(C.vips_type_find_load_go(imageTypeJPEG)) != 0 {
		vipsTypeSupportLoad[imageTypeJPEG] = true
	}
	if int(C.vips_type_find_load_go(imageTypePNG)) != 0 {
		vipsTypeSupportLoad[imageTypePNG] = true
	}
	if int(C.vips_type_find_load_go(imageTypeWEBP)) != 0 {
		vipsTypeSupportLoad[imageTypeWEBP] = true
	}
	if int(C.vips_type_find_load_go(imageTypeGIF)) != 0 {
		vipsTypeSupportLoad[imageTypeGIF] = true
	}

	if int(C.vips_type_find_save_go(imageTypeJPEG)) != 0 {
		vipsTypeSupportSave[imageTypeJPEG] = true
	}
	if int(C.vips_type_find_save_go(imageTypePNG)) != 0 {
		vipsTypeSupportSave[imageTypePNG] = true
	}
	if int(C.vips_type_find_save_go(imageTypeWEBP)) != 0 {
		vipsTypeSupportSave[imageTypeWEBP] = true
	}

	cConf.Quality = C.int(conf.Quality)

	if conf.JpegProgressive {
		cConf.JpegProgressive = C.int(1)
	}

	if conf.PngInterlaced {
		cConf.PngInterlaced = C.int(1)
	}

	cConf.WatermarkOpacity = C.double(conf.WatermarkOpacity)

	if err := vipsPrepareWatermark(); err != nil {
		log.Fatal(err)
	}
}

func shutdownVips() {
	C.clear_image(&watermark)
	C.vips_shutdown()
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

func needToScale(width, height int, po *processingOptions) bool {
	return ((po.Width != 0 && po.Width != width) || (po.Height != 0 && po.Height != height)) &&
		(po.Resize == resizeFill || po.Resize == resizeFit)
}

func calcScale(width, height int, po *processingOptions) float64 {
	srcW, srcH := float64(width), float64(height)

	wr := float64(po.Width) / srcW
	hr := float64(po.Height) / srcH

	var scale float64

	if po.Width == 0 {
		scale = hr
	} else if po.Height == 0 {
		scale = wr
	} else if po.Resize == resizeFit {
		scale = math.Min(wr, hr)
	} else {
		scale = math.Max(wr, hr)
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
		case shrink >= 16:
			return 8
		case shrink >= 8:
			return 4
		case shrink >= 4:
			return 2
		}
	}

	return 1
}

func calcCrop(width, height int, po *processingOptions) (left, top int) {
	if po.Gravity.Type == gravityFocusPoint {
		pointX := int(float64(width) * po.Gravity.X)
		pointY := int(float64(height) * po.Gravity.Y)

		left = maxInt(0, minInt(pointX-po.Width/2, width-po.Width))
		top = maxInt(0, minInt(pointY-po.Height/2, height-po.Height))

		return
	}

	left = (width - po.Width + 1) / 2
	top = (height - po.Height + 1) / 2

	if po.Gravity.Type == gravityNorth || po.Gravity.Type == gravityNorthEast || po.Gravity.Type == gravityNorthWest {
		top = 0
	}

	if po.Gravity.Type == gravityEast || po.Gravity.Type == gravityNorthEast || po.Gravity.Type == gravitySouthEast {
		left = width - po.Width
	}

	if po.Gravity.Type == gravitySouth || po.Gravity.Type == gravitySouthEast || po.Gravity.Type == gravitySouthWest {
		top = height - po.Height
	}

	if po.Gravity.Type == gravityWest || po.Gravity.Type == gravityNorthWest || po.Gravity.Type == gravitySouthWest {
		left = 0
	}

	return
}

func processImage(ctx context.Context) ([]byte, error) {
	if newRelicEnabled {
		newRelicCancel := startNewRelicSegment(ctx, "Processing image")
		defer newRelicCancel()
	}

	defer C.vips_cleanup()

	data := getImageData(ctx).Bytes()
	po := getProcessingOptions(ctx)
	imgtype := getImageType(ctx)

	if po.Gravity.Type == gravitySmart && !vipsSupportSmartcrop {
		return nil, errSmartCropNotSupported
	}

	img, err := vipsLoadImage(data, imgtype, 1)
	if err != nil {
		return nil, err
	}
	defer C.clear_image(&img)

	checkTimeout(ctx)

	imgWidth, imgHeight, angle, flip := extractMeta(img)

	// Ensure we won't crop out of bounds
	if !po.Enlarge || po.Resize == resizeCrop {
		if imgWidth < po.Width {
			po.Width = imgWidth
		}

		if imgHeight < po.Height {
			po.Height = imgHeight
		}
	}

	hasAlpha := vipsImageHasAlpha(img)

	if needToScale(imgWidth, imgHeight, po) {
		scale := calcScale(imgWidth, imgHeight, po)

		// Do some shrink-on-load
		if scale < 1.0 {
			if shrink := calcShink(scale, imgtype); shrink != 1 {
				scale = scale * float64(shrink)

				if tmp, e := vipsLoadImage(data, imgtype, shrink); e == nil {
					C.swap_and_clear(&img, tmp)
				} else {
					return nil, e
				}
			}
		}

		premultiplied := false
		var bandFormat C.VipsBandFormat

		if hasAlpha {
			if bandFormat, err = vipsPremultiply(&img); err != nil {
				return nil, err
			}
			premultiplied = true
		}

		if err = vipsResize(&img, scale); err != nil {
			return nil, err
		}

		// Update actual image size after resize
		imgWidth, imgHeight, _, _ = extractMeta(img)

		if premultiplied {
			if err = vipsUnpremultiply(&img, bandFormat); err != nil {
				return nil, err
			}
		}
	}

	if err = vipsImportColourProfile(&img); err != nil {
		return nil, err
	}

	if err = vipsFixColourspace(&img); err != nil {
		return nil, err
	}

	checkTimeout(ctx)

	if angle != C.VIPS_ANGLE_D0 || flip {
		if err = vipsImageCopyMemory(&img); err != nil {
			return nil, err
		}

		if angle != C.VIPS_ANGLE_D0 {
			if err = vipsRotate(&img, angle); err != nil {
				return nil, err
			}
		}

		if flip {
			if err = vipsFlip(&img); err != nil {
				return nil, err
			}
		}
	}

	checkTimeout(ctx)

	if po.Width == 0 {
		po.Width = imgWidth
	}

	if po.Height == 0 {
		po.Height = imgHeight
	}

	if po.Width < imgWidth || po.Height < imgHeight {
		if po.Gravity.Type == gravitySmart {
			if err = vipsImageCopyMemory(&img); err != nil {
				return nil, err
			}
			if err = vipsSmartCrop(&img, po.Width, po.Height); err != nil {
				return nil, err
			}
			// Applying additional modifications after smart crop causes SIGSEGV on Alpine
			// so we have to copy memory after it
			if err = vipsImageCopyMemory(&img); err != nil {
				return nil, err
			}
		} else {
			left, top := calcCrop(imgWidth, imgHeight, po)
			if err = vipsCrop(&img, left, top, po.Width, po.Height); err != nil {
				return nil, err
			}
		}

		checkTimeout(ctx)
	}

	if hasAlpha && po.Flatten {
		if err = vipsFlatten(&img, po.Background); err != nil {
			return nil, err
		}
	}

	if po.Blur > 0 {
		if err = vipsBlur(&img, po.Blur); err != nil {
			return nil, err
		}
	}

	if po.Sharpen > 0 {
		if err = vipsSharpen(&img, po.Sharpen); err != nil {
			return nil, err
		}
	}

	checkTimeout(ctx)

	if po.Watermark.Enabled {
		if err = vipsApplyWatermark(&img, &po.Watermark); err != nil {
			return nil, err
		}
	}

	return vipsSaveImage(img, po.Format)
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

	if C.vips_load_buffer(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(imgtype), 1, &watermark) != 0 {
		return vipsError()
	}

	var tmp *C.struct__VipsImage

	if cConf.WatermarkOpacity < 1 {
		if vipsImageHasAlpha(watermark) {
			var alpha *C.struct__VipsImage
			defer C.clear_image(&alpha)

			if C.vips_extract_band_go(watermark, &tmp, (*watermark).Bands-1, 1) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&alpha, tmp)

			if C.vips_extract_band_go(watermark, &tmp, 0, (*watermark).Bands-1) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&watermark, tmp)

			if C.vips_linear_go(alpha, &tmp, cConf.WatermarkOpacity, 0) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&alpha, tmp)

			if C.vips_bandjoin_go(watermark, alpha, &tmp) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&watermark, tmp)
		} else {
			if C.vips_bandjoin_const_go(watermark, &tmp, cConf.WatermarkOpacity*255) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&watermark, tmp)
		}
	}

	if tmp = C.vips_image_copy_memory(watermark); tmp == nil {
		return vipsError()
	}
	C.swap_and_clear(&watermark, tmp)

	return nil
}

func vipsLoadImage(data []byte, imgtype imageType, shrink int) (*C.struct__VipsImage, error) {
	var img *C.struct__VipsImage
	if C.vips_load_buffer(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(imgtype), C.int(shrink), &img) != 0 {
		return nil, vipsError()
	}
	return img, nil
}

func vipsSaveImage(img *C.struct__VipsImage, imgtype imageType) ([]byte, error) {
	var ptr unsafe.Pointer
	defer C.g_free_go(&ptr)

	err := C.int(0)

	imgsize := C.size_t(0)

	switch imgtype {
	case imageTypeJPEG:
		err = C.vips_jpegsave_go(img, &ptr, &imgsize, 1, cConf.Quality, cConf.JpegProgressive)
	case imageTypePNG:
		err = C.vips_pngsave_go(img, &ptr, &imgsize, cConf.PngInterlaced)
	case imageTypeWEBP:
		err = C.vips_webpsave_go(img, &ptr, &imgsize, 1, cConf.Quality)
	}
	if err != 0 {
		return nil, vipsError()
	}

	return C.GoBytes(ptr, C.int(imgsize)), nil
}

func vipsImageHasAlpha(img *C.struct__VipsImage) bool {
	return C.vips_image_hasalpha_go(img) > 0
}

func vipsPremultiply(img **C.struct__VipsImage) (C.VipsBandFormat, error) {
	var tmp *C.struct__VipsImage

	format := C.vips_band_format(*img)

	if C.vips_premultiply_go(*img, &tmp) != 0 {
		return 0, vipsError()
	}

	C.swap_and_clear(img, tmp)
	return format, nil
}

func vipsUnpremultiply(img **C.struct__VipsImage, format C.VipsBandFormat) error {
	var tmp *C.struct__VipsImage

	if C.vips_unpremultiply_go(*img, &tmp) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	if C.vips_cast_go(*img, &tmp, format) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsResize(img **C.struct__VipsImage, scale float64) error {
	var tmp *C.struct__VipsImage

	if C.vips_resize_go(*img, &tmp, C.double(scale)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsRotate(img **C.struct__VipsImage, angle int) error {
	var tmp *C.struct__VipsImage

	if C.vips_rot_go(*img, &tmp, C.VipsAngle(angle)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsFlip(img **C.struct__VipsImage) error {
	var tmp *C.struct__VipsImage

	if C.vips_flip_horizontal_go(*img, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsCrop(img **C.struct__VipsImage, left, top, width, height int) error {
	var tmp *C.struct__VipsImage

	if C.vips_extract_area_go(*img, &tmp, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsSmartCrop(img **C.struct__VipsImage, width, height int) error {
	var tmp *C.struct__VipsImage

	if C.vips_smartcrop_go(*img, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsFlatten(img **C.struct__VipsImage, bg color) error {
	var tmp *C.struct__VipsImage

	if C.vips_flatten_go(*img, &tmp, C.double(bg.R), C.double(bg.G), C.double(bg.B)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsBlur(img **C.struct__VipsImage, sigma float32) error {
	var tmp *C.struct__VipsImage

	if C.vips_gaussblur_go(*img, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsSharpen(img **C.struct__VipsImage, sigma float32) error {
	var tmp *C.struct__VipsImage

	if C.vips_sharpen_go(*img, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(img, tmp)
	return nil
}

func vipsImportColourProfile(img **C.struct__VipsImage) error {
	var tmp *C.struct__VipsImage

	if C.vips_need_icc_import(*img) > 0 {
		profile, err := cmykProfilePath()
		if err != nil {
			return err
		}

		if C.vips_icc_import_go(*img, &tmp, C.CString(profile)) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsFixColourspace(img **C.struct__VipsImage) error {
	var tmp *C.struct__VipsImage

	if C.vips_image_guess_interpretation(*img) != C.VIPS_INTERPRETATION_sRGB {
		if C.vips_colourspace_go(*img, &tmp, C.VIPS_INTERPRETATION_sRGB) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsImageCopyMemory(img **C.struct__VipsImage) error {
	var tmp *C.struct__VipsImage
	if tmp = C.vips_image_copy_memory(*img); tmp == nil {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)
	return nil
}

func vipsReplicate(img **C.struct__VipsImage, width, height C.int) error {
	var tmp *C.struct__VipsImage

	if C.vips_replicate_go(*img, &tmp, 1+width/(*img).Xsize, 1+height/(*img).Ysize) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	if C.vips_extract_area_go(*img, &tmp, 0, 0, width, height) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsEmbed(img **C.struct__VipsImage, gravity gravityType, width, height C.int, offX, offY C.int) error {
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

	var tmp *C.struct__VipsImage
	if C.vips_embed_go(*img, &tmp, left, top, width, height) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	return nil
}

func vipsResizeWatermark(width, height int) (wm *C.struct__VipsImage, err error) {
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

	if C.vips_resize_go(watermark, &wm, C.double(scale)) != 0 {
		err = vipsError()
	}

	return
}

func vipsApplyWatermark(img **C.struct__VipsImage, opts *watermarkOptions) error {
	if watermark == nil {
		return nil
	}

	var wm, wmAlpha, tmp *C.struct__VipsImage
	var err error

	defer C.clear_image(&wm)
	defer C.clear_image(&wmAlpha)

	imgW := (*img).Xsize
	imgH := (*img).Ysize

	if opts.Scale == 0 {
		if wm = C.vips_image_copy_memory(watermark); wm == nil {
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

	if C.vips_extract_band_go(wm, &tmp, (*wm).Bands-1, 1) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&wmAlpha, tmp)

	if C.vips_extract_band_go(wm, &tmp, 0, (*wm).Bands-1) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&wm, tmp)

	if C.vips_image_guess_interpretation(*img) != C.vips_image_guess_interpretation(wm) {
		if C.vips_colourspace_go(wm, &tmp, C.vips_image_guess_interpretation(*img)) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&wm, tmp)
	}

	if opts.Opacity < 1 {
		if C.vips_linear_go(wmAlpha, &tmp, C.double(opts.Opacity), 0) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&wmAlpha, tmp)
	}

	var imgAlpha *C.struct__VipsImage
	defer C.clear_image(&imgAlpha)

	hasAlpha := vipsImageHasAlpha(*img)

	if hasAlpha {
		if C.vips_extract_band_go(*img, &imgAlpha, (**img).Bands-1, 1) != 0 {
			return vipsError()
		}

		if C.vips_extract_band_go(*img, &tmp, 0, (**img).Bands-1) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	if C.vips_ifthenelse_go(wmAlpha, wm, *img, &tmp) != 0 {
		return vipsError()
	}
	C.swap_and_clear(img, tmp)

	if hasAlpha {
		if C.vips_bandjoin_go(*img, imgAlpha, &tmp) != 0 {
			return vipsError()
		}
		C.swap_and_clear(img, tmp)
	}

	return nil
}

func vipsError() error {
	return errors.New(C.GoString(C.vips_error_buffer()))
}
