package vips

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include "source.h"
#include "vips.h"
*/
import "C"
import (
	"io"
	"runtime/cgo"
	"unsafe"
)

// newVipsSource creates a new VipsAsyncSource from an io.ReadSeeker.
func newVipsImgproxySource(r io.ReadSeeker) *C.VipsImgproxySource {
	handler := cgo.NewHandle(r)
	return C.vips_new_imgproxy_source(C.uintptr_t(handler))
}

//export closeImgproxyReader
func closeImgproxyReader(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	h.Delete()
}

// calls seek() on the async reader via it's handle from the C side
//
//export imgproxyReaderSeek
func imgproxyReaderSeek(handle C.uintptr_t, offset C.int64_t, whence int) C.int64_t {
	h := cgo.Handle(handle)
	r, ok := h.Value().(io.ReadSeeker)
	if !ok {
		vipsError("imgproxyReaderSeek", "failed to cast handle to *source")
		return -1
	}

	pos, err := r.Seek(int64(offset), whence)
	if err != nil {
		vipsError("imgproxyReaderSeek", "failed to seek: %v", err)
		return -1
	}

	return C.int64_t(pos)
}

// calls read() on the async reader via it's handle from the C side
//
//export imgproxyReaderRead
func imgproxyReaderRead(handle C.uintptr_t, pointer unsafe.Pointer, size C.int64_t) C.int64_t {
	h := cgo.Handle(handle)
	r, ok := h.Value().(io.ReadSeeker)
	if !ok {
		vipsError("imgproxyReaderRead", "invalid reader handle")
		return -1
	}

	buf := unsafe.Slice((*byte)(pointer), size)
	n, err := r.Read(buf)
	if err == io.EOF {
		return 0
	} else if err != nil {
		vipsError("imgproxyReaderRead", "error reading from imgproxy source: %v", err)
		return -1
	}

	return C.int64_t(n)
}
