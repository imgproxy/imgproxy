package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// fixPathRe is used in path re-denormalization
var fixPathRe = regexp.MustCompile(`^(\S+)\:/([^/])`)

// SplitPathSignature splits signature and path components from the request URI
func SplitPathSignature(r *http.Request) (string, string, error) {
	uri := r.RequestURI

	// cut query params
	uri, _, _ = strings.Cut(uri, "?")

	// Cut path prefix.
	// r.Pattern is set by the router and contains both global and route-specific prefixes combined.
	if len(r.Pattern) > 0 {
		uri = strings.TrimPrefix(uri, r.Pattern)
	}

	// cut leading slash
	uri = strings.TrimPrefix(uri, "/")

	signature, path, _ := strings.Cut(uri, "/")
	if len(signature) == 0 || len(path) == 0 {
		return "", "", NewInvalidPathError(r.Context(), path)
	}

	// restore broken slashes in the path
	path = redenormalizePath(path)

	return path, signature, nil
}

// redenormalizePath undoes path normalization done by some browsers and revers proxies
func redenormalizePath(path string) string {
	// Cut the path at the `/plain/` segment to process those parts separately
	options, plainURL, hasPlain := strings.Cut(path, "/plain/")

	// Some proxies/CDNs may encode `:` in options as `%3A`, so we need to unescape it first
	path = strings.ReplaceAll(options, "%3A", ":")

	if !hasPlain {
		return path
	}

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
