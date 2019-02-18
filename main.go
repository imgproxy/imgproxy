package main

import (
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"net/http"
	_ "net/http/pprof"
)

const version = "2.2.4"

type ctxKey string

func main() {
	go func() {
		var logMemStats = len(os.Getenv("IMGPROXY_LOG_MEM_STATS")) > 0

		for range time.Tick(time.Duration(conf.FreeMemoryInterval) * time.Second) {
			debug.FreeOSMemory()

			if logMemStats {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				logNotice("[MEMORY USAGE] Sys: %d; HeapIdle: %d; HeapInuse: %d", m.Sys/1024/1024, m.HeapIdle/1024/1024, m.HeapInuse/1024/1024)
			}
		}
	}()

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
