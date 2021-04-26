package security

import (
	"strings"

	"github.com/imgproxy/imgproxy/v2/config"
)

func VerifySourceURL(imageURL string) bool {
	if len(config.AllowedSources) == 0 {
		return true
	}
	for _, val := range config.AllowedSources {
		if strings.HasPrefix(imageURL, string(val)) {
			return true
		}
	}
	return false
}
