// +build pprof

package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
)

func init() {
	bind := os.Getenv("IMGPROXY_PPROF_BIND")

	if len(bind) == 0 {
		bind = ":8088"
	}

	go func() {
		http.ListenAndServe(bind, nil)
	}()
}
