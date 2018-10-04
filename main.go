package main

import (
	"os"
	"os/signal"
	"runtime/debug"
	"time"
)

const version = "1.1.7"

func main() {
	// Force garbage collection
	go func() {
		for range time.Tick(10 * time.Second) {
			debug.FreeOSMemory()
		}
	}()

	s := startServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	shutdownServer(s)
	shutdownVips()
}
