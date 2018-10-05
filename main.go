package main

import (
	"os"
	"os/signal"

	_ "net/http/pprof"
)

const version = "2.0.0"

type ctxKey string

func main() {
	s := startServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	shutdownServer(s)
	shutdownVips()
}
