package server

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/server/responsewriter"
)

// Config represents HTTP server config
type Config struct {
	Listen                string        // Address to listen on
	Network               string        // Network type (tcp, unix)
	Bind                  string        // Bind address
	PathPrefix            string        // Path prefix for the server
	MaxClients            int           // Maximum number of concurrent clients
	ReadRequestTimeout    time.Duration // Timeout for reading requests
	KeepAliveTimeout      time.Duration // Timeout for keep-alive connections
	GracefulTimeout       time.Duration // Timeout for graceful shutdown
	CORSAllowOrigin       string        // CORS allowed origin
	Secret                string        // Secret for authorization
	DevelopmentErrorsMode bool          // Enable development mode for detailed error messages
	SocketReusePort       bool          // Enable SO_REUSEPORT socket option
	HealthCheckPath       string        // Health check path from config

	ResponseWriter responsewriter.Config // Response writer config

	// TODO: We are not sure where to put it yet
	FreeMemoryInterval time.Duration // Interval for freeing memory
	LogMemStats        bool          // Log memory stats
}

// NewDefaultConfig returns default config values
func NewDefaultConfig() Config {
	return Config{
		Network:               "tcp",
		Bind:                  ":8080",
		PathPrefix:            "",
		MaxClients:            2048,
		ReadRequestTimeout:    10 * time.Second,
		KeepAliveTimeout:      10 * time.Second,
		GracefulTimeout:       20 * time.Second,
		CORSAllowOrigin:       "",
		Secret:                "",
		DevelopmentErrorsMode: false,
		SocketReusePort:       false,
		HealthCheckPath:       "",
		FreeMemoryInterval:    10 * time.Second,
		LogMemStats:           false,

		ResponseWriter: responsewriter.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv overrides current values with environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.Network = config.Network
	c.Bind = config.Bind
	c.PathPrefix = config.PathPrefix
	c.MaxClients = config.MaxClients
	c.ReadRequestTimeout = time.Duration(config.ReadRequestTimeout) * time.Second
	c.KeepAliveTimeout = time.Duration(config.KeepAliveTimeout) * time.Second
	c.GracefulTimeout = time.Duration(config.GracefulStopTimeout) * time.Second
	c.CORSAllowOrigin = config.AllowOrigin
	c.Secret = config.Secret
	c.DevelopmentErrorsMode = config.DevelopmentErrorsMode
	c.SocketReusePort = config.SoReuseport
	c.HealthCheckPath = config.HealthCheckPath
	c.FreeMemoryInterval = time.Duration(config.FreeMemoryInterval) * time.Second
	c.LogMemStats = len(os.Getenv("IMGPROXY_LOG_MEM_STATS")) > 0

	if _, err := responsewriter.LoadConfigFromEnv(&c.ResponseWriter); err != nil {
		return nil, err
	}

	return c, nil
}

// Validate checks that the config values are valid
func (c *Config) Validate() error {
	if len(c.Bind) == 0 {
		return errors.New("bind address is not defined")
	}

	if c.MaxClients < 0 {
		return fmt.Errorf("max clients number should be greater than or equal 0, now - %d", c.MaxClients)
	}

	if c.ReadRequestTimeout <= 0 {
		return fmt.Errorf("read request timeout should be greater than 0, now - %d", c.ReadRequestTimeout)
	}

	if c.KeepAliveTimeout < 0 {
		return fmt.Errorf("keep alive timeout should be greater than or equal to 0, now - %d", c.KeepAliveTimeout)
	}

	if c.GracefulTimeout < 0 {
		return fmt.Errorf("graceful timeout should be greater than or equal to 0, now - %d", c.GracefulTimeout)
	}

	if c.FreeMemoryInterval <= 0 {
		return errors.New("free memory interval should be greater than zero")
	}

	return nil
}
