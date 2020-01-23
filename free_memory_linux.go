// +build linux

package main

/*
#include <features.h>
#ifdef __GLIBC__
#include <malloc.h>
#else
void malloc_trim(size_t pad){}
#endif
*/
import "C"
import "runtime/debug"

func freeMemory() {
	debug.FreeOSMemory()

	C.malloc_trim(0)
}
