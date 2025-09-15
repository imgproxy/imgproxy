package imgproxy

import (
	"context"
	"net"
	"time"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	healthhandler "github.com/imgproxy/imgproxy/v3/handlers/health"
	landinghandler "github.com/imgproxy/imgproxy/v3/handlers/landing"
	processinghandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	streamhandler "github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/workers"
)

const (
	faviconPath = "/favicon.ico"
	healthPath  = "/health"
)

// ImgproxyHandlers holds the handlers for imgproxy
type ImgproxyHandlers struct {
	Health     *healthhandler.Handler
	Landing    *landinghandler.Handler
	Processing *processinghandler.Handler
	Stream     *streamhandler.Handler
}

// Imgproxy holds all the components needed for imgproxy to function
type Imgproxy struct {
	workers                  *workers.Workers
	fallbackImage            auximageprovider.Provider
	watermarkImage           auximageprovider.Provider
	fetcher                  *fetcher.Fetcher
	imageDataFactory         *imagedata.Factory
	handlers                 ImgproxyHandlers
	security                 *security.Checker
	processingOptionsFactory *options.Factory
	config                   *Config
}

// New creates a new imgproxy instance
func New(ctx context.Context, config *Config) (*Imgproxy, error) {
	fetcher, err := fetcher.New(&config.Fetcher)
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

	workers, err := workers.New(&config.Workers)
	if err != nil {
		return nil, err
	}

	security, err := security.New(&config.Security)
	if err != nil {
		return nil, err
	}

	processingOptionsFactory, err := options.NewFactory(&config.ProcessingOptions, security)
	if err != nil {
		return nil, err
	}

	imgproxy := &Imgproxy{
		workers:                  workers,
		fallbackImage:            fallbackImage,
		watermarkImage:           watermarkImage,
		fetcher:                  fetcher,
		imageDataFactory:         idf,
		config:                   config,
		security:                 security,
		processingOptionsFactory: processingOptionsFactory,
	}

	imgproxy.handlers.Health = healthhandler.New()
	imgproxy.handlers.Landing = landinghandler.New()

	imgproxy.handlers.Stream, err = streamhandler.New(&config.Handlers.Stream, fetcher)
	if err != nil {
		return nil, err
	}

	imgproxy.handlers.Processing, err = processinghandler.New(
		imgproxy, imgproxy.handlers.Stream, &config.Handlers.Processing,
	)
	if err != nil {
		return nil, err
	}

	return imgproxy, nil
}

// BuildRouter sets up the HTTP routes and middleware
func (i *Imgproxy) BuildRouter() (*server.Router, error) {
	r, err := server.NewRouter(&i.config.Server)
	if err != nil {
		return nil, err
	}

	r.GET("/", i.handlers.Landing.Execute)
	r.GET("", i.handlers.Landing.Execute)

	r.GET(faviconPath, r.NotFoundHandler).Silent()
	r.GET(healthPath, i.handlers.Health.Execute).Silent()
	if i.config.Server.HealthCheckPath != "" {
		r.GET(i.config.Server.HealthCheckPath, i.handlers.Health.Execute).Silent()
	}

	r.GET(
		"/*", i.handlers.Processing.Execute,
		r.WithSecret, r.WithCORS, r.WithPanic, r.WithReportError, r.WithMonitoring,
	)

	r.HEAD("/*", r.OkHandler, r.WithCORS)
	r.OPTIONS("/*", r.OkHandler, r.WithCORS)

	return r, nil
}

// Start runs the imgproxy server. This function blocks until the context is cancelled.
// If hasStarted is not nil, it will be notified with the server address once
// the server is ready or about to be ready to accept requests.
func (i *Imgproxy) StartServer(ctx context.Context, hasStarted chan net.Addr) error {
	go i.startMemoryTicker(ctx)

	ctx, cancel := context.WithCancel(ctx)

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

	if hasStarted != nil {
		hasStarted <- s.Addr
		close(hasStarted)
	}

	<-ctx.Done()

	return nil
}

// startMemoryTicker starts a ticker that periodically frees memory and optionally logs memory stats
func (i *Imgproxy) startMemoryTicker(ctx context.Context) {
	ticker := time.NewTicker(i.config.Server.FreeMemoryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			memory.Free()

			if i.config.Server.LogMemStats {
				memory.LogStats()
			}
		}
	}
}

func (i *Imgproxy) Workers() *workers.Workers {
	return i.workers
}

func (i *Imgproxy) FallbackImage() auximageprovider.Provider {
	return i.fallbackImage
}

func (i *Imgproxy) WatermarkImage() auximageprovider.Provider {
	return i.watermarkImage
}

func (i *Imgproxy) ImageDataFactory() *imagedata.Factory {
	return i.imageDataFactory
}

func (i *Imgproxy) Security() *security.Checker {
	return i.security
}

func (i *Imgproxy) ProcessingOptionsFactory() *options.Factory {
	return i.processingOptionsFactory
}
