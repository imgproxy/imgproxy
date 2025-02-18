package options

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
)

const urlTokenPlain = "plain"

func preprocessURL(u string) string {
	for _, repl := range config.URLReplacements {
		u = repl.Regexp.ReplaceAllString(u, repl.Replacement)
	}

	if len(config.BaseURL) == 0 || strings.HasPrefix(u, config.BaseURL) {
		return u
	}

	return fmt.Sprintf("%s%s", config.BaseURL, u)
}

func decodeBase64URL(parts []string) (string, string, error) {
	var format string

	if len(parts) > 1 && config.Base64URLIncludesFilename {
		parts = parts[:len(parts)-1]
	}

	encoded := strings.Join(parts, "")
	urlParts := strings.Split(encoded, ".")

	if len(urlParts[0]) == 0 {
		return "", "", newInvalidURLError("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", newInvalidURLError("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	imageURL, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(urlParts[0], "="))
	if err != nil {
		return "", "", newInvalidURLError("Invalid url encoding: %s", encoded)
	}

	return preprocessURL(string(imageURL)), format, nil
}

func decodePlainURL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "/")
	urlParts := strings.Split(encoded, "@")

	if len(urlParts[0]) == 0 {
		return "", "", newInvalidURLError("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", newInvalidURLError("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	unescaped, err := url.PathUnescape(urlParts[0])
	if err != nil {
		return "", "", newInvalidURLError("Invalid url encoding: %s", encoded)
	}

	return preprocessURL(unescaped), format, nil
}

func DecodeURL(parts []string) (string, string, error) {
	if len(parts) == 0 {
		return "", "", newInvalidURLError("Image URL is empty")
	}

	if parts[0] == urlTokenPlain && len(parts) > 1 {
		return decodePlainURL(parts[1:])
	}

	return decodeBase64URL(parts)
}
