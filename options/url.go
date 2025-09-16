package options

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

const urlTokenPlain = "plain"

func (f *Factory) preprocessURL(u string) string {
	for _, repl := range f.config.URLReplacements {
		u = repl.Regexp.ReplaceAllString(u, repl.Replacement)
	}

	if len(f.config.BaseURL) == 0 || strings.HasPrefix(u, f.config.BaseURL) {
		return u
	}

	return fmt.Sprintf("%s%s", f.config.BaseURL, u)
}

func (f *Factory) decodeBase64URL(parts []string) (string, string, error) {
	var format string

	if len(parts) > 1 && f.config.Base64URLIncludesFilename {
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

	return f.preprocessURL(string(imageURL)), format, nil
}

func (f *Factory) decodePlainURL(parts []string) (string, string, error) {
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

	return f.preprocessURL(unescaped), format, nil
}

func (f *Factory) DecodeURL(parts []string) (string, string, error) {
	if len(parts) == 0 {
		return "", "", newInvalidURLError("Image URL is empty")
	}

	if parts[0] == urlTokenPlain && len(parts) > 1 {
		return f.decodePlainURL(parts[1:])
	}

	return f.decodeBase64URL(parts)
}
