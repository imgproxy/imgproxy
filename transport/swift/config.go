package swift

import (
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
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
func NewDefaultConfig() *Config {
	return &Config{
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

// LoadFromEnv loads configuration from the global config package
func LoadFromEnv(c *Config) (*Config, error) {
	c.Username = config.SwiftUsername
	c.APIKey = config.SwiftAPIKey
	c.AuthURL = config.SwiftAuthURL
	c.Domain = config.SwiftDomain
	c.Tenant = config.SwiftTenant
	c.AuthVersion = config.SwiftAuthVersion
	c.ConnectTimeout = time.Duration(config.SwiftConnectTimeoutSeconds) * time.Second
	c.Timeout = time.Duration(config.SwiftTimeoutSeconds) * time.Second

	return c, nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
