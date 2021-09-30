package security

import (
	"github.com/imgproxy/imgproxy/v3/config"
)

func VerifySourceURL(imageURL string) bool {
	if len(config.AllowedSources) == 0 {
		return true
	}
	for _, allowedSource := range config.AllowedSources {
		if allowedSource.MatchString(imageURL) {
			return true
		}
	}
	return false
}
