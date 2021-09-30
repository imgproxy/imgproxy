package options

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
)

const urlTokenPlain = "plain"

func addBaseURL(u string) string {
	if len(config.BaseURL) == 0 || strings.HasPrefix(u, config.BaseURL) {
		return u
	}

	return fmt.Sprintf("%s%s", config.BaseURL, u)
}

func decodeBase64URL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "")
	urlParts := strings.Split(encoded, ".")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	imageURL, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(urlParts[0], "="))
	if err != nil {
		return "", "", fmt.Errorf("Invalid url encoding: %s", encoded)
	}

	return addBaseURL(string(imageURL)), format, nil
}

func decodePlainURL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "/")
	urlParts := strings.Split(encoded, "@")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	unescaped, err := url.PathUnescape(urlParts[0])
	if err != nil {
		return "", "", fmt.Errorf("Invalid url encoding: %s", encoded)
	}

	return addBaseURL(unescaped), format, nil
}

func DecodeURL(parts []string) (string, string, error) {
	if len(parts) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if parts[0] == urlTokenPlain && len(parts) > 1 {
		return decodePlainURL(parts[1:])
	}

	return decodeBase64URL(parts)
}
