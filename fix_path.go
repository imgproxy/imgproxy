package main

import (
	"fmt"
	"regexp"
	"strings"
)

var fixPathRe = regexp.MustCompile(`^(\S+)\:/([^/])`)

func fixPath(path string) string {
	// Cut the path at the `/plain/` segment to process those parts separately
	options, plainURL, hasPlain := strings.Cut(path, "/plain/")

	// Some proxies/CDNs may encode `:` in options as `%3A`, so we need to unescape it first
	path = strings.ReplaceAll(options, "%3A", ":")

	if hasPlain {
		// Some proxies/CDNs may "normalize" URLs by replacing `scheme://` with `scheme:/`
		// in the plain URL part, so we need to fix it back.
		if match := fixPathRe.FindStringSubmatch(plainURL); match != nil {
			repl := fmt.Sprintf("%s://", match[1])
			if match[1] == "local" {
				repl += "/"
			}
			repl += match[2]
			plainURL = strings.Replace(plainURL, match[0], repl, 1)
		}

		return path + "/plain/" + plainURL
	}

	return path
}
