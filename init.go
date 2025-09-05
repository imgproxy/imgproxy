// init_once.go contains global initialization/teardown functions that should be called exactly once
// per process.
package imgproxy

import (
	"sync"
	"sync/atomic"

	"github.com/DataDog/datadog-agent/pkg/trace/log"
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

// Init performs the global resources initialization.
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

	return nil
}
