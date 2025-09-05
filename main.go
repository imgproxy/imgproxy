package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/loadenv"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/gliblog"
	"github.com/imgproxy/imgproxy/v3/handlers"
	processingHandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/transport"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/imgproxy/imgproxy/v3/vips"
)

const (
	faviconPath    = "/favicon.ico"
	healthPath     = "/health"
	categoryConfig = "(tmp)config" // NOTE: temporary category for reporting configration errors
)

func callHandleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) error {
	// NOTE: This is temporary, will be moved level up at once
	hwc, err := headerwriter.LoadFromEnv(headerwriter.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	hw, err := headerwriter.New(hwc)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	sc, err := stream.LoadFromEnv(stream.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	tcfg, err := transport.LoadFromEnv(transport.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	tr, err := transport.New(tcfg)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	fc, err := fetcher.LoadFromEnv(fetcher.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	fetcher, err := fetcher.New(tr, fc)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	idf := imagedata.NewFactory(fetcher)

	stream, err := stream.New(sc, hw, fetcher)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	phc, err := processingHandler.LoadFromEnv(processingHandler.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	semc, err := semaphores.LoadFromEnv(semaphores.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	semaphores, err := semaphores.New(semc)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	fic := auximageprovider.NewDefaultStaticConfig()
	fic, err = auximageprovider.LoadFallbackStaticConfigFromEnv(fic)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	fi, err := auximageprovider.NewStaticProvider(
		r.Context(),
		fic,
		"fallback image",
		idf,
	)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	wic := auximageprovider.NewDefaultStaticConfig()
	wic, err = auximageprovider.LoadWatermarkStaticConfigFromEnv(wic)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	wi, err := auximageprovider.NewStaticProvider(
		r.Context(),
		wic,
		"watermark image",
		idf,
	)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	h, err := processingHandler.New(stream, hw, semaphores, fi, wi, idf, phc)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	return h.Execute(reqID, rw, r)
}

func buildRouter(r *server.Router) *server.Router {
	r.GET("/", handlers.LandingHandler)
	r.GET("", handlers.LandingHandler)

	r.GET(faviconPath, r.NotFoundHandler).Silent()
	r.GET(healthPath, handlers.HealthHandler).Silent()
	if config.HealthCheckPath != "" {
		r.GET(config.HealthCheckPath, handlers.HealthHandler).Silent()
	}

	r.GET(
		"/*", callHandleProcessing,
		r.WithSecret, r.WithCORS, r.WithPanic, r.WithReportError, r.WithMonitoring,
	)

	r.HEAD("/*", r.OkHandler, r.WithCORS)
	r.OPTIONS("/*", r.OkHandler, r.WithCORS)

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

	if err := monitoring.Init(); err != nil {
		return err
	}

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
	monitoring.Stop()
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

	cfg, err := server.LoadFromEnv(server.NewDefaultConfig())
	if err != nil {
		return err
	}

	r, err := server.NewRouter(cfg)
	if err != nil {
		return err
	}

	s, err := server.Start(cancel, buildRouter(r))
	if err != nil {
		return err
	}
	defer s.Shutdown(context.Background())

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
