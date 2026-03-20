package testutil

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include <vips/vips.h>
*/
import "C"
import "strings"

// vipsErrorMessage reads VIPS error message
func vipsErrorMessage() string {
	defer C.vips_error_clear()
	return strings.TrimSpace(C.GoString(C.vips_error_buffer()))
}
