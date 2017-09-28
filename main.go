package main

import (
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"golang.org/x/net/netutil"
)

func main() {
	// Force garbage collection
	go func() {
		for _ = range time.Tick(10 * time.Second) {
			debug.FreeOSMemory()
		}
	}()

	l, err := net.Listen("tcp", conf.Bind)
	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		Handler:        newHTTPHandler(),
		ReadTimeout:    time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(conf.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server at %s\n", conf.Bind)

	log.Fatal(s.Serve(netutil.LimitListener(l, conf.MaxClients)))
}
