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

	errSmartCropNotSupported = errors.New("Smart crop is not supported by used version of libvips")
)

type cConfig struct {
	Quality         C.int
	JpegProgressive C.int
	PngInterlaced   C.int
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
}

func shutdownVips() {
	C.vips_shutdown()
}

func round(f float64) int {
	return int(f + .5)
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

func needToCrop(width, height int, po *processingOptions) bool {
	return (po.Width != width || po.Height != height) &&
		(po.Resize == resizeFill || po.Resize == resizeCrop)
}

func calcScale(width, height int, po *processingOptions) float64 {
	fsw, fsh, fow, foh := float64(width), float64(height), float64(po.Width), float64(po.Height)

	wr := fow / fsw
	hr := foh / fsh

	if po.Width == 0 {
		return hr
	}

	if po.Height == 0 {
		return wr
	}

	if po.Resize == resizeFit {
		return math.Min(wr, hr)
	}

	return math.Max(wr, hr)
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
	left = (width - po.Width + 1) / 2
	top = (height - po.Height + 1) / 2

	if po.Gravity.Type == gravityNorth {
		top = 0
	}

	if po.Gravity.Type == gravityEast {
		left = width - po.Width
	}

	if po.Gravity.Type == gravitySouth {
		top = height - po.Height
	}

	if po.Gravity.Type == gravityWest {
		left = 0
	}

	if po.Gravity.Type == gravityFocusPoint {
		pointX := int(float64(width) * po.Gravity.X)
		pointY := int(float64(height) * po.Gravity.Y)

		left = maxInt(0, minInt(pointX-po.Width/2, width-po.Width))
		top = maxInt(0, minInt(pointY-po.Height/2, height-po.Height))
	}

	return
}

func processImage(ctx context.Context) ([]byte, error) {
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

	if needToCrop(imgWidth, imgHeight, po) {
		if po.Gravity.Type == gravitySmart {
			if err = vipsImageCopyMemory(&img); err != nil {
				return nil, err
			}
			if err = vipsSmartCrop(&img, po.Width, po.Height); err != nil {
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

	return vipsSaveImage(img, po.Format)
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

func vipsError() error {
	return errors.New(C.GoString(C.vips_error_buffer()))
}
