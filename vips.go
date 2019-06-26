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
	"math"
	"os"
	"runtime"
	"time"
	"unsafe"
)

type vipsImage struct {
	VipsImage *C.VipsImage
}

var (
	vipsSupportSmartcrop bool
	vipsTypeSupportLoad  = make(map[imageType]bool)
	vipsTypeSupportSave  = make(map[imageType]bool)

	watermark *vipsImage
)

var vipsConf struct {
	JpegProgressive       C.int
	PngInterlaced         C.int
	PngQuantize           C.int
	PngQuantizationColors C.int
	WatermarkOpacity      C.double
}

const (
	vipsAngleD0   = C.VIPS_ANGLE_D0
	vipsAngleD90  = C.VIPS_ANGLE_D90
	vipsAngleD180 = C.VIPS_ANGLE_D180
	vipsAngleD270 = C.VIPS_ANGLE_D270
)

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
	if int(C.vips_type_find_load_go(C.int(imageTypeHEIC))) != 0 {
		vipsTypeSupportLoad[imageTypeHEIC] = true
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
	if int(C.vips_type_find_save_go(C.int(imageTypeHEIC))) != 0 {
		vipsTypeSupportSave[imageTypeHEIC] = true
	}

	if conf.JpegProgressive {
		vipsConf.JpegProgressive = C.int(1)
	}

	if conf.PngInterlaced {
		vipsConf.PngInterlaced = C.int(1)
	}

	if conf.PngQuantize {
		vipsConf.PngQuantize = C.int(1)
	}

	vipsConf.PngQuantizationColors = C.int(conf.PngQuantizationColors)

	vipsConf.WatermarkOpacity = C.double(conf.WatermarkOpacity)

	if err := vipsPrepareWatermark(); err != nil {
		logFatal(err.Error())
	}

	vipsCollectMetrics()
}

func shutdownVips() {
	if watermark != nil {
		watermark.Clear()
	}

	C.vips_shutdown()
}

func vipsCollectMetrics() {
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

func vipsCleanup() {
	C.vips_cleanup()
}

func vipsError() error {
	return newUnexpectedError(C.GoString(C.vips_error_buffer()), 1)
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

	watermark = new(vipsImage)

	if err = watermark.Load(data, imgtype, 1, 1.0, 1); err != nil {
		return err
	}

	var tmp *C.VipsImage

	if C.vips_apply_opacity(watermark.VipsImage, &tmp, C.double(conf.WatermarkOpacity)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&watermark.VipsImage, tmp)

	if err = watermark.CopyMemory(); err != nil {
		return err
	}

	return nil
}

func vipsResizeWatermark(width, height int) (wm *vipsImage, err error) {
	wmW := float64(watermark.VipsImage.Xsize)
	wmH := float64(watermark.VipsImage.Ysize)

	wr := float64(width) / wmW
	hr := float64(height) / wmH

	scale := math.Min(wr, hr)

	if wmW*scale < 1 {
		scale = 1 / wmW
	}

	if wmH*scale < 1 {
		scale = 1 / wmH
	}

	wm = new(vipsImage)

	if C.vips_resize_with_premultiply(watermark.VipsImage, &wm.VipsImage, C.double(scale)) != 0 {
		err = vipsError()
	}

	return
}

func (img *vipsImage) Width() int {
	return int(img.VipsImage.Xsize)
}

func (img *vipsImage) Height() int {
	return int(img.VipsImage.Ysize)
}

func (img *vipsImage) Load(data []byte, imgtype imageType, shrink int, scale float64, pages int) error {
	var tmp *C.VipsImage

	err := C.int(0)

	switch imgtype {
	case imageTypeJPEG:
		err = C.vips_jpegload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(shrink), &tmp)
	case imageTypePNG:
		err = C.vips_pngload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	case imageTypeWEBP:
		err = C.vips_webpload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.double(scale), C.int(pages), &tmp)
	case imageTypeGIF:
		err = C.vips_gifload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(pages), &tmp)
	case imageTypeSVG:
		err = C.vips_svgload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.double(scale), &tmp)
	case imageTypeICO:
		rawData, width, height, icoErr := icoData(data)
		if icoErr != nil {
			return icoErr
		}

		tmp = C.vips_image_new_from_memory_copy(unsafe.Pointer(&rawData[0]), C.size_t(width*height*4), C.int(width), C.int(height), 4, C.VIPS_FORMAT_UCHAR)
	case imageTypeHEIC:
		err = C.vips_heifload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	}
	if err != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) Save(imgtype imageType, quality int) ([]byte, context.CancelFunc, error) {
	var ptr unsafe.Pointer

	cancel := func() {
		C.g_free_go(&ptr)
	}

	err := C.int(0)

	imgsize := C.size_t(0)

	switch imgtype {
	case imageTypeJPEG:
		err = C.vips_jpegsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), vipsConf.JpegProgressive)
	case imageTypePNG:
		err = C.vips_pngsave_go(img.VipsImage, &ptr, &imgsize, vipsConf.PngInterlaced, vipsConf.PngQuantize, vipsConf.PngQuantizationColors)
	case imageTypeWEBP:
		err = C.vips_webpsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality))
	case imageTypeGIF:
		err = C.vips_gifsave_go(img.VipsImage, &ptr, &imgsize)
	case imageTypeICO:
		err = C.vips_icosave_go(img.VipsImage, &ptr, &imgsize)
	case imageTypeHEIC:
		err = C.vips_heifsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality))
	}
	if err != 0 {
		C.g_free_go(&ptr)
		return nil, cancel, vipsError()
	}

	const maxBufSize = ^uint32(0)

	b := (*[maxBufSize]byte)(ptr)[:int(imgsize):int(imgsize)]

	return b, cancel, nil
}

func (img *vipsImage) Clear() {
	if img.VipsImage != nil {
		C.clear_image(&img.VipsImage)
	}
}

func (img *vipsImage) Arrayjoin(in []*vipsImage) error {
	var tmp *C.VipsImage

	arr := make([]*C.VipsImage, len(in))
	for i, im := range in {
		arr[i] = im.VipsImage
	}

	if C.vips_arrayjoin_go(&arr[0], &tmp, C.int(len(arr))) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func vipsSupportAnimation(imgtype imageType) bool {
	return imgtype == imageTypeGIF ||
		(imgtype == imageTypeWEBP && C.vips_support_webp_animation() != 0)
}

func (img *vipsImage) IsAnimated() bool {
	return C.vips_is_animated(img.VipsImage) > 0
}

func (img *vipsImage) HasAlpha() bool {
	return C.vips_image_hasalpha_go(img.VipsImage) > 0
}

func (img *vipsImage) GetInt(name string) (int, error) {
	var i C.int

	if C.vips_image_get_int(img.VipsImage, cachedCString(name), &i) != 0 {
		return 0, vipsError()
	}
	return int(i), nil
}

func (img *vipsImage) SetInt(name string, value int) {
	C.vips_image_set_int(img.VipsImage, cachedCString(name), C.int(value))
}

func (img *vipsImage) CastUchar() error {
	var tmp *C.VipsImage

	if C.vips_image_get_format(img.VipsImage) != C.VIPS_FORMAT_UCHAR {
		if C.vips_cast_go(img.VipsImage, &tmp, C.VIPS_FORMAT_UCHAR) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *vipsImage) Rad2Float() error {
	var tmp *C.VipsImage

	if C.vips_image_get_coding(img.VipsImage) == C.VIPS_CODING_RAD {
		if C.vips_rad2float_go(img.VipsImage, &tmp) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *vipsImage) Resize(scale float64, hasAlpa bool) error {
	var tmp *C.VipsImage

	if hasAlpa {
		if C.vips_resize_with_premultiply(img.VipsImage, &tmp, C.double(scale)) != 0 {
			return vipsError()
		}
	} else {
		if C.vips_resize_go(img.VipsImage, &tmp, C.double(scale)) != 0 {
			return vipsError()
		}
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) Orientation() C.int {
	return C.vips_get_exif_orientation(img.VipsImage)
}

func (img *vipsImage) Rotate(angle int) error {
	var tmp *C.VipsImage

	if C.vips_rot_go(img.VipsImage, &tmp, C.VipsAngle(angle)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) Flip() error {
	var tmp *C.VipsImage

	if C.vips_flip_horizontal_go(img.VipsImage, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) Crop(left, top, width, height int) error {
	var tmp *C.VipsImage

	if C.vips_extract_area_go(img.VipsImage, &tmp, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) Extract(out *vipsImage, left, top, width, height int) error {
	if C.vips_extract_area_go(img.VipsImage, &out.VipsImage, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}
	return nil
}

func (img *vipsImage) SmartCrop(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_smartcrop_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) EnsureAlpha() error {
	var tmp *C.VipsImage

	if C.vips_ensure_alpha(img.VipsImage, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) Flatten(bg rgbColor) error {
	var tmp *C.VipsImage

	if C.vips_flatten_go(img.VipsImage, &tmp, C.double(bg.R), C.double(bg.G), C.double(bg.B)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) Blur(sigma float32) error {
	var tmp *C.VipsImage

	if C.vips_gaussblur_go(img.VipsImage, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) Sharpen(sigma float32) error {
	var tmp *C.VipsImage

	if C.vips_sharpen_go(img.VipsImage, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) ImportColourProfile(evenSRGB bool) error {
	var tmp *C.VipsImage

	if img.VipsImage.Coding != C.VIPS_CODING_NONE {
		return nil
	}

	if img.VipsImage.BandFmt != C.VIPS_FORMAT_UCHAR && img.VipsImage.BandFmt != C.VIPS_FORMAT_USHORT {
		return nil
	}

	profile := (*C.char)(nil)

	if C.vips_has_embedded_icc(img.VipsImage) == 0 {
		// No embedded profile
		// If vips doesn't have built-in profile, use profile built-in to imgproxy for CMYK
		// TODO: Remove this. Supporting built-in profiles is pain, vips does it better
		if img.VipsImage.Type == C.VIPS_INTERPRETATION_CMYK && C.vips_support_builtin_icc() == 0 {
			p, err := cmykProfilePath()
			if err != nil {
				return err
			}
			profile = cachedCString(p)
		} else {
			// imgproxy doesn't have built-in profile for other interpretations,
			// so we can't do anything here
			return nil
		}
	}

	// Don't import sRGB IEC61966 2.1 unless evenSRGB
	if img.VipsImage.Type == C.VIPS_INTERPRETATION_sRGB && !evenSRGB && C.vips_icc_is_srgb_iec61966(img.VipsImage) != 0 {
		return nil
	}

	if C.vips_icc_import_go(img.VipsImage, &tmp, profile) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		logWarning("Can't import ICC profile: %s", vipsError())
	}

	return nil
}

func (img *vipsImage) IsSRGB() bool {
	return img.VipsImage.Type == C.VIPS_INTERPRETATION_sRGB
}

func (img *vipsImage) LinearColourspace() error {
	return img.Colorspace(C.VIPS_INTERPRETATION_scRGB)
}

func (img *vipsImage) RgbColourspace() error {
	return img.Colorspace(C.VIPS_INTERPRETATION_sRGB)
}

func (img *vipsImage) Colorspace(colorspace C.VipsInterpretation) error {
	if img.VipsImage.Type != colorspace {
		var tmp *C.VipsImage

		if C.vips_colourspace_go(img.VipsImage, &tmp, colorspace) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *vipsImage) CopyMemory() error {
	var tmp *C.VipsImage
	if tmp = C.vips_image_copy_memory(img.VipsImage); tmp == nil {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *vipsImage) Replicate(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_replicate_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) Embed(gravity gravityType, width, height int, offX, offY int, bg rgbColor) error {
	wmWidth := img.Width()
	wmHeight := img.Height()

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

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	var bgc []C.double
	if img.HasAlpha() {
		bgc = []C.double{C.double(0)}
	} else {
		bgc = []C.double{C.double(bg.R), C.double(bg.G), C.double(bg.B)}
	}

	var tmp *C.VipsImage
	if C.vips_embed_go(img.VipsImage, &tmp, C.int(left), C.int(top), C.int(width), C.int(height), &bgc[0], C.int(len(bgc))) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) ApplyWatermark(opts *watermarkOptions) error {
	if watermark == nil {
		return nil
	}

	var (
		wm  *vipsImage
		tmp *C.VipsImage
	)
	defer func() { wm.Clear() }()

	var err error

	imgW := img.Width()
	imgH := img.Height()

	if opts.Scale == 0 {
		wm = new(vipsImage)

		if C.vips_copy_go(watermark.VipsImage, &wm.VipsImage) != 0 {
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
		if err = wm.Replicate(imgW, imgH); err != nil {
			return err
		}
	} else {
		if err = wm.Embed(opts.Gravity, imgW, imgH, opts.OffsetX, opts.OffsetY, rgbColor{0, 0, 0}); err != nil {
			return err
		}
	}

	if C.vips_apply_watermark(img.VipsImage, wm.VipsImage, &tmp, C.double(opts.Opacity)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}
