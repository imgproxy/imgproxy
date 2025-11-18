package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// fixPathRe is used in path re-denormalization
var fixPathRe = regexp.MustCompile(`/plain/(\S+)\:/([^/])`)

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
		return "", "", NewInvalidPathError(path)
	}

	// restore broken slashes in the path
	path = redenormalizePath(path)

	return path, signature, nil
}

// redenormalizePath undoes path normalization done by some browsers and revers proxies
func redenormalizePath(path string) string {
	for _, match := range fixPathRe.FindAllStringSubmatch(path, -1) {
		repl := fmt.Sprintf("/plain/%s://", match[1])
		if match[1] == "local" {
			repl += "/"
		}
		repl += match[2]
		path = strings.Replace(path, match[0], repl, 1)
	}

	return path
}
