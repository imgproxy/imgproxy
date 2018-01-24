// Package common contains the common dictionary used by the enc and dec packages
package common // import "github.com/itchio/go-brotli/common"

/*
#cgo CFLAGS: -I${SRCDIR}/../include

#include "dictionary.h"
*/
import "C"

import "unsafe"

// GetDictionary retrieves a pointer to the dictionary data structure
func GetDictionary() unsafe.Pointer {
	return unsafe.Pointer(C.BrotliGetDictionary())
}
