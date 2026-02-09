package imgproxy

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// Init performs the global resources initialization. This should be done once per process.
func Init(ctx context.Context) error {
	if err := env.Load(ctx); err != nil {
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
