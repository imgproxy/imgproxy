// init_once.go contains global initialization/teardown functions that should be called exactly once
// per process.
package imgproxy

import (
	"fmt"
	"log/slog"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/loadenv"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// Init performs the global resources initialization. This should be done once per process.
func Init() error {
	if err := loadenv.Load(); err != nil {
		return err
	}

	logCfg := logger.LoadConfigFromEnv(nil)
	if err := logger.Init(logCfg); err != nil {
		return err
	}

	// NOTE: This is temporary workaround. We have to load env vars in config.go before
	// actually configuring ImgProxy instance because for now we use it as a source of truth.
	// Will be removed once we move env var loading to imgproxy.go
	if err := config.Configure(); err != nil {
		return err
	}
	// NOTE: End of temporary workaround.

	maxprocs.Set(maxprocs.Logger(func(msg string, args ...any) {
		slog.Debug(fmt.Sprintf(msg, args...))
	}))

	if err := monitoring.Init(); err != nil {
		return err
	}

	if err := vips.Init(); err != nil {
		return err
	}

	errorreport.Init()

	return nil
}

// Shutdown performs global cleanup
func Shutdown() {
	monitoring.Stop()
	errorreport.Close()
	vips.Shutdown()
}
