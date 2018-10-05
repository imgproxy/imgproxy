package main

import (
	"log"
	"os"
	"os/signal"

	"net/http"
	_ "net/http/pprof"
)

const version = "2.0.0"

type ctxKey string

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	s := startServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	shutdownServer(s)
	shutdownVips()
}
