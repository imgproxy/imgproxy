package testutil

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include <vips/vips.h>
#include "image_hash.h"
*/
import "C"
import (
	"fmt"
	"image"
	"unsafe"

	"github.com/corona10/goimagehash"
)

// ImageHash calculates a hash of the VipsImage
func ImageHash(vipsImgPtr unsafe.Pointer) (*goimagehash.ImageHash, error) {
	vipsImg := (*C.VipsImage)(vipsImgPtr)

	// Convert to RGBA and read into memory using VIPS
	var data unsafe.Pointer
	var size C.size_t

	// no one knows why this triggers linter
	//nolint:gocritic
	loadErr := C.vips_image_read_to_memory(vipsImg, &data, &size)
	if loadErr != 0 {
		return nil, fmt.Errorf("failed to convert VipsImage to RGBA memory")
	}
	defer C.vips_memory_buffer_free(data)

	// Convert raw RGBA pixel data to Go image.Image
	goImg, err := createRGBAFromRGBAPixels(vipsImg, data, size)
	if err != nil {
		return nil, fmt.Errorf("failed to convert RGBA pixel data to image.Image: %v", err)
	}

	hash, err := goimagehash.DifferenceHash(goImg)
	if err != nil {
		return nil, err
	}

	return hash, err
}

// createRGBAFromRGBAPixels creates a Go image.Image from raw RGBA VIPS pixel data
func createRGBAFromRGBAPixels(vipsImg *C.VipsImage, data unsafe.Pointer, size C.size_t) (*image.RGBA, error) {
	width := int(vipsImg.Xsize)
	height := int(vipsImg.Ysize)

	// RGBA should have 4 bands
	expectedSize := width * height * 4
	if int(size) != expectedSize {
		return nil, fmt.Errorf("size mismatch: expected %d bytes for RGBA, got %d", expectedSize, int(size))
	}

	pixels := unsafe.Slice((*byte)(data), int(size))

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy RGBA pixel data directly
	copy(img.Pix, pixels)

	return img, nil
}
