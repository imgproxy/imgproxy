package main

import (
	"os"
	"os/signal"

	"net/http"
	_ "net/http/pprof"
)

const version = "2.1.0.beta2"

type ctxKey string

func main() {
	if len(os.Getenv("IMGPROXY_PPROF_BIND")) > 0 {
		go func() {
			http.ListenAndServe(os.Getenv("IMGPROXY_PPROF_BIND"), nil)
		}()
	}

	s := startServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	shutdownServer(s)
	shutdownVips()
}
