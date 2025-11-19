package bugsnag

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/imgproxy/imgproxy/v3/errctx"
)

// logger is the logger forwarder for bugsnag
type logger struct{}

func (l logger) Printf(format string, v ...interface{}) {
	slog.Debug(fmt.Sprintf(format, v...), "source", "bugsnag")
}

type reporter struct {
	notifier *bugsnag.Notifier
}

func New(config *Config) (*reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(config.Key) == 0 {
		return nil, nil
	}

	notifier := bugsnag.New(bugsnag.Configuration{
		APIKey:       config.Key,
		ReleaseStage: config.Stage,
		PanicHandler: func() {}, // Disable forking the process
		Logger:       logger{},
		Synchronous:  true,
	})

	return &reporter{notifier: notifier}, nil
}

func (r *reporter) Report(err errctx.Error, req *http.Request, meta map[string]any) {
	extra := make(bugsnag.MetaData)
	for k, v := range meta {
		extra.Add("Processing Context", k, v)
	}

	// imgproxy may wrap errors using errctx.WrappedError to add context, so Bugsnag
	// would report the error type as *errctx.WrappedError.
	//
	// To avoid this, we provide error class information explicitly.
	errClass := bugsnag.ErrorClass{Name: errctx.ErrorType(err)}

	if repErr := r.notifier.Notify(err, errClass, req, extra); repErr != nil {
		slog.Warn("Failed to report error to Bugsnag", "error", repErr)
	}
}

func (r *reporter) Close() {
	// noop
}
