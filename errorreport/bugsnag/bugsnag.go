package bugsnag

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bugsnag/bugsnag-go/v2"

	"github.com/imgproxy/imgproxy/v3/env"
)

// logger is the logger forwarder for bugsnag
type logger struct{}

var (
	IMGPROXY_BUGSNAG_KEY   = env.Describe("IMGPROXY_BUGSNAG_KEY", "string")
	IMGPROXY_BUGSNAG_STAGE = env.Describe("IMGPROXY_BUGSNAG_STAGE", "string")
)

type reporter struct{}

func New() (*reporter, error) {
	key := ""
	stage := "production"

	err := errors.Join(
		env.String(&key, IMGPROXY_BUGSNAG_KEY),
		env.String(&stage, IMGPROXY_BUGSNAG_STAGE),
	)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, nil
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       key,
		ReleaseStage: stage,
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
