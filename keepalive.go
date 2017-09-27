// +build go1.7

package main

import "runtime"

func keepAlive(i interface{}) {
	runtime.KeepAlive(i)
}
