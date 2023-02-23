package security

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

var ErrSecurityOptionsNotAllowed = ierrors.New(403, "Security processing options are not allowed", "Invalid URL")

type Options struct {
	MaxSrcResolution            int
	MaxSrcFileSize              int
	MaxAnimationFrames          int
	MaxAnimationFrameResolution int
}

func DefaultOptions() Options {
	return Options{
		MaxSrcResolution:            config.MaxSrcResolution,
		MaxSrcFileSize:              config.MaxSrcFileSize,
		MaxAnimationFrames:          config.MaxAnimationFrames,
		MaxAnimationFrameResolution: config.MaxAnimationFrameResolution,
	}
}

func IsSecurityOptionsAllowed() error {
	if config.AllowSecurityOptions {
		return nil
	}

	return ErrSecurityOptionsNotAllowed
}
