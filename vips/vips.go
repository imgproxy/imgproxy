package vips

/*
#cgo pkg-config: vips
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
	"strings"
	"sync"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/metrics/cloudwatch"
	"github.com/imgproxy/imgproxy/v3/metrics/datadog"
	"github.com/imgproxy/imgproxy/v3/metrics/newrelic"
	"github.com/imgproxy/imgproxy/v3/metrics/otel"
	"github.com/imgproxy/imgproxy/v3/metrics/prometheus"
)

type Image struct {
	VipsImage *C.VipsImage
}

var (
	typeSupportLoad sync.Map
	typeSupportSave sync.Map

	gifResolutionLimit int
)

var vipsConf struct {
	JpegProgressive       C.int
	PngInterlaced         C.int
	PngQuantize           C.int
	PngQuantizationColors C.int
	AvifSpeed             C.int
}

func Init() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		return errors.New("unable to start vips!")
	}

	// Disable libvips cache. Since processing pipeline is fine tuned, we won't get much profit from it.
	// Enabled cache can cause SIGSEGV on Musl-based systems like Alpine.
	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)

	C.vips_concurrency_set(1)

	C.vips_vector_set_enabled(1)

	if len(os.Getenv("IMGPROXY_VIPS_LEAK_CHECK")) > 0 {
		C.vips_leak_set(C.gboolean(1))
	}

	if len(os.Getenv("IMGPROXY_VIPS_CACHE_TRACE")) > 0 {
		C.vips_cache_set_trace(C.gboolean(1))
	}

	gifResolutionLimit = int(C.gif_resolution_limit())

	vipsConf.JpegProgressive = gbool(config.JpegProgressive)
	vipsConf.PngInterlaced = gbool(config.PngInterlaced)
	vipsConf.PngQuantize = gbool(config.PngQuantize)
	vipsConf.PngQuantizationColors = C.int(config.PngQuantizationColors)
	vipsConf.AvifSpeed = C.int(config.AvifSpeed)

	prometheus.AddGaugeFunc(
		"vips_memory_bytes",
		"A gauge of the vips tracked memory usage in bytes.",
		GetMem,
	)
	prometheus.AddGaugeFunc(
		"vips_max_memory_bytes",
		"A gauge of the max vips tracked memory usage in bytes.",
		GetMemHighwater,
	)
	prometheus.AddGaugeFunc(
		"vips_allocs",
		"A gauge of the number of active vips allocations.",
		GetAllocs,
	)

	datadog.AddGaugeFunc("vips.memory", GetMem)
	datadog.AddGaugeFunc("vips.max_memory", GetMemHighwater)
	datadog.AddGaugeFunc("vips.allocs", GetAllocs)

	newrelic.AddGaugeFunc("vips.memory", GetMem)
	newrelic.AddGaugeFunc("vips.max_memory", GetMemHighwater)
	newrelic.AddGaugeFunc("vips.allocs", GetAllocs)

	otel.AddGaugeFunc(
		"vips_memory_bytes",
		"A gauge of the vips tracked memory usage in bytes.",
		"By",
		GetMem,
	)
	otel.AddGaugeFunc(
		"vips_max_memory_bytes",
		"A gauge of the max vips tracked memory usage in bytes.",
		"By",
		GetMemHighwater,
	)
	otel.AddGaugeFunc(
		"vips_allocs",
		"A gauge of the number of active vips allocations.",
		"By",
		GetAllocs,
	)

	cloudwatch.AddGaugeFunc("VipsMemory", "Bytes", GetMem)
	cloudwatch.AddGaugeFunc("VipsMaxMemory", "Bytes", GetMemHighwater)
	cloudwatch.AddGaugeFunc("VipsAllocs", "Count", GetAllocs)

	return nil
}

func Shutdown() {
	C.vips_shutdown()
}

func GetMem() float64 {
	return float64(C.vips_tracked_get_mem())
}

func GetMemHighwater() float64 {
	return float64(C.vips_tracked_get_mem_highwater())
}

func GetAllocs() float64 {
	return float64(C.vips_tracked_get_allocs())
}

func Health() error {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	done := make(chan struct{})

	var err error

	go func(done chan struct{}) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		defer Cleanup()

		if C.vips_health() != 0 {
			err = Error()
		}

		close(done)
	}(done)

	select {
	case <-done:
		return err
	case <-timer.C:
		return context.DeadlineExceeded
	}
}

func Cleanup() {
	C.vips_cleanup()
}

func Error() error {
	defer C.vips_error_clear()

	errstr := strings.TrimSpace(C.GoString(C.vips_error_buffer()))
	err := ierrors.NewUnexpected(errstr, 1)

	if strings.Contains(errstr, "load_buffer: ") {
		err.StatusCode = 422
		err.PublicMessage = "Broken or unsupported image"
	}

	return err
}

func hasOperation(name string) bool {
	return C.vips_type_find(cachedCString("VipsOperation"), cachedCString(name)) != 0
}

func SupportsLoad(it imagetype.Type) bool {
	if sup, ok := typeSupportLoad.Load(it); ok {
		return sup.(bool)
	}

	sup := false

	switch it {
	case imagetype.JPEG:
		sup = hasOperation("jpegload_buffer")
	case imagetype.PNG:
		sup = hasOperation("pngload_buffer")
	case imagetype.WEBP:
		sup = hasOperation("webpload_buffer")
	case imagetype.GIF:
		sup = hasOperation("gifload_buffer")
	case imagetype.ICO, imagetype.BMP:
		sup = true
	case imagetype.SVG:
		sup = hasOperation("svgload_buffer")
	case imagetype.HEIC, imagetype.AVIF:
		sup = hasOperation("heifload_buffer")
	case imagetype.TIFF:
		sup = hasOperation("tiffload_buffer")
	}

	typeSupportLoad.Store(it, sup)

	return sup
}

func SupportsSave(it imagetype.Type) bool {
	if sup, ok := typeSupportSave.Load(it); ok {
		return sup.(bool)
	}

	sup := false

	switch it {
	case imagetype.JPEG:
		sup = hasOperation("jpegsave_buffer")
	case imagetype.PNG, imagetype.ICO:
		sup = hasOperation("pngsave_buffer")
	case imagetype.WEBP:
		sup = hasOperation("webpsave_buffer")
	case imagetype.GIF:
		sup = hasOperation("gifsave_buffer")
	case imagetype.AVIF:
		sup = hasOperation("heifsave_buffer")
	case imagetype.BMP:
		sup = true
	case imagetype.TIFF:
		sup = hasOperation("tiffsave_buffer")
	}

	typeSupportSave.Store(it, sup)

	return sup
}

func GifResolutionLimit() int {
	return gifResolutionLimit
}

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func ptrToBytes(ptr unsafe.Pointer, size int) []byte {
	return (*[math.MaxInt32]byte)(ptr)[:int(size):int(size)]
}

func (img *Image) Width() int {
	return int(img.VipsImage.Xsize)
}

func (img *Image) Height() int {
	return int(img.VipsImage.Ysize)
}

func (img *Image) Load(imgdata *imagedata.ImageData, shrink int, scale float64, pages int) error {
	if imgdata.Type == imagetype.ICO {
		return img.loadIco(imgdata.Data, shrink, scale, pages)
	}

	if imgdata.Type == imagetype.BMP {
		return img.loadBmp(imgdata.Data, true)
	}

	var tmp *C.VipsImage

	data := unsafe.Pointer(&imgdata.Data[0])
	dataSize := C.size_t(len(imgdata.Data))
	err := C.int(0)

	switch imgdata.Type {
	case imagetype.JPEG:
		err = C.vips_jpegload_go(data, dataSize, C.int(shrink), &tmp)
	case imagetype.PNG:
		err = C.vips_pngload_go(data, dataSize, &tmp)
	case imagetype.WEBP:
		err = C.vips_webpload_go(data, dataSize, C.double(scale), C.int(pages), &tmp)
	case imagetype.GIF:
		err = C.vips_gifload_go(data, dataSize, C.int(pages), &tmp)
	case imagetype.SVG:
		err = C.vips_svgload_go(data, dataSize, C.double(scale), &tmp)
	case imagetype.HEIC, imagetype.AVIF:
		err = C.vips_heifload_go(data, dataSize, &tmp, C.int(0))
	case imagetype.TIFF:
		err = C.vips_tiffload_go(data, dataSize, &tmp)
	default:
		return errors.New("Usupported image type to load")
	}
	if err != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) LoadThumbnail(imgdata *imagedata.ImageData) error {
	if imgdata.Type != imagetype.HEIC && imgdata.Type != imagetype.AVIF {
		return errors.New("Usupported image type to load thumbnail")
	}

	var tmp *C.VipsImage

	data := unsafe.Pointer(&imgdata.Data[0])
	dataSize := C.size_t(len(imgdata.Data))

	if err := C.vips_heifload_go(data, dataSize, &tmp, C.int(1)); err != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Save(imgtype imagetype.Type, quality int) (*imagedata.ImageData, error) {
	if imgtype == imagetype.ICO {
		return img.saveAsIco()
	}

	if imgtype == imagetype.BMP {
		return img.saveAsBmp()
	}

	var ptr unsafe.Pointer
	cancel := func() {
		C.g_free_go(&ptr)
	}

	err := C.int(0)
	imgsize := C.size_t(0)

	switch imgtype {
	case imagetype.JPEG:
		err = C.vips_jpegsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), vipsConf.JpegProgressive)
	case imagetype.PNG:
		err = C.vips_pngsave_go(img.VipsImage, &ptr, &imgsize, vipsConf.PngInterlaced, vipsConf.PngQuantize, vipsConf.PngQuantizationColors)
	case imagetype.WEBP:
		err = C.vips_webpsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality))
	case imagetype.GIF:
		err = C.vips_gifsave_go(img.VipsImage, &ptr, &imgsize)
	case imagetype.AVIF:
		err = C.vips_avifsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), vipsConf.AvifSpeed)
	case imagetype.TIFF:
		err = C.vips_tiffsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality))
	default:
		return nil, errors.New("Usupported image type to save")
	}
	if err != 0 {
		cancel()
		return nil, Error()
	}

	imgdata := imagedata.ImageData{
		Type: imgtype,
		Data: ptrToBytes(ptr, int(imgsize)),
	}

	imgdata.SetCancel(cancel)

	return &imgdata, nil
}

func (img *Image) Clear() {
	if img.VipsImage != nil {
		C.clear_image(&img.VipsImage)
	}
}

func (img *Image) Arrayjoin(in []*Image) error {
	var tmp *C.VipsImage

	arr := make([]*C.VipsImage, len(in))
	for i, im := range in {
		arr[i] = im.VipsImage
	}

	if C.vips_arrayjoin_go(&arr[0], &tmp, C.int(len(arr))) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Swap(in *Image) {
	img.VipsImage, in.VipsImage = in.VipsImage, img.VipsImage
}

func (img *Image) IsAnimated() bool {
	return C.vips_is_animated(img.VipsImage) > 0
}

func (img *Image) HasAlpha() bool {
	return C.vips_image_hasalpha(img.VipsImage) > 0
}

func (img *Image) GetInt(name string) (int, error) {
	var i C.int

	if C.vips_image_get_int(img.VipsImage, cachedCString(name), &i) != 0 {
		return 0, Error()
	}
	return int(i), nil
}

func (img *Image) GetIntDefault(name string, def int) (int, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetInt(name)
}

func (img *Image) GetIntSlice(name string) ([]int, error) {
	var ptr unsafe.Pointer
	size := C.int(0)

	if C.vips_image_get_array_int_go(img.VipsImage, cachedCString(name), (**C.int)(unsafe.Pointer(&ptr)), &size) != 0 {
		return nil, Error()
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

func (img *Image) GetIntSliceDefault(name string, def []int) ([]int, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetIntSlice(name)
}

func (img *Image) GetDouble(name string) (float64, error) {
	var d C.double

	if C.vips_image_get_double(img.VipsImage, cachedCString(name), &d) != 0 {
		return 0, Error()
	}
	return float64(d), nil
}

func (img *Image) GetDoubleDefault(name string, def float64) (float64, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetDouble(name)
}

func (img *Image) GetBlob(name string) ([]byte, error) {
	var (
		tmp  unsafe.Pointer
		size C.size_t
	)

	if C.vips_image_get_blob(img.VipsImage, cachedCString(name), &tmp, &size) != 0 {
		return nil, Error()
	}
	return C.GoBytes(tmp, C.int(size)), nil
}

func (img *Image) SetInt(name string, value int) {
	C.vips_image_set_int(img.VipsImage, cachedCString(name), C.int(value))
}

func (img *Image) SetIntSlice(name string, value []int) {
	in := make([]C.int, len(value))
	for i, el := range value {
		in[i] = C.int(el)
	}
	C.vips_image_set_array_int_go(img.VipsImage, cachedCString(name), &in[0], C.int(len(value)))
}

func (img *Image) SetDouble(name string, value float64) {
	C.vips_image_set_double(img.VipsImage, cachedCString(name), C.double(value))
}

func (img *Image) SetBlob(name string, value []byte) {
	defer runtime.KeepAlive(value)
	C.vips_image_set_blob_copy(img.VipsImage, cachedCString(name), unsafe.Pointer(&value[0]), C.size_t(len(value)))
}

func (img *Image) RemoveBitsPerSampleHeader() {
	C.vips_remove_bits_per_sample(img.VipsImage)
}

func (img *Image) CastUchar() error {
	var tmp *C.VipsImage

	if C.vips_image_get_format(img.VipsImage) != C.VIPS_FORMAT_UCHAR {
		if C.vips_cast_go(img.VipsImage, &tmp, C.VIPS_FORMAT_UCHAR) != 0 {
			return Error()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *Image) Rad2Float() error {
	var tmp *C.VipsImage

	if C.vips_image_get_coding(img.VipsImage) == C.VIPS_CODING_RAD {
		if C.vips_rad2float_go(img.VipsImage, &tmp) != 0 {
			return Error()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *Image) Resize(wscale, hscale float64) error {
	var tmp *C.VipsImage

	if C.vips_resize_go(img.VipsImage, &tmp, C.double(wscale), C.double(hscale)) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Orientation() C.int {
	return C.vips_get_orientation(img.VipsImage)
}

func (img *Image) Rotate(angle int) error {
	var tmp *C.VipsImage

	vipsAngle := (angle / 90) % 4

	if C.vips_rot_go(img.VipsImage, &tmp, C.VipsAngle(vipsAngle)) != 0 {
		return Error()
	}

	C.vips_autorot_remove_angle(tmp)

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Flip() error {
	var tmp *C.VipsImage

	if C.vips_flip_horizontal_go(img.VipsImage, &tmp) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Crop(left, top, width, height int) error {
	var tmp *C.VipsImage

	if C.vips_extract_area_go(img.VipsImage, &tmp, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Extract(out *Image, left, top, width, height int) error {
	if C.vips_extract_area_go(img.VipsImage, &out.VipsImage, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return Error()
	}
	return nil
}

func (img *Image) SmartCrop(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_smartcrop_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Trim(threshold float64, smart bool, color Color, equalHor bool, equalVer bool) error {
	var tmp *C.VipsImage

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if C.vips_trim(img.VipsImage, &tmp, C.double(threshold),
		gbool(smart), C.double(color.R), C.double(color.G), C.double(color.B),
		gbool(equalHor), gbool(equalVer)) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Flatten(bg Color) error {
	var tmp *C.VipsImage

	if C.vips_flatten_go(img.VipsImage, &tmp, C.double(bg.R), C.double(bg.G), C.double(bg.B)) != 0 {
		return Error()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) ApplyFilters(blurSigma, sharpSigma float32, pixelatePixels int) error {
	var tmp *C.VipsImage

	if C.vips_apply_filters(img.VipsImage, &tmp, C.double(blurSigma), C.double(sharpSigma), C.int(pixelatePixels)) != 0 {
		return Error()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) IsRGB() bool {
	format := C.vips_image_guess_interpretation(img.VipsImage)
	return format == C.VIPS_INTERPRETATION_sRGB ||
		format == C.VIPS_INTERPRETATION_scRGB ||
		format == C.VIPS_INTERPRETATION_RGB16
}

func (img *Image) IsLinear() bool {
	return C.vips_image_guess_interpretation(img.VipsImage) == C.VIPS_INTERPRETATION_scRGB
}

func (img *Image) ImportColourProfile() error {
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
		log.Warningf("Can't import ICC profile: %s", Error())
	}

	return nil
}

func (img *Image) ExportColourProfile() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		log.Warningf("Can't export ICC profile: %s", Error())
	}

	return nil
}

func (img *Image) ExportColourProfileToSRGB() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_srgb(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		log.Warningf("Can't export ICC profile: %s", Error())
	}

	return nil
}

func (img *Image) TransformColourProfile() error {
	var tmp *C.VipsImage

	// Don't transform is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_transform_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		log.Warningf("Can't transform ICC profile: %s", Error())
	}

	return nil
}

func (img *Image) RemoveColourProfile() error {
	var tmp *C.VipsImage

	if C.vips_icc_remove(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		log.Warningf("Can't remove ICC profile: %s", Error())
	}

	return nil
}

func (img *Image) LinearColourspace() error {
	return img.Colorspace(C.VIPS_INTERPRETATION_scRGB)
}

func (img *Image) RgbColourspace() error {
	return img.Colorspace(C.VIPS_INTERPRETATION_sRGB)
}

func (img *Image) Colorspace(colorspace C.VipsInterpretation) error {
	if img.VipsImage.Type != colorspace {
		var tmp *C.VipsImage

		if C.vips_colourspace_go(img.VipsImage, &tmp, colorspace) != 0 {
			return Error()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *Image) CopyMemory() error {
	var tmp *C.VipsImage
	if tmp = C.vips_image_copy_memory(img.VipsImage); tmp == nil {
		return Error()
	}
	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Replicate(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_replicate_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return Error()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Embed(width, height int, offX, offY int) error {
	var tmp *C.VipsImage

	if C.vips_embed_go(img.VipsImage, &tmp, C.int(offX), C.int(offY), C.int(width), C.int(height)) != 0 {
		return Error()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) ApplyWatermark(wm *Image, left, top int, opacity float64) error {
	var tmp *C.VipsImage

	if C.vips_apply_watermark(img.VipsImage, wm.VipsImage, &tmp, C.int(left), C.int(top), C.double(opacity)) != 0 {
		return Error()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Strip(keepExifCopyright bool) error {
	var tmp *C.VipsImage

	if C.vips_strip(img.VipsImage, &tmp, gbool(keepExifCopyright)) != 0 {
		return Error()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}
