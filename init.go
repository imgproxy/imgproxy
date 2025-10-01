// init_once.go contains global initialization/teardown functions that should be called exactly once
// per process.
package imgproxy

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// Init performs the global resources initialization. This should be done once per process.
func Init() error {
	if err := env.Load(context.TODO()); err != nil {
		return err
	}

	logCfg, logErr := logger.LoadConfigFromEnv(nil)
	if logErr != nil {
		return logErr
	}

	// Initialize logger as early as possible to log further initialization steps
	if err := logger.Init(logCfg); err != nil {
		return err
	}

	// NOTE: This is temporary workaround. We have to load env vars in config.go before
	// actually configuring ImgProxy instance because for now we use it as a source of truth.
	// Will be removed once we move env var loading to imgproxy.go
	if err := config.Configure(); err != nil {
		// we moved validations to specific config files, hence, no need to return err
		slog.Warn("old config validation warning", "err", err)
	}
	// NOTE: End of temporary workaround.

	maxprocs.Set(maxprocs.Logger(func(msg string, args ...any) {
		slog.Debug(fmt.Sprintf(msg, args...))
	}))

	vipsCfg, err := vips.LoadConfigFromEnv(nil)
	if err != nil {
		return err
	}
	if vipsErr := vips.Init(vipsCfg); vipsErr != nil {
		return vipsErr
	}

	return nil
}

// Shutdown performs global cleanup
func Shutdown() {
	vips.Shutdown()
}
