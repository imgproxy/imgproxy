// init_once.go contains global initialization/teardown functions that should be called exactly once
// per process.
package imgproxy

import (
	"sync"
	"sync/atomic"

	"github.com/DataDog/datadog-agent/pkg/trace/log"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/loadenv"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/gliblog"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/vips"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	done atomic.Bool // done indicates that initialization has been performed
	once sync.Once   // once is used to ensure initialization is performed only once
)

// Init performs the global resources initialization. This should be done once per process.
func Init() error {
	var err error
	once.Do(func() {
		err = initialize()
		done.Store(true)
	})
	return err
}

// Shutdown performs global cleanup
func Shutdown() {
	if !done.Load() {
		return
	}

	vips.Shutdown()
	monitoring.Stop()
	errorreport.Close()
}

// initialize contains the actual initialization logic
func initialize() error {
	if err := logger.Init(); err != nil {
		return err
	}

	gliblog.Init()

	maxprocs.Set(maxprocs.Logger(log.Debugf))

	if err := monitoring.Init(); err != nil {
		return err
	}

	if err := vips.Init(); err != nil {
		return err
	}

	errorreport.Init()

	// NOTE: This is temporary workaround. We have to load env vars in config.go before
	// actually configuring ImgProxy instance because for now we use it as a source of truth.
	// Will be removed once we move env var loading to imgproxy.go
	if err := loadenv.Load(); err != nil {
		return err
	}

	if err := config.Configure(); err != nil {
		return err
	}
	// NOTE: End of temporary workaround.

	return nil
}
