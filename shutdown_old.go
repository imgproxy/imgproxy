// +build !go1.8

package main

import (
	"net/http"
)

func shutdownServer(_ *http.Server) {
	// Nothing we can do here
}
