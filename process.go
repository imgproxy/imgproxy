package main

/*
#cgo pkg-config: vips
#cgo LDFLAGS: -s -w
#include "vips.h"
*/
import "C"

import (
	"errors"
	"log"
	"math"
	"os"
	"runtime"
	"unsafe"
)

type imageType int

const (
	UNKNOWN = C.UNKNOWN
	JPEG    = C.JPEG
	PNG     = C.PNG
	WEBP    = C.WEBP
	GIF     = C.GIF
)

var imageTypes = map[string]imageType{
	"jpeg": JPEG,
	"jpg":  JPEG,
	"png":  PNG,
	"webp": WEBP,
	"gif":  GIF,
}

type gravityType int

const (
	CENTER gravityType = iota
	NORTH
	EAST
	SOUTH
	WEST
	SMART
)

var gravityTypes = map[string]gravityType{
	"ce": CENTER,
	"no": NORTH,
	"ea": EAST,
	"so": SOUTH,
	"we": WEST,
	"sm": SMART,
}

type resizeType int

const (
	FIT resizeType = iota
	FILL
	CROP
)

var resizeTypes = map[string]resizeType{
	"fit":  FIT,
	"fill": FILL,
	"crop": CROP,
}

type processingOptions struct {
	resize  resizeType
	width   int
	height  int
	gravity gravityType
	enlarge bool
	format  imageType
}

var vipsSupportSmartcrop bool
var vipsTypeSupportLoad = make(map[imageType]bool)
var vipsTypeSupportSave = make(map[imageType]bool)

func initVips() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		log.Fatalln("unable to start vips!")
	}

	C.vips_cache_set_max_mem(100 * 1024 * 1024) // 100Mb
	C.vips_cache_set_max(500)

	if len(os.Getenv("IMGPROXY_VIPS_LEAK_CHECK")) > 0 {
		C.vips_leak_set(C.gboolean(1))
	}

	if len(os.Getenv("IMGPROXY_VIPS_CACHE_TRACE")) > 0 {
		C.vips_cache_set_trace(C.gboolean(1))
	}

	vipsSupportSmartcrop = C.vips_support_smartcrop() == 1

	if int(C.vips_type_find_load_go(C.JPEG)) != 0 {
		vipsTypeSupportLoad[JPEG] = true
	}
	if int(C.vips_type_find_load_go(C.PNG)) != 0 {
		vipsTypeSupportLoad[PNG] = true
	}
	if int(C.vips_type_find_load_go(C.WEBP)) != 0 {
		vipsTypeSupportLoad[WEBP] = true
	}
	if int(C.vips_type_find_load_go(C.GIF)) != 0 {
		vipsTypeSupportLoad[GIF] = true
	}

	if int(C.vips_type_find_save_go(C.JPEG)) != 0 {
		vipsTypeSupportSave[JPEG] = true
	}
	if int(C.vips_type_find_save_go(C.PNG)) != 0 {
		vipsTypeSupportSave[PNG] = true
	}
	if int(C.vips_type_find_save_go(C.WEBP)) != 0 {
		vipsTypeSupportSave[WEBP] = true
	}
}

func shutdownVips() {
	C.vips_shutdown()
}

func randomAccessRequired(po processingOptions) int {
	if po.gravity == SMART {
		return 1
	}
	return 0
}

func round(f float64) int {
	return int(f + .5)
}

func extractMeta(img *C.VipsImage) (int, int, int, int) {
	width := int(img.Xsize)
	height := int(img.Ysize)

	angle := C.VIPS_ANGLE_D0
	flip := C.FALSE

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
		flip = C.TRUE
	}

	return width, height, angle, flip
}

func calcScale(width, height int, po processingOptions) float64 {
	if (po.width == width && po.height == height) || (po.resize != FILL && po.resize != FIT) {
		return 1
	}

	fsw, fsh, fow, foh := float64(width), float64(height), float64(po.width), float64(po.height)

	wr := fow / fsw
	hr := foh / fsh

	if po.resize == FIT {
		return math.Min(wr, hr)
	}

	return math.Max(wr, hr)
}

func calcShink(scale float64, imgtype imageType) int {
	shrink := int(1.0 / scale)

	if imgtype != JPEG {
		return shrink
	}

	switch {
	case shrink >= 16:
		return 8
	case shrink >= 8:
		return 4
	case shrink >= 4:
		return 2
	}

	return 1
}

func calcCrop(width, height int, po processingOptions) (left, top int) {
	left = (width - po.width + 1) / 2
	top = (height - po.height + 1) / 2

	if po.gravity == NORTH {
		top = 0
	}

	if po.gravity == EAST {
		left = width - po.width
	}

	if po.gravity == SOUTH {
		top = height - po.height
	}

	if po.gravity == WEST {
		left = 0
	}

	return
}

func processImage(data []byte, imgtype imageType, po processingOptions, t *timer) ([]byte, error) {
	defer keepAlive(data)

	if po.gravity == SMART && !vipsSupportSmartcrop {
		return nil, errors.New("Smart crop is not supported by used version of libvips")
	}

	err := C.int(0)

	var img *C.struct__VipsImage
	defer C.clear_image(&img)

	defer C.vips_cleanup()

	// Load the image
	err = C.vips_load_buffer(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(imgtype), 1, &img)
	if err != 0 {
		return nil, vipsError()
	}

	t.Check()

	imgWidth, imgHeight, angle, flip := extractMeta(img)

	// Ensure we won't crop out of bounds
	if !po.enlarge || po.resize == CROP {
		if imgWidth < po.width {
			po.width = imgWidth
		}

		if imgHeight < po.height {
			po.height = imgHeight
		}
	}

	if po.width != imgWidth || po.height != imgHeight {
		pCrop, pSmart := 0, 0
		pScale := 1.0

		var pLeft, pTop, pWidth, pHeight int

		if po.resize == FILL || po.resize == FIT {
			pScale = calcScale(imgWidth, imgHeight, po)
		}

		if po.resize == FILL || po.resize == CROP {
			pCrop = 1
			pWidth, pHeight = po.width, po.height

			if po.gravity == SMART {
				pSmart = 1
			} else {
				pLeft, pTop = calcCrop(round(float64(imgWidth)*pScale), round(float64(imgHeight)*pScale), po)
			}
		}

		// Do some shrink-on-load
		if pScale < 1.0 {
			if imgtype == JPEG || imgtype == WEBP {
				shrink := calcShink(pScale, imgtype)
				pScale = pScale * float64(shrink)

				var tmp *C.struct__VipsImage
				err = C.vips_load_buffer(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(imgtype), C.int(shrink), &tmp)
				if err != 0 {
					return nil, vipsError()
				}
				C.swap_and_clear(&img, tmp)
			}
		}

		err = C.vips_process_image(&img, C.double(pScale), C.gboolean(pCrop), C.gboolean(pSmart), C.int(pLeft), C.int(pTop), C.int(pWidth), C.int(pHeight), C.VipsAngle(angle), C.gboolean(flip))
		if err != 0 {
			return nil, vipsError()
		}
	}

	t.Check()

	// Finally, save
	var ptr unsafe.Pointer
	defer C.g_free_go(&ptr)

	imgsize := C.size_t(0)

	switch po.format {
	case JPEG:
		err = C.vips_jpegsave_go(img, &ptr, &imgsize, 1, C.int(conf.Quality), 0)
	case PNG:
		err = C.vips_pngsave_go(img, &ptr, &imgsize)
	case WEBP:
		err = C.vips_webpsave_go(img, &ptr, &imgsize, 1, C.int(conf.Quality))
	}
	if err != 0 {
		return nil, vipsError()
	}

	t.Check()

	buf := C.GoBytes(ptr, C.int(imgsize))

	return buf, nil
}

func vipsError() error {
	return errors.New(C.GoString(C.vips_error_buffer()))
}
