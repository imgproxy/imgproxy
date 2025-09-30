package swift

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_SWIFT_USERNAME                = env.Describe("IMGPROXY_SWIFT_USERNAME", "string")
	IMGPROXY_SWIFT_API_KEY                 = env.Describe("IMGPROXY_SWIFT_API_KEY", "string")
	IMGPROXY_SWIFT_AUTH_URL                = env.Describe("IMGPROXY_SWIFT_AUTH_URL", "string")
	IMGPROXY_SWIFT_DOMAIN                  = env.Describe("IMGPROXY_SWIFT_DOMAIN", "string")
	IMGPROXY_SWIFT_TENANT                  = env.Describe("IMGPROXY_SWIFT_TENANT", "string")
	IMGPROXY_SWIFT_AUTH_VERSION            = env.Describe("IMGPROXY_SWIFT_AUTH_VERSION", "number")
	IMGPROXY_SWIFT_CONNECT_TIMEOUT_SECONDS = env.Describe("IMGPROXY_SWIFT_CONNECT_TIMEOUT_SECONDS", "number")
	IMGPROXY_SWIFT_TIMEOUT_SECONDS         = env.Describe("IMGPROXY_SWIFT_TIMEOUT_SECONDS", "number")
)

// Config holds the configuration for Swift transport
type Config struct {
	Username       string        // Username for Swift authentication
	APIKey         string        // API key for Swift authentication
	AuthURL        string        // Authentication URL for Swift
	Domain         string        // Domain for Swift authentication
	Tenant         string        // Tenant for Swift authentication
	AuthVersion    int           // Authentication version for Swift
	ConnectTimeout time.Duration // Connection timeout for Swift
	Timeout        time.Duration // Request timeout for Swift
}

// NewDefaultConfig returns a new default configuration for Swift transport
func NewDefaultConfig() Config {
	return Config{
		Username:       "",
		APIKey:         "",
		AuthURL:        "",
		Domain:         "",
		Tenant:         "",
		AuthVersion:    0,
		ConnectTimeout: 10 * time.Second,
		Timeout:        60 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Username, IMGPROXY_SWIFT_USERNAME),
		env.String(&c.APIKey, IMGPROXY_SWIFT_API_KEY),
		env.String(&c.AuthURL, IMGPROXY_SWIFT_AUTH_URL),
		env.String(&c.Domain, IMGPROXY_SWIFT_DOMAIN),
		env.String(&c.Tenant, IMGPROXY_SWIFT_TENANT),
		env.Int(&c.AuthVersion, IMGPROXY_SWIFT_AUTH_VERSION),
		env.Duration(&c.ConnectTimeout, IMGPROXY_SWIFT_CONNECT_TIMEOUT_SECONDS),
		env.Duration(&c.Timeout, IMGPROXY_SWIFT_TIMEOUT_SECONDS),
	)

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
