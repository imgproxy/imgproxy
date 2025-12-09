package imgproxy

import (
	"context"
	"net"
	"time"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/clientfeatures"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	healthhandler "github.com/imgproxy/imgproxy/v3/handlers/health"
	landinghandler "github.com/imgproxy/imgproxy/v3/handlers/landing"
	processinghandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	streamhandler "github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/memory"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	optionsparser "github.com/imgproxy/imgproxy/v3/options/parser"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/workers"
)

const (
	faviconPath   = "/favicon.ico"
	healthPath    = "/health"
	wellKnownPath = "/.well-known/*"
)

// ImgproxyHandlers holds the handlers for imgproxy.
type ImgproxyHandlers struct {
	Health     *healthhandler.Handler
	Landing    *landinghandler.Handler
	Processing *processinghandler.Handler
	Stream     *streamhandler.Handler
}

// Imgproxy holds all the components needed for imgproxy to function.
type Imgproxy struct {
	workers                *workers.Workers
	fallbackImage          auximageprovider.Provider
	watermarkImage         auximageprovider.Provider
	fetcher                *fetcher.Fetcher
	imageDataFactory       *imagedata.Factory
	clientFeaturesDetector *clientfeatures.Detector
	handlers               ImgproxyHandlers
	securityChecker        *security.Checker
	optionsParser          *optionsparser.Parser
	processor              *processing.Processor
	cookies                *cookies.Cookies
	monitoring             *monitoring.Monitoring
	config                 *Config
	errorReporter          *errorreport.Reporter
}

// New creates a new imgproxy instance.
func New(ctx context.Context, config *Config) (*Imgproxy, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	monitoring, err := monitoring.New(ctx, &config.Monitoring, config.Workers.WorkersNumber)
	if err != nil {
		return nil, err
	}

	errorReporter, err := errorreport.New(&config.ErrorReport)
	if err != nil {
		return nil, err
	}

	securityChecker, err := security.New(&config.Security)
	if err != nil {
		return nil, err
	}

	fetcher, err := fetcher.New(&config.Fetcher)
	if err != nil {
		return nil, err
	}

	idf := imagedata.NewFactory(fetcher, monitoring)

	clientFeaturesDetector := clientfeatures.NewDetector(&config.ClientFeatures)

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

	processor, err := processing.New(&config.Processing, securityChecker, watermarkImage)
	if err != nil {
		return nil, err
	}

	optionsParser, err := optionsparser.New(&config.OptionsParser)
	if err != nil {
		return nil, err
	}

	cookies, err := cookies.New(&config.Cookies)
	if err != nil {
		return nil, err
	}

	imgproxy := &Imgproxy{
		workers:                workers,
		fallbackImage:          fallbackImage,
		watermarkImage:         watermarkImage,
		fetcher:                fetcher,
		imageDataFactory:       idf,
		clientFeaturesDetector: clientFeaturesDetector,
		config:                 config,
		securityChecker:        securityChecker,
		optionsParser:          optionsParser,
		processor:              processor,
		cookies:                cookies,
		monitoring:             monitoring,
		errorReporter:          errorReporter,
	}

	imgproxy.handlers.Health = healthhandler.New()
	imgproxy.handlers.Landing = landinghandler.New()

	imgproxy.handlers.Stream, err = streamhandler.New(imgproxy, &config.Handlers.Stream)
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
	r, err := server.NewRouter(&i.config.Server, i.monitoring, i.errorReporter)
	if err != nil {
		return nil, err
	}

	r.GET("/", i.handlers.Landing.Execute)
	r.GET("", i.handlers.Landing.Execute)

	r.GET(faviconPath, r.NotFoundHandler).Silent()
	r.GET(healthPath, i.handlers.Health.Execute).Silent()
	r.GET(wellKnownPath, r.NotFoundHandler).Silent()

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

// StartServer runs the imgproxy server. This function blocks until the context is cancelled.
// If hasStarted is not nil, it will be notified with the server address once
// the server is ready or about to be ready to accept requests.
func (i *Imgproxy) StartServer(ctx context.Context, hasStarted chan net.Addr) error {
	go i.startMemoryTicker(ctx)

	ctx, cancel := context.WithCancel(ctx)

	if err := i.monitoring.StartPrometheus(cancel); err != nil {
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

// Close gracefully shuts down the imgproxy instance
func (i *Imgproxy) Close(ctx context.Context) {
	i.monitoring.Stop(ctx)
	i.errorReporter.Close()
}

func (i *Imgproxy) Fetcher() *fetcher.Fetcher {
	return i.fetcher
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

func (i *Imgproxy) ClientFeaturesDetector() *clientfeatures.Detector {
	return i.clientFeaturesDetector
}

func (i *Imgproxy) Security() *security.Checker {
	return i.securityChecker
}

func (i *Imgproxy) OptionsParser() *optionsparser.Parser {
	return i.optionsParser
}

func (i *Imgproxy) Processor() *processing.Processor {
	return i.processor
}

func (i *Imgproxy) Cookies() *cookies.Cookies {
	return i.cookies
}

func (i *Imgproxy) Monitoring() *monitoring.Monitoring {
	return i.monitoring
}

func (i *Imgproxy) ErrorReporter() *errorreport.Reporter {
	return i.errorReporter
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
