//go:build !linux
// +build !linux

package memory

import "runtime/debug"

func Free() {
	debug.FreeOSMemory()
}
