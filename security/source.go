package security

import (
	"github.com/imgproxy/imgproxy/v3/config"
)

func VerifySourceURL(imageURL string) error {
	if len(config.AllowedSources) == 0 {
		return nil
	}

	for _, allowedSource := range config.AllowedSources {
		if allowedSource.MatchString(imageURL) {
			return nil
		}
	}

	return newSourceURLError(imageURL)
}
