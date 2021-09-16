package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"github.com/joho/godotenv"
)

const version = "2.16.7"

type ctxKey string

func initialize() error {
	log.SetOutput(os.Stdout)

	if err := initLog(); err != nil {
		return err
	}

	if err := configure(); err != nil {
		return err
	}

	if err := initNewrelic(); err != nil {
		return err
	}

	initPrometheus()

	if err := initDownloading(); err != nil {
		return err
	}

	initErrorsReporting()

	if err := initVips(); err != nil {
		return err
	}

	if err := checkPresets(conf.Presets); err != nil {
		shutdownVips()
		return err
	}

	return nil
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")
  
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
  
	return os.Getenv(key)
}

func run() error {
	goDotEnvVariable("APP_START")
	if err := initialize(); err != nil {
		return err
	}

	defer shutdownVips()
	defer closeErrorsReporting()

	go func() {
		var logMemStats = len(os.Getenv("IMGPROXY_LOG_MEM_STATS")) > 0

		for range time.Tick(time.Duration(conf.FreeMemoryInterval) * time.Second) {
			freeMemory()

			if logMemStats {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				logDebug("MEMORY USAGE: Sys=%d HeapIdle=%d HeapInuse=%d", m.Sys/1024/1024, m.HeapIdle/1024/1024, m.HeapInuse/1024/1024)
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	if prometheusEnabled {
		if err := startPrometheusServer(cancel); err != nil {
			return err
		}
	}

	s, err := startServer(cancel)
	if err != nil {
		return err
	}
	defer shutdownServer(s)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-stop:
	}

	return nil
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "health":
			os.Exit(healthcheck())
		case "version":
			fmt.Println(version)
			os.Exit(0)
		}
	}

	if err := run(); err != nil {
		logFatal(err.Error())
	}
}
