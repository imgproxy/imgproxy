package main

/*
#cgo pkg-config: vips
#cgo LDFLAGS: -s -w
#cgo CFLAGS: -O3
#include "vips.h"
*/
import "C"
import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"unsafe"
)

type vipsImage struct {
	VipsImage *C.VipsImage
}

var (
	vipsSupportSmartcrop bool
	vipsTypeSupportLoad  = make(map[imageType]bool)
	vipsTypeSupportSave  = make(map[imageType]bool)

	watermark *imageData
)

var vipsConf struct {
	JpegProgressive       C.int
	PngInterlaced         C.int
	PngQuantize           C.int
	PngQuantizationColors C.int
	AvifSpeed             C.int
	WatermarkOpacity      C.double
}

func initVips() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		return fmt.Errorf("unable to start vips!")
	}

	// Disable libvips cache. Since processing pipeline is fine tuned, we won't get much profit from it.
	// Enabled cache can cause SIGSEGV on Musl-based systems like Alpine.
	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)

	C.vips_concurrency_set(1)

	// Vector calculations cause SIGSEGV sometimes when working with JPEG.
	// It's better to disable it since profit it quite small
	C.vips_vector_set_enabled(0)

	if len(os.Getenv("IMGPROXY_VIPS_LEAK_CHECK")) > 0 {
		C.vips_leak_set(C.gboolean(1))
	}

	if len(os.Getenv("IMGPROXY_VIPS_CACHE_TRACE")) > 0 {
		C.vips_cache_set_trace(C.gboolean(1))
	}

	vipsSupportSmartcrop = C.vips_support_smartcrop() == 1

	for _, imgtype := range imageTypes {
		vipsTypeSupportLoad[imgtype] = int(C.vips_type_find_load_go(C.int(imgtype))) != 0
		vipsTypeSupportSave[imgtype] = int(C.vips_type_find_save_go(C.int(imgtype))) != 0
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
	vipsConf.AvifSpeed = C.int(conf.AvifSpeed)
	vipsConf.WatermarkOpacity = C.double(conf.WatermarkOpacity)

	if err := vipsLoadWatermark(); err != nil {
		C.vips_shutdown()
		return fmt.Errorf("Can't load watermark: %s", err)
	}

	return nil
}

func shutdownVips() {
	C.vips_shutdown()
}

func vipsGetMem() float64 {
	return float64(C.vips_tracked_get_mem())
}

func vipsGetMemHighwater() float64 {
	return float64(C.vips_tracked_get_mem_highwater())
}

func vipsGetAllocs() float64 {
	return float64(C.vips_tracked_get_allocs())
}

func vipsCleanup() {
	C.vips_cleanup()
}

func vipsError() error {
	return newUnexpectedError(C.GoString(C.vips_error_buffer()), 1)
}

func vipsLoadWatermark() (err error) {
	watermark, err = getWatermarkData()
	return
}

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
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
	case imageTypeHEIC, imageTypeAVIF:
		err = C.vips_heifload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	case imageTypeBMP:
		err = C.vips_bmpload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	case imageTypeTIFF:
		err = C.vips_tiffload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	}
	if err != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) Save(imgtype imageType, quality int) ([]byte, context.CancelFunc, error) {
	if imgtype == imageTypeICO {
		b, err := img.SaveAsIco()
		return b, func() {}, err
	}

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
	case imageTypeAVIF:
		err = C.vips_avifsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), vipsConf.AvifSpeed)
	case imageTypeBMP:
		err = C.vips_bmpsave_go(img.VipsImage, &ptr, &imgsize)
	case imageTypeTIFF:
		err = C.vips_tiffsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality))
	}
	if err != 0 {
		C.g_free_go(&ptr)
		return nil, cancel, vipsError()
	}

	b := ptrToBytes(ptr, int(imgsize))

	return b, cancel, nil
}

func (img *vipsImage) SaveAsIco() ([]byte, error) {
	if img.Width() > 256 || img.Height() > 256 {
		return nil, errors.New("Image dimensions is too big. Max dimension size for ICO is 256")
	}

	var ptr unsafe.Pointer
	imgsize := C.size_t(0)

	defer func() {
		C.g_free_go(&ptr)
	}()

	if C.vips_pngsave_go(img.VipsImage, &ptr, &imgsize, 0, 0, 256) != 0 {
		return nil, vipsError()
	}

	b := ptrToBytes(ptr, int(imgsize))

	buf := new(bytes.Buffer)
	buf.Grow(22 + int(imgsize))

	// ICONDIR header
	if _, err := buf.Write([]byte{0, 0, 1, 0, 1, 0}); err != nil {
		return nil, err
	}

	// ICONDIRENTRY
	if _, err := buf.Write([]byte{
		byte(img.Width() % 256),
		byte(img.Height() % 256),
	}); err != nil {
		return nil, err
	}
	// Number of colors. Not supported in our case
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// Reserved
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// Color planes. Always 1 in our case
	if _, err := buf.Write([]byte{1, 0}); err != nil {
		return nil, err
	}
	// Bits per pixel
	if img.HasAlpha() {
		if _, err := buf.Write([]byte{32, 0}); err != nil {
			return nil, err
		}
	} else {
		if _, err := buf.Write([]byte{24, 0}); err != nil {
			return nil, err
		}
	}
	// Image data size
	if err := binary.Write(buf, binary.LittleEndian, uint32(imgsize)); err != nil {
		return nil, err
	}
	// Image data offset. Always 22 in our case
	if _, err := buf.Write([]byte{22, 0, 0, 0}); err != nil {
		return nil, err
	}

	if _, err := buf.Write(b); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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

func (img *vipsImage) GetIntDefault(name string, def int) (int, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetInt(name)
}

func (img *vipsImage) GetIntSlice(name string) ([]int, error) {
	var ptr unsafe.Pointer
	size := C.int(0)

	if C.vips_image_get_array_int_go(img.VipsImage, cachedCString(name), (**C.int)(unsafe.Pointer(&ptr)), &size) != 0 {
		return nil, vipsError()
	}

	if size == 0 {
		return []int{}, nil
	}

	cOut := (*[math.MaxInt32]C.int)(ptr)[:int(size):int(size)]
	out := make([]int, int(size))

	for i, el := range cOut {
		out[i] = int(el)
	}

	return out, nil
}

func (img *vipsImage) GetIntSliceDefault(name string, def []int) ([]int, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetIntSlice(name)
}

func (img *vipsImage) SetInt(name string, value int) {
	C.vips_image_set_int(img.VipsImage, cachedCString(name), C.int(value))
}

func (img *vipsImage) SetIntSlice(name string, value []int) {
	in := make([]C.int, len(value))
	for i, el := range value {
		in[i] = C.int(el)
	}
	C.vips_image_set_array_int_go(img.VipsImage, cachedCString(name), &in[0], C.int(len(value)))
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

func (img *vipsImage) ColorAdjust(scale float64) error {
	var tmp *C.VipsImage
	if C.vips_color_adjust(img.VipsImage, &tmp, C.double(scale)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

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
	return C.vips_get_orientation(img.VipsImage)
}

func (img *vipsImage) Rotate(angle int) error {
	var tmp *C.VipsImage

	vipsAngle := (angle / 90) % 4

	if C.vips_rot_go(img.VipsImage, &tmp, C.VipsAngle(vipsAngle)) != 0 {
		return vipsError()
	}

	C.vips_autorot_remove_angle(tmp)

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

func (img *vipsImage) Trim(threshold float64, smart bool, color rgbColor, equalHor bool, equalVer bool) error {
	var tmp *C.VipsImage

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if C.vips_trim(img.VipsImage, &tmp, C.double(threshold),
		gbool(smart), C.double(color.R), C.double(color.G), C.double(color.B),
		gbool(equalHor), gbool(equalVer)) != 0 {
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

func (img *vipsImage) ImportColourProfile() error {
	var tmp *C.VipsImage

	if img.VipsImage.Coding != C.VIPS_CODING_NONE {
		return nil
	}

	if img.VipsImage.BandFmt != C.VIPS_FORMAT_UCHAR && img.VipsImage.BandFmt != C.VIPS_FORMAT_USHORT {
		return nil
	}

	if C.vips_has_embedded_icc(img.VipsImage) == 0 {
		return nil
	}

	if C.vips_icc_import_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		logWarning("Can't import ICC profile: %s", vipsError())
	}

	return nil
}

func (img *vipsImage) ExportColourProfile() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		logWarning("Can't export ICC profile: %s", vipsError())
	}

	return nil
}

func (img *vipsImage) ExportColourProfileToSRGB() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_srgb(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		logWarning("Can't export ICC profile: %s", vipsError())
	}

	return nil
}

func (img *vipsImage) TransformColourProfile() error {
	var tmp *C.VipsImage

	// Don't transform is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_transform_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		logWarning("Can't transform ICC profile: %s", vipsError())
	}

	return nil
}

func (img *vipsImage) RemoveColourProfile() error {
	var tmp *C.VipsImage

	if C.vips_icc_remove(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		logWarning("Can't remove ICC profile: %s", vipsError())
	}

	return nil
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

func (img *vipsImage) Embed(width, height int, offX, offY int, bg rgbColor, transpBg bool) error {
	var tmp *C.VipsImage

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	var bgc []C.double
	if transpBg {
		if !img.HasAlpha() {
			if C.vips_addalpha_go(img.VipsImage, &tmp) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&img.VipsImage, tmp)
		}

		bgc = []C.double{C.double(0)}
	} else {
		bgc = []C.double{C.double(bg.R), C.double(bg.G), C.double(bg.B), 1.0}
	}

	bgn := minInt(int(img.VipsImage.Bands), len(bgc))

	if C.vips_embed_go(img.VipsImage, &tmp, C.int(offX), C.int(offY), C.int(width), C.int(height), &bgc[0], C.int(bgn)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) EmbedImage(offX, offY int, sub *vipsImage) error {
	var tmp *C.VipsImage

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	if C.vips_embed_image_go(img.VipsImage, sub.VipsImage, &tmp, C.int(offX), C.int(offY), gbool(true)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) ApplyWatermark(wm *vipsImage, opacity float64) error {
	var tmp *C.VipsImage

	if C.vips_apply_watermark(img.VipsImage, wm.VipsImage, &tmp, C.double(opacity)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *vipsImage) Strip() error {
	var tmp *C.VipsImage

	if C.vips_strip(img.VipsImage, &tmp) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}
