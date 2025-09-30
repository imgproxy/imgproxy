package bugsnag

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bugsnag/bugsnag-go/v2"
)

// logger is the logger forwarder for bugsnag
type logger struct{}

type reporter struct{}

func New(config *Config) (*reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(config.Key) == 0 {
		return nil, nil
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       config.Key,
		ReleaseStage: config.Stage,
		PanicHandler: func() {}, // Disable forking the process
		Logger:       logger{},
	})

	return &reporter{}, nil
}

func (l logger) Printf(format string, v ...interface{}) {
	slog.Debug(
		fmt.Sprintf(format, v...),
		"source", "bugsnag",
	)
}

func (r *reporter) Report(err error, req *http.Request, meta map[string]any) {
	extra := make(bugsnag.MetaData)
	for k, v := range meta {
		extra.Add("Processing Context", k, v)
	}

	bugsnag.Notify(err, req, extra)
}

func (r *reporter) Close() {
	// noop
}
