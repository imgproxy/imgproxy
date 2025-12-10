package vips

import "C"
import "sync"

var cStringsCache sync.Map

func cachedCString(str string) *C.char {
	if cstr, ok := cStringsCache.Load(str); ok {
		return cstr.(*C.char) //nolint:forcetypeassert
	}

	cstr := C.CString(str)
	cStringsCache.Store(str, cstr)

	return cstr
}
