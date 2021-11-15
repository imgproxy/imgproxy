//go:build !linux

package memory

import "runtime/debug"

func Free() {
	debug.FreeOSMemory()
}
