package testutil

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include <vips/vips.h>
#include "image_loader.h"
*/
import "C"
import (
	"fmt"
	"image"
	"io"
	"strings"
	"unsafe"
)

// LoadImage loads an image from an io.Reader and converts it to RGBA format using VIPS.
// It supports all image formats that VIPS can decode (JPEG, PNG, WebP, AVIF, SVG, etc.).
// The returned image.RGBA contains the decoded pixel data ready for processing.
//
// This function uses VIPS internally to handle image decoding and color space conversion,
// making it suitable for testing image processing operations that require consistent RGBA input.
//
// Returns an error if the image cannot be read, decoded, or converted to RGBA format.
func LoadImage(r io.Reader) (*image.RGBA, error) {
	// Read all image data into memory
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Convert to RGBA using VIPS
	var data unsafe.Pointer
	var size C.size_t
	var width, height C.int

	bufPtr := unsafe.Pointer(unsafe.SliceData(buf))

	// Call VIPS function to decode and convert image to RGBA
	//nolint:gocritic
	readErr := C.vips_image_read_from_to_memory(bufPtr, C.size_t(len(buf)), &data, &size, &width, &height)
	if readErr != 0 {
		return nil, fmt.Errorf("failed to decode image with VIPS: %s", vipsErrorMessage())
	}
	defer C.g_free(C.gpointer(data))

	// Convert raw RGBA pixel data to Go image.RGBA
	img, err := createRGBAFromRGBAPixels(int(width), int(height), data, size)
	if err != nil {
		return nil, fmt.Errorf("failed to create RGBA image: %w", err)
	}

	return img, nil
}

// createRGBAFromRGBAPixels creates a Go image.RGBA from raw RGBA VIPS pixel data
func createRGBAFromRGBAPixels(width, height int, data unsafe.Pointer, size C.size_t) (*image.RGBA, error) {
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

// vipsErrorMessage reads VIPS error message
func vipsErrorMessage() string {
	defer C.vips_error_clear()
	return strings.TrimSpace(C.GoString(C.vips_error_buffer()))
}
