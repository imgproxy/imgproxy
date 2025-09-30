package generichttp

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT         = env.Describe("IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT", "seconds => 0")
	IMGPROXY_IGNORE_SSL_VERIFICATION           = env.Describe("IMGPROXY_IGNORE_SSL_VERIFICATION", "boolean")
	IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES   = env.Describe("IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES", "boolean")
	IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES = env.Describe("IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES", "boolean")
	IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES    = env.Describe("IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES", "boolean")
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
		env.Duration(&c.ClientKeepAliveTimeout, IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT),
		env.Bool(&c.IgnoreSslVerification, IMGPROXY_IGNORE_SSL_VERIFICATION),
		env.Bool(&c.AllowLinkLocalSourceAddresses, IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES),
		env.Bool(&c.AllowLoopbackSourceAddresses, IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES),
		env.Bool(&c.AllowPrivateSourceAddresses, IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES),
	)

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.ClientKeepAliveTimeout < 0 {
		return IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT.ErrorZeroOrNegative()
	}

	if c.IgnoreSslVerification {
		IMGPROXY_IGNORE_SSL_VERIFICATION.Warn("ignoring SSL verification is very unsafe") // ⎈
	}

	return nil
}
