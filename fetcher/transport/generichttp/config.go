package generichttp

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT         = env.Duration("IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT")
	IMGPROXY_IGNORE_SSL_VERIFICATION           = env.Bool("IMGPROXY_IGNORE_SSL_VERIFICATION")
	IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES   = env.Bool("IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES")
	IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES = env.Bool("IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES")
	IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES    = env.Bool("IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES")
)

// Config holds the configuration for the generic HTTP transport
type Config struct {
	ClientKeepAliveTimeout        time.Duration
	IgnoreSslVerification         bool
	AllowLoopbackSourceAddresses  bool
	AllowLinkLocalSourceAddresses bool
	AllowPrivateSourceAddresses   bool
}

// NewDefaultConfig returns a new default configuration for the generic HTTP transport
func NewDefaultConfig() Config {
	return Config{
		ClientKeepAliveTimeout:        90 * time.Second,
		IgnoreSslVerification:         false,
		AllowLoopbackSourceAddresses:  false,
		AllowLinkLocalSourceAddresses: false,
		AllowPrivateSourceAddresses:   true,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT.Parse(&c.ClientKeepAliveTimeout),
		IMGPROXY_IGNORE_SSL_VERIFICATION.Parse(&c.IgnoreSslVerification),
		IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES.Parse(&c.AllowLinkLocalSourceAddresses),
		IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES.Parse(&c.AllowLoopbackSourceAddresses),
		IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES.Parse(&c.AllowPrivateSourceAddresses),
	)

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.ClientKeepAliveTimeout < 0 {
		return IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT.ErrorZeroOrNegative()
	}

	return nil
}
