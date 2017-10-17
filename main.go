package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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
		MaxHeaderBytes: 1 << 20,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	go func() {
		log.Printf("Starting server at %s\n", conf.Bind)
		log.Fatal(s.Serve(netutil.LimitListener(l, conf.MaxClients)))
	}()

	<-stop

	shutdownVips()
	shutdownServer(s)
}
