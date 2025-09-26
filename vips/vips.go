package vips

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include "vips.h"
#include "source.h"
*/
import "C"
import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips/color"
)

type Image struct {
	VipsImage *C.VipsImage
}

var (
	typeSupportLoad sync.Map
	typeSupportSave sync.Map

	gifResolutionLimit int

	initOnce sync.Once
)

// Global vips config. Can be set with [Init]
var config *Config

var badImageErrRe = []*regexp.Regexp{
	regexp.MustCompile(`^(\S+)load_source: `),
	regexp.MustCompile(`^VipsJpeg: `),
	regexp.MustCompile(`^tiff2vips: `),
	regexp.MustCompile(`^webp2vips: `),
}

func init() {
	// Just get sure that we have some config
	c := NewDefaultConfig()
	config = &c
}

func Init(c *Config) error {
	if err := c.Validate(); err != nil {
		return err
	}

	config = c

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// vips_initialize must be called only once
	var initErr error
	initOnce.Do(func() {
		if err := C.vips_initialize(); err != 0 {
			C.vips_shutdown()
			initErr = newVipsError("unable to start vips!")
		}
	})
	if initErr != nil {
		return initErr
	}

	// Disable libvips cache. Since processing pipeline is fine tuned, we won't get much profit from it.
	// Enabled cache can cause SIGSEGV on Musl-based systems like Alpine.
	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)

	if lambdaFn := os.Getenv("AWS_LAMBDA_FUNCTION_NAME"); len(lambdaFn) > 0 {
		// Set vips concurrency level to GOMAXPROCS if we are running in AWS Lambda
		// since each function processes only one request at a time
		// so we can use all available CPU cores
		C.vips_concurrency_set(C.int(max(1, runtime.GOMAXPROCS(0))))
	} else {
		C.vips_concurrency_set(1)
	}

	C.vips_leak_set(gbool(config.LeakCheck))
	C.vips_cache_set_trace(gbool(config.CacheTrace))

	gifResolutionLimit = int(C.gif_resolution_limit())

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
	err := newVipsError(errstr)

	for _, re := range badImageErrRe {
		if re.MatchString(errstr) {
			return ierrors.Wrap(
				err, 0,
				ierrors.WithStatusCode(http.StatusUnprocessableEntity),
				ierrors.WithPublicMessage("Broken or unsupported image"),
			)
		}
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
		sup = hasOperation("jpegload_source")
	case imagetype.JXL:
		sup = hasOperation("jxlload_source")
	case imagetype.PNG:
		sup = hasOperation("pngload_source")
	case imagetype.WEBP:
		sup = hasOperation("webpload_source")
	case imagetype.GIF:
		sup = hasOperation("gifload_source")
	case imagetype.BMP:
		sup = hasOperation("bmpload_source")
	case imagetype.ICO:
		sup = hasOperation("icoload_source")
	case imagetype.SVG:
		sup = hasOperation("svgload_source")
	case imagetype.HEIC, imagetype.AVIF:
		sup = hasOperation("heifload_source")
	case imagetype.TIFF:
		sup = hasOperation("tiffload_source")
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
		sup = hasOperation("jpegsave_target")
	case imagetype.JXL:
		sup = hasOperation("jxlsave_target")
	case imagetype.PNG:
		sup = hasOperation("pngsave_target")
	case imagetype.WEBP:
		sup = hasOperation("webpsave_target")
	case imagetype.GIF:
		sup = hasOperation("gifsave_target")
	case imagetype.HEIC, imagetype.AVIF:
		sup = hasOperation("heifsave_target")
	case imagetype.BMP:
		sup = hasOperation("bmpsave_target")
	case imagetype.TIFF:
		sup = hasOperation("tiffsave_target")
	case imagetype.ICO:
		sup = hasOperation("icosave_target")
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

func cRGB(c color.RGB) C.RGB {
	return C.RGB{
		r: C.double(c.R),
		g: C.double(c.G),
		b: C.double(c.B),
	}
}

func ptrToBytes(ptr unsafe.Pointer, size int) []byte {
	return (*[math.MaxInt32]byte)(ptr)[:int(size):int(size)]
}

func (img *Image) swapAndUnref(newImg *C.VipsImage) {
	if img.VipsImage != nil {
		C.unref_image(img.VipsImage)
	}
	img.VipsImage = newImg
}

func (img *Image) Width() int {
	return int(img.VipsImage.Xsize)
}

func (img *Image) Height() int {
	return int(img.VipsImage.Ysize)
}

func (img *Image) PageHeight() int {
	return int(C.vips_image_get_page_height(img.VipsImage))
}

// Pages returns number of pages in the image file.
//
// WARNING: It's not the number of pages in the loaded image.
// Use [Image.PagesLoaded] for that.
func (img *Image) Pages() int {
	p, err := img.GetIntDefault("n-pages", 1)
	if err != nil {
		return 1
	}
	return p
}

// PagesLoaded returns number of pages in the loaded image.
func (img *Image) PagesLoaded() int {
	return img.Height() / img.PageHeight()
}

func (img *Image) Load(
	imgdata imagedata.ImageData,
	shrink float64,
	page, pages int,
) error {
	var tmp *C.VipsImage

	source := newVipsImgproxySource(imgdata.Reader())
	defer C.unref_imgproxy_source(source)

	lo := newLoadOptions(shrink, page, pages)

	err := C.int(0)

	switch imgdata.Format() {
	case imagetype.JPEG:
		err = C.vips_jpegload_source_go(source, &tmp, lo)
	case imagetype.JXL:
		err = C.vips_jxlload_source_go(source, &tmp, lo)
	case imagetype.PNG:
		err = C.vips_pngload_source_go(source, &tmp, lo)
	case imagetype.WEBP:
		err = C.vips_webpload_source_go(source, &tmp, lo)
	case imagetype.GIF:
		err = C.vips_gifload_source_go(source, &tmp, lo)
	case imagetype.SVG:
		err = C.vips_svgload_source_go(source, &tmp, lo)
	case imagetype.HEIC, imagetype.AVIF:
		err = C.vips_heifload_source_go(source, &tmp, lo)
	case imagetype.TIFF:
		err = C.vips_tiffload_source_go(source, &tmp, lo)
	case imagetype.BMP:
		err = C.vips_bmpload_source_go(source, &tmp, lo)
	case imagetype.ICO:
		err = C.vips_icoload_source_go(source, &tmp, lo)
	default:
		return newVipsError("Usupported image type to load")
	}
	if err != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)

	if imgdata.Format() == imagetype.TIFF {
		if C.vips_fix_float_tiff(img.VipsImage, &tmp) == 0 {
			img.swapAndUnref(tmp)
		} else {
			slog.Warn("Can't fix TIFF", "error", Error())
		}
	}

	return nil
}

func (img *Image) LoadThumbnail(imgdata imagedata.ImageData) error {
	if imgdata.Format() != imagetype.HEIC && imgdata.Format() != imagetype.AVIF {
		return newVipsError("Usupported image type to load thumbnail")
	}

	var tmp *C.VipsImage

	source := newVipsImgproxySource(imgdata.Reader())
	defer C.unref_imgproxy_source(source)

	lo := newLoadOptions(1.0, 0, 1)
	lo.Thumbnail = 1

	if err := C.vips_heifload_source_go(source, &tmp, lo); err != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)

	return nil
}

func (img *Image) Save(
	imgtype imagetype.Type,
	quality int,
	o *options.Options,
) (imagedata.ImageData, error) {
	target := C.vips_target_new_to_memory()

	cancel := func() {
		C.vips_unref_target(target)
	}

	so := newSaveOptions(o)

	err := C.int(0)
	imgsize := C.size_t(0)

	switch imgtype {
	case imagetype.JPEG:
		err = C.vips_jpegsave_go(img.VipsImage, target, C.int(quality), so)
	case imagetype.JXL:
		err = C.vips_jxlsave_go(img.VipsImage, target, C.int(quality), so)
	case imagetype.PNG:
		err = C.vips_pngsave_go(img.VipsImage, target, so)
	case imagetype.WEBP:
		err = C.vips_webpsave_go(img.VipsImage, target, C.int(quality), so)
	case imagetype.GIF:
		err = C.vips_gifsave_go(img.VipsImage, target, so)
	case imagetype.HEIC:
		err = C.vips_heifsave_go(img.VipsImage, target, C.int(quality), so)
	case imagetype.AVIF:
		err = C.vips_avifsave_go(img.VipsImage, target, C.int(quality), so)
	case imagetype.TIFF:
		err = C.vips_tiffsave_go(img.VipsImage, target, C.int(quality), so)
	case imagetype.BMP:
		err = C.vips_bmpsave_target_go(img.VipsImage, target, so)
	case imagetype.ICO:
		err = C.vips_icosave_target_go(img.VipsImage, target, so)
	default:
		// NOTE: probably, it would be better to use defer unref + additionally ref the target
		// before passing it to the imagedata.ImageData
		cancel()
		return nil, newVipsError("Usupported image type to save")
	}
	if err != 0 {
		cancel()
		return nil, Error()
	}

	var blob_ptr = C.vips_blob_get(target.blob, &imgsize)
	var ptr unsafe.Pointer = unsafe.Pointer(blob_ptr)

	b := ptrToBytes(ptr, int(imgsize))

	i := imagedata.NewFromBytesWithFormat(imgtype, b)
	i.AddCancel(cancel)

	return i, nil
}

func (img *Image) Clear() {
	if img.VipsImage != nil {
		C.unref_image(img.VipsImage)
		img.VipsImage = nil
	}
}

func (img *Image) LineCache(lines int) error {
	var tmp *C.VipsImage

	if C.vips_linecache_seq(img.VipsImage, &tmp, C.int(lines)) != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)
	return nil
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

	img.swapAndUnref(tmp)
	return nil
}

func (img *Image) Swap(in *Image) {
	img.VipsImage, in.VipsImage = in.VipsImage, img.VipsImage
}

func (img *Image) IsAnimated() bool {
	return C.vips_image_is_animated(img.VipsImage) > 0
}

// RemoveAnimation removes all animation-related data from the image
// making it a static image.
//
// It doesn't remove already loaded frames and keeps them vertically stacked.
func (img *Image) RemoveAnimation() error {
	var tmp *C.VipsImage

	if C.vips_image_remove_animation(img.VipsImage, &tmp) != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)
	return nil
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

func (img *Image) CastUchar() error {
	var tmp *C.VipsImage

	if C.vips_image_get_format(img.VipsImage) != C.VIPS_FORMAT_UCHAR {
		if C.vips_cast_go(img.VipsImage, &tmp, C.VIPS_FORMAT_UCHAR) != 0 {
			return Error()
		}
		img.swapAndUnref(tmp)
	}

	return nil
}

func (img *Image) Rad2Float() error {
	var tmp *C.VipsImage

	if C.vips_image_get_coding(img.VipsImage) == C.VIPS_CODING_RAD {
		if C.vips_rad2float_go(img.VipsImage, &tmp) != 0 {
			return Error()
		}
		img.swapAndUnref(tmp)
	}

	return nil
}

func (img *Image) Resize(wscale, hscale float64) error {
	var tmp *C.VipsImage

	if C.vips_resize_go(img.VipsImage, &tmp, C.double(wscale), C.double(hscale)) != 0 {
		return Error()
	}

	if wscale < 1.0 || hscale < 1.0 {
		C.vips_image_set_int(tmp, cachedCString("imgproxy-scaled-down"), 1)
	}

	img.swapAndUnref(tmp)

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

	img.swapAndUnref(tmp)
	return nil
}

func (img *Image) Flip() error {
	var tmp *C.VipsImage

	if C.vips_flip_horizontal_go(img.VipsImage, &tmp) != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)
	return nil
}

func (img *Image) Crop(left, top, width, height int) error {
	var tmp *C.VipsImage

	if C.vips_extract_area_go(img.VipsImage, &tmp, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)
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

	img.swapAndUnref(tmp)
	return nil
}

func (img *Image) Trim(threshold float64, smart bool, color color.RGB, equalHor bool, equalVer bool) error {
	var tmp *C.VipsImage

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if C.vips_trim(img.VipsImage, &tmp, C.double(threshold),
		gbool(smart), cRGB(color), gbool(equalHor), gbool(equalVer)) != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)
	return nil
}

func (img *Image) Flatten(bg color.RGB) error {
	var tmp *C.VipsImage

	if C.vips_flatten_go(img.VipsImage, &tmp, cRGB(bg)) != 0 {
		return Error()
	}
	img.swapAndUnref(tmp)

	return nil
}

func (img *Image) ApplyFilters(blurSigma, sharpSigma float64, pixelatePixels int) error {
	var tmp *C.VipsImage

	if C.vips_apply_filters(img.VipsImage, &tmp, C.double(blurSigma), C.double(sharpSigma), C.int(pixelatePixels)) != 0 {
		return Error()
	}

	img.swapAndUnref(tmp)

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

func (img *Image) BackupColourProfile() {
	var tmp *C.VipsImage

	if C.vips_icc_backup(img.VipsImage, &tmp) == 0 {
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't backup ICC profile", "error", Error())
	}
}

func (img *Image) RestoreColourProfile() {
	var tmp *C.VipsImage

	if C.vips_icc_restore(img.VipsImage, &tmp) == 0 {
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't restore ICC profile", "error", Error())
	}
}

func (img *Image) ImportColourProfile() error {
	var tmp *C.VipsImage

	if img.VipsImage.Coding != C.VIPS_CODING_NONE {
		return nil
	}

	if img.VipsImage.BandFmt != C.VIPS_FORMAT_UCHAR && img.VipsImage.BandFmt != C.VIPS_FORMAT_USHORT {
		return nil
	}

	// Don't import is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 ||
		(C.vips_image_guess_interpretation(img.VipsImage) == C.VIPS_INTERPRETATION_sRGB &&
			C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1) {
		return nil
	}

	if C.vips_icc_import_go(img.VipsImage, &tmp) == 0 {
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't import ICC profile", "error", Error())
	}

	return nil
}

func (img *Image) ColourProfileImported() bool {
	imported, err := img.GetIntDefault("imgproxy-icc-imported", 0)
	return imported > 0 && err == nil
}

func (img *Image) ExportColourProfile() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_go(img.VipsImage, &tmp) == 0 {
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't export ICC profile", "error", Error())
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
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't export ICC profile", "error", Error())
	}

	return nil
}

func (img *Image) TransformColourProfileToSRGB() error {
	var tmp *C.VipsImage

	// Don't transform is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 ||
		(C.vips_image_guess_interpretation(img.VipsImage) == C.VIPS_INTERPRETATION_sRGB &&
			C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1) {
		return nil
	}

	if C.vips_icc_transform_srgb(img.VipsImage, &tmp) == 0 {
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't transform ICC profile to sRGB", "error", Error())
	}

	return nil
}

func (img *Image) RemoveColourProfile() error {
	var tmp *C.VipsImage

	if C.vips_icc_remove(img.VipsImage, &tmp) == 0 {
		img.swapAndUnref(tmp)
	} else {
		slog.Warn("Can't remove ICC profile", "error", Error())
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
		img.swapAndUnref(tmp)
	}

	return nil
}

func (img *Image) CopyMemory() error {
	var tmp *C.VipsImage
	if tmp = C.vips_image_copy_memory(img.VipsImage); tmp == nil {
		return Error()
	}
	img.swapAndUnref(tmp)
	return nil
}

func (img *Image) Replicate(width, height int, centered bool) error {
	var tmp *C.VipsImage

	if C.vips_replicate_go(img.VipsImage, &tmp, C.int(width), C.int(height), gbool(centered)) != 0 {
		return Error()
	}
	img.swapAndUnref(tmp)

	return nil
}

func (img *Image) Embed(width, height int, offX, offY int) error {
	var tmp *C.VipsImage

	if C.vips_embed_go(img.VipsImage, &tmp, C.int(offX), C.int(offY), C.int(width), C.int(height)) != 0 {
		return Error()
	}
	img.swapAndUnref(tmp)

	return nil
}

func (img *Image) ApplyWatermark(wm *Image, left, top int, opacity float64) error {
	var tmp *C.VipsImage

	if C.vips_apply_watermark(img.VipsImage, wm.VipsImage, &tmp, C.int(left), C.int(top), C.double(opacity)) != 0 {
		return Error()
	}
	img.swapAndUnref(tmp)

	return nil
}

func (img *Image) Strip(keepExifCopyright bool) error {
	var tmp *C.VipsImage

	if C.vips_strip(img.VipsImage, &tmp, gbool(keepExifCopyright)) != 0 {
		return Error()
	}
	img.swapAndUnref(tmp)

	return nil
}

func (img *Image) StripAll() error {
	var tmp *C.VipsImage

	if C.vips_strip_all(img.VipsImage, &tmp) != 0 {
		return Error()
	}
	img.swapAndUnref(tmp)

	return nil
}

func vipsError(fn string, msg string, args ...any) {
	fnStr := C.CString(fn)
	defer C.free(unsafe.Pointer(fnStr))

	msg = fmt.Sprintf(msg, args...)

	msgStr := C.CString(msg)
	defer C.free(unsafe.Pointer(msgStr))

	C.vips_error_go(fnStr, msgStr)
}
