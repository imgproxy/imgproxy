package options

import (
	"net/url"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
)

type urlOption struct {
	Name string
	Args []string
}

type urlOptions []urlOption

func parseURLOptions(opts []string) (urlOptions, []string) {
	parsed := make(urlOptions, 0, len(opts))
	urlStart := len(opts) + 1

	for i, opt := range opts {
		// URL-decode the option segment to support percent-encoded separators
		if decoded, err := url.PathUnescape(opt); err == nil {
			opt = decoded
		}

		args := strings.Split(opt, config.ArgumentsSeparator)

		if len(args) == 1 {
			urlStart = i
			break
		}

		parsed = append(parsed, urlOption{Name: args[0], Args: args[1:]})
	}

	var rest []string

	if urlStart < len(opts) {
		rest = opts[urlStart:]
	} else {
		rest = []string{}
	}

	return parsed, rest
}
