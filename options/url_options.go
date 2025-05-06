package options

import (
	"net/url"
	"strconv"
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

func parseURLOptionsIPC(qs url.Values, path string) (urlOptions, string, error) {
	dimensions, path, err := parseDimensions(path)

	if err != nil {
		return nil, "", err
	}

	// Initialize parsed options with "rs"
	parsed := urlOptions{
		{Name: "rs", Args: []string{"fill-down", dimensions[0], dimensions[1]}},
	}

	// Define allowed query parameters
	validKeys := map[string]bool{"qp": true, "wm": true, "art": true, "fmt" : true, "fit" : true}

	if isMediaPath(path) {
		parsed = append(parsed, urlOption{
			Name: "msr",
			Args: []string{strconv.Itoa(config.MaxMediaSrcResolution)},
		})
	}

	// Append valid query parameters
	for key, val := range qs {
		if validKeys[key] {
			if key == "fit" {
				parsed[0].Args[0] = "fit"
				continue
			}
			parsed = append(parsed, urlOption{Name: key, Args: val})
		}
	}

	return parsed, path, nil
}

// Helper to detect media-related paths
func isMediaPath(path string) bool {
	for _, prefix := range config.MediaPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func parseDimensions(path string) ([]string, string, error) {
	if path == "" {
		return nil, "", newInvalidURLError("path is empty")
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, "", newInvalidURLError("path does not contain '/' separator")
	}

	dimParts := strings.Split(parts[0], "x")
	if len(dimParts) != 2 {
		return nil, "", newInvalidURLError("dimension part does not contain 2 segments separated by 'x'")
	}

	dimensions := []string{dimParts[0], dimParts[1]}
	return dimensions, parts[1], nil
}