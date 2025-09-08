package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
)

// Config represents HTTP server config
type Config struct {
	Listen                string        // Address to listen on
	Network               string        // Network type (tcp, unix)
	Bind                  string        // Bind address
	PathPrefix            string        // Path prefix for the server
	MaxClients            int           // Maximum number of concurrent clients
	ReadRequestTimeout    time.Duration // Timeout for reading requests
	WriteResponseTimeout  time.Duration // Timeout for writing responses
	KeepAliveTimeout      time.Duration // Timeout for keep-alive connections
	GracefulTimeout       time.Duration // Timeout for graceful shutdown
	CORSAllowOrigin       string        // CORS allowed origin
	Secret                string        // Secret for authorization
	DevelopmentErrorsMode bool          // Enable development mode for detailed error messages
	SocketReusePort       bool          // Enable SO_REUSEPORT socket option
	HealthCheckPath       string        // Health check path from config
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
		WriteResponseTimeout:  10 * time.Second,
		GracefulTimeout:       20 * time.Second,
		CORSAllowOrigin:       "",
		Secret:                "",
		DevelopmentErrorsMode: false,
		SocketReusePort:       false,
		HealthCheckPath:       "",
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

	if c.WriteResponseTimeout <= 0 {
		return fmt.Errorf("write response timeout should be greater than 0, now - %d", c.WriteResponseTimeout)
	}

	if c.KeepAliveTimeout < 0 {
		return fmt.Errorf("keep alive timeout should be greater than or equal to 0, now - %d", c.KeepAliveTimeout)
	}

	if c.GracefulTimeout < 0 {
		return fmt.Errorf("graceful timeout should be greater than or equal to 0, now - %d", c.GracefulTimeout)
	}

	return nil
}
