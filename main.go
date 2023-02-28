package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/loadenv"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/gliblog"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/prometheus"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func initialize() error {
	if err := logger.Init(); err != nil {
		return err
	}

	gliblog.Init()

	maxprocs.Set(maxprocs.Logger(log.Debugf))

	if err := loadenv.Load(); err != nil {
		return err
	}

	if err := config.Configure(); err != nil {
		return err
	}

	if err := metrics.Init(); err != nil {
		return err
	}

	if err := imagedata.Init(); err != nil {
		return err
	}

	initProcessingHandler()

	errorreport.Init()

	if err := vips.Init(); err != nil {
		return err
	}

	if err := processing.ValidatePreferredFormats(); err != nil {
		vips.Shutdown()
		return err
	}

	if err := options.ParsePresets(config.Presets); err != nil {
		vips.Shutdown()
		return err
	}

	if err := options.ValidatePresets(); err != nil {
		vips.Shutdown()
		return err
	}

	return nil
}

func shutdown() {
	vips.Shutdown()
	metrics.Stop()
	errorreport.Close()
}

func run() error {
	if err := initialize(); err != nil {
		return err
	}

	defer shutdown()

	go func() {
		var logMemStats = len(os.Getenv("IMGPROXY_LOG_MEM_STATS")) > 0

		for range time.Tick(time.Duration(config.FreeMemoryInterval) * time.Second) {
			memory.Free()

			if logMemStats {
				memory.LogStats()
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	if err := prometheus.StartServer(cancel); err != nil {
		return err
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
	flag.Parse()

	switch flag.Arg(0) {
	case "health":
		os.Exit(healthcheck())
	case "version":
		fmt.Println(version.Version())
		os.Exit(0)
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
