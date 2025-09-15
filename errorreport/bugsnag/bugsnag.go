package bugsnag

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bugsnag/bugsnag-go/v2"

	"github.com/imgproxy/imgproxy/v3/config"
)

var enabled bool

type logger struct{}

func (l logger) Printf(format string, v ...interface{}) {
	slog.Debug(
		fmt.Sprintf(format, v...),
		"source", "bugsnag",
	)
}

func Init() {
	if len(config.BugsnagKey) > 0 {
		bugsnag.Configure(bugsnag.Configuration{
			APIKey:       config.BugsnagKey,
			ReleaseStage: config.BugsnagStage,
			PanicHandler: func() {}, // Disable forking the process
			Logger:       logger{},
		})
		enabled = true
	}
}

func Report(err error, req *http.Request, meta map[string]any) {
	if !enabled {
		return
	}

	extra := make(bugsnag.MetaData)
	for k, v := range meta {
		extra.Add("Processing Context", k, v)
	}

	bugsnag.Notify(err, req, extra)
}
