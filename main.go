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
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/prometheus"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func buildRouter(r *server.Router) *server.Router {
	r.GET("/", handlers.HandleLanding, true)
	r.GET("", handlers.HandleLanding, true)

	r.GET("/", r.WithMetrics(
		r.WithCORS(
			r.WithReportError(
				r.WithPanic(
					r.WithSecret(handleProcessing),
				),
			),
		),
	), false)

	r.HEAD("/", r.WithCORS(handlers.HeadHandler), false)
	r.OPTIONS("/", r.WithCORS(handlers.HeadHandler), false)

	r.HealthHandler = handlers.HealthHandler

	return r
}

func initialize() error {
	if err := loadenv.Load(); err != nil {
		return err
	}

	if err := logger.Init(); err != nil {
		return err
	}

	gliblog.Init()

	maxprocs.Set(maxprocs.Logger(log.Debugf))

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

func run(ctx context.Context) error {
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

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	if err := prometheus.StartServer(cancel); err != nil {
		return err
	}

	cfg := server.NewConfigFromEnv()
	r := server.NewRouter(cfg)
	s, err := server.Start(cancel, buildRouter(r), cfg)
	if err != nil {
		return err
	}
	defer s.Shutdown(ctx)

	<-ctx.Done()

	return nil
}

func main() {
	flag.Parse()

	switch flag.Arg(0) {
	case "health":
		os.Exit(healthcheck())
	case "version":
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
