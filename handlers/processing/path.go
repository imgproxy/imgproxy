package processing

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

// fixPathRe is used in path re-denormalization
var fixPathRe = regexp.MustCompile(`/plain/(\S+)\:/([^/])`)

// splitPathSignature splits signature and path components from the request URI
func splitPathSignature(r *http.Request, config *Config) (string, string, error) {
	uri := r.RequestURI

	// cut query params
	uri, _, _ = strings.Cut(uri, "?")

	// cut path prefix
	if len(config.PathPrefix) > 0 {
		uri = strings.TrimPrefix(uri, config.PathPrefix)
	}

	// cut leading slash
	uri = strings.TrimPrefix(uri, "/")

	signature, path, _ := strings.Cut(uri, "/")
	if len(signature) == 0 || len(path) == 0 {
		return "", "", ierrors.Wrap(
			newInvalidURLErrorf(http.StatusNotFound, "Invalid path: %s", path), 0,
			ierrors.WithCategory(categoryPathParsing),
		)
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
