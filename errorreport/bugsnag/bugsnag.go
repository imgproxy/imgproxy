package bugsnag

import (
	"net/http"

	"github.com/bugsnag/bugsnag-go"
	"github.com/imgproxy/imgproxy/v2/config"
)

var enabled bool

func Init() {
	if len(config.BugsnagKey) > 0 {
		bugsnag.Configure(bugsnag.Configuration{
			APIKey:       config.BugsnagKey,
			ReleaseStage: config.BugsnagStage,
		})
		enabled = true
	}
}

func Report(err error, req *http.Request) {
	if enabled {
		bugsnag.Notify(err, req)
	}
}
