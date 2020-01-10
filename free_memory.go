// +build !linux

package main

import "runtime/debug"

func freeMemory() {
	debug.FreeOSMemory()
}
