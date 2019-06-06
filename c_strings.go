package main

import "C"

var cStringsCache = make(map[string]*C.char)

func cachedCString(str string) *C.char {
	if cstr, ok := cStringsCache[str]; ok {
		return cstr
	}

	cstr := C.CString(str)
	cStringsCache[str] = cstr

	return cstr
}
