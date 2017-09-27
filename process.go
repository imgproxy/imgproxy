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
	"runtime"
	"unsafe"
)

type imageType int

const (
	UNKNOWN imageType = iota
	JPEG
	PNG
	WEBP
	GIF
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

func initVips() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		log.Fatalln("unable to start vips!")
	}

	C.vips_concurrency_set(1)
	C.vips_cache_set_max_mem(100 * 1024 * 1024) // 100Mb
	C.vips_cache_set_max(500)

	vipsSupportSmartcrop = C.vips_support_smartcrop() == 1
}

func vipsTypeSupportedLoad(imgtype imageType) bool {
	switch imgtype {
	case JPEG:
		return int(C.vips_type_find_load_go(C.JPEG)) != 0
	case PNG:
		return int(C.vips_type_find_load_go(C.PNG)) != 0
	case WEBP:
		return int(C.vips_type_find_load_go(C.WEBP)) != 0
	case GIF:
		return int(C.vips_type_find_load_go(C.GIF)) != 0
	}
	return false
}

func vipsTypeSupportedSave(imgtype imageType) bool {
	switch imgtype {
	case JPEG:
		return int(C.vips_type_find_save_go(C.JPEG)) != 0
	case PNG:
		return int(C.vips_type_find_save_go(C.PNG)) != 0
	case WEBP:
		return int(C.vips_type_find_save_go(C.WEBP)) != 0
	}
	return false
}

func round(f float64) int {
	return int(f + .5)
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

func processImage(data []byte, imgtype imageType, po processingOptions) ([]byte, error) {
	defer keepAlive(data)

	err := C.int(0)

	var img *C.struct__VipsImage

	defer C.vips_cleanup()

	// Load the image
	switch imgtype {
	case JPEG:
		err = C.vips_jpegload_buffer_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &img)
	case PNG:
		err = C.vips_pngload_buffer_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &img)
	case GIF:
		err = C.vips_gifload_buffer_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &img)
	case WEBP:
		err = C.vips_webpload_buffer_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &img)
	}
	if err != 0 {
		return nil, vipsError()
	}

	imgWidth := int(img.Xsize)
	imgHeight := int(img.Ysize)

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
		var (
			pResize, pCrop               int
			pScale                       float64
			pSmart                       int
			pLeft, pTop, pWidth, pHeight int
		)

		if po.resize == FILL || po.resize == FIT {
			pResize = 1
			pScale = calcScale(imgWidth, imgHeight, po)
		} else {
			pScale = 1.0
		}

		if po.resize == FILL || po.resize == CROP {
			pCrop = 1
			pWidth, pHeight = po.width, po.height

			if po.gravity == SMART && vipsSupportSmartcrop {
				pSmart = 1
			} else {
				pLeft, pTop = calcCrop(round(float64(imgWidth)*pScale), round(float64(imgHeight)*pScale), po)
			}
		}

		err = C.vips_process_image(&img, C.int(pResize), C.double(pScale), C.int(pCrop), C.int(pSmart), C.int(pLeft), C.int(pTop), C.int(pWidth), C.int(pHeight))
		if err != 0 {
			return nil, vipsError()
		}
	}

	// Finally, save
	imgsize := C.size_t(0)
	var ptr unsafe.Pointer
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

	C.g_object_unref(C.gpointer(img))

	buf := C.GoBytes(ptr, C.int(imgsize))
	C.g_free(C.gpointer(ptr))

	return buf, nil
}

func vipsError() error {
	return errors.New(C.GoString(C.vips_error_buffer()))
}
