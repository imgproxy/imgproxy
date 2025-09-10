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
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
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
	HeaderWriter     *headerwriter.Writer
	Semaphores       *semaphores.Semaphores
	FallbackImage    auximageprovider.Provider
	WatermarkImage   auximageprovider.Provider
	Fetcher          *fetcher.Fetcher
	ImageDataFactory *imagedata.Factory
	Handlers         ImgproxyHandlers
	Config           *Config
}

// New creates a new imgproxy instance
func New(ctx context.Context, config *Config) (*Imgproxy, error) {
	headerWriter, err := headerwriter.New(&config.HeaderWriter)
	if err != nil {
		return nil, err
	}

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

	semaphores, err := semaphores.New(&config.Semaphores)
	if err != nil {
		return nil, err
	}

	imgproxy := &Imgproxy{
		HeaderWriter:     headerWriter,
		Semaphores:       semaphores,
		FallbackImage:    fallbackImage,
		WatermarkImage:   watermarkImage,
		Fetcher:          fetcher,
		ImageDataFactory: idf,
		Config:           config,
	}

	imgproxy.Handlers.Health = healthhandler.New()
	imgproxy.Handlers.Landing = landinghandler.New()

	imgproxy.Handlers.Stream, err = streamhandler.New(&config.Handlers.Stream, headerWriter, fetcher)
	if err != nil {
		return nil, err
	}

	imgproxy.Handlers.Processing, err = processinghandler.New(
		imgproxy.Handlers.Stream, headerWriter, semaphores, fallbackImage, watermarkImage, idf, &config.Handlers.Processing,
	)
	if err != nil {
		return nil, err
	}

	return imgproxy, nil
}

// BuildRouter sets up the HTTP routes and middleware
func (i *Imgproxy) BuildRouter() (*server.Router, error) {
	r, err := server.NewRouter(&i.Config.Server)
	if err != nil {
		return nil, err
	}

	r.GET("/", i.Handlers.Landing.Execute)
	r.GET("", i.Handlers.Landing.Execute)

	r.GET(faviconPath, r.NotFoundHandler).Silent()
	r.GET(healthPath, i.Handlers.Health.Execute).Silent()
	if i.Config.Server.HealthCheckPath != "" {
		r.GET(i.Config.Server.HealthCheckPath, i.Handlers.Health.Execute).Silent()
	}

	r.GET(
		"/*", i.Handlers.Processing.Execute,
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
