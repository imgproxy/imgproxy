package imgproxy

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/handlers"
	processinghandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/transport"
)

const (
	faviconPath = "/favicon.ico"
	healthPath  = "/health"
)

// Imgproxy holds all the components needed for imgproxy to function
type Imgproxy struct {
	HeaderWriter      *headerwriter.Writer
	Semaphores        *semaphores.Semaphores
	FallbackImage     auximageprovider.Provider
	WatermarkImage    auximageprovider.Provider
	Fetcher           *fetcher.Fetcher
	ProcessingHandler *processinghandler.Handler
	StreamHandler     *stream.Handler
	ImageDataFactory  *imagedata.Factory
	Config            *Config
}

// New creates a new imgproxy instance
func New(ctx context.Context, config *Config) (*Imgproxy, error) {
	headerWriter, err := headerwriter.New(&config.HeaderWriter)
	if err != nil {
		return nil, err
	}

	ts, err := transport.New(&config.Transport)
	if err != nil {
		return nil, err
	}

	fetcher, err := fetcher.New(ts, &config.Fetcher)
	if err != nil {
		return nil, err
	}

	idf := imagedata.NewFactory(fetcher)

	fallbackImage, err := auximageprovider.NewStaticProvider(ctx, &config.FallbackImage, "fallback", idf)
	if err != nil {
		return nil, err
	}

	watermarkImage, err := auximageprovider.NewStaticProvider(ctx, &config.WatermarkImage, "watermark", idf)
	if err != nil {
		return nil, err
	}

	semaphores, err := semaphores.New(&config.Semaphores)
	if err != nil {
		return nil, err
	}

	streamHandler, err := stream.New(&config.StreamHandler, headerWriter, fetcher)
	if err != nil {
		return nil, err
	}

	ph, err := processinghandler.New(
		streamHandler, headerWriter, semaphores, fallbackImage, watermarkImage, idf, &config.ProcessingHandler,
	)
	if err != nil {
		return nil, err
	}

	return &Imgproxy{
		HeaderWriter:      headerWriter,
		Semaphores:        semaphores,
		FallbackImage:     fallbackImage,
		WatermarkImage:    watermarkImage,
		Fetcher:           fetcher,
		StreamHandler:     streamHandler,
		ProcessingHandler: ph,
		ImageDataFactory:  idf,
		Config:            config,
	}, nil
}

// BuildRouter sets up the HTTP routes and middleware
func (i *Imgproxy) BuildRouter() (*server.Router, error) {
	r, err := server.NewRouter(&i.Config.Server)
	if err != nil {
		return nil, err
	}

	r.GET("/", handlers.LandingHandler)
	r.GET("", handlers.LandingHandler)

	r.GET(faviconPath, r.NotFoundHandler).Silent()
	r.GET(healthPath, handlers.HealthHandler).Silent()
	if i.Config.Server.HealthCheckPath != "" {
		r.GET(i.Config.Server.HealthCheckPath, handlers.HealthHandler).Silent()
	}

	r.GET(
		"/*", i.ProcessingHandler.Execute,
		r.WithSecret, r.WithCORS, r.WithPanic, r.WithReportError, r.WithMonitoring,
	)

	r.HEAD("/*", r.OkHandler, r.WithCORS)
	r.OPTIONS("/*", r.OkHandler, r.WithCORS)

	return r, nil
}

// Start runs the imgproxy server. This function blocks until the context is cancelled.
func (i *Imgproxy) StartServer(ctx context.Context) error {
	go i.startMemoryTicker(ctx)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	if err := prometheus.StartServer(cancel); err != nil {
		return err
	}

	router, err := i.BuildRouter()
	if err != nil {
		return err
	}

	s, err := server.Start(cancel, router)
	if err != nil {
		return err
	}
	defer s.Shutdown(context.Background())

	<-ctx.Done()

	return nil
}

// startMemoryTicker starts a ticker that periodically frees memory and optionally logs memory stats
func (i *Imgproxy) startMemoryTicker(ctx context.Context) {
	ticker := time.NewTicker(i.Config.Server.FreeMemoryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			memory.Free()

			if i.Config.Server.LogMemStats {
				memory.LogStats()
			}
		}
	}
}
