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

func parseURLOptionsIPC(size string, qs url.Values) urlOptions {
	// Split size only once
	dimensions := strings.SplitN(size, "x", 2)
	if len(dimensions) != 2 {
		return nil // Return nil or handle error if size is invalid
	}

	// Initialize parsed options with "rs"
	parsed := urlOptions{
		{Name: "rs", Args: []string{"fill-down", dimensions[0], dimensions[1]}},
	}

	// Define allowed query parameters
	validKeys := map[string]bool{"q": true, "wm": true, "art": true}

	// Append valid query parameters
	for key, val := range qs {
		if validKeys[key] {
			parsed = append(parsed, urlOption{Name: key, Args: val})
		}
	}

	return parsed
}


