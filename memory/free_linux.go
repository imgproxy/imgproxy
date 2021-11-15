//go:build linux
// +build linux

package memory

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

func Free() {
	debug.FreeOSMemory()

	C.malloc_trim(0)
}
