package server

import (
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
)

const (
	// gracefulTimeout represents graceful shutdown timeout
	gracefulTimeout = time.Duration(5 * time.Second)
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
func NewDefaultConfig() *Config {
	return &Config{
		Network:               "tcp",
		Bind:                  ":8080",
		PathPrefix:            "",
		MaxClients:            2048,
		ReadRequestTimeout:    10 * time.Second,
		KeepAliveTimeout:      10 * time.Second,
		WriteResponseTimeout:  10 * time.Second,
		GracefulTimeout:       gracefulTimeout,
		CORSAllowOrigin:       "",
		Secret:                "",
		DevelopmentErrorsMode: false,
		SocketReusePort:       false,
		HealthCheckPath:       "",
	}
}

// OverrideFromEnv overrides current values with environment variables
func (c *Config) OverrideFromEnv() *Config {
	c.Network = config.Network
	c.Bind = config.Bind
	c.PathPrefix = config.PathPrefix
	c.MaxClients = config.MaxClients
	c.ReadRequestTimeout = time.Duration(config.ReadRequestTimeout) * time.Second
	c.KeepAliveTimeout = time.Duration(config.KeepAliveTimeout) * time.Second
	c.GracefulTimeout = gracefulTimeout
	c.CORSAllowOrigin = config.AllowOrigin
	c.Secret = config.Secret
	c.DevelopmentErrorsMode = config.DevelopmentErrorsMode
	c.SocketReusePort = config.SoReuseport
	c.HealthCheckPath = config.HealthCheckPath

	return c
}

// NewConfigFromEnv creates a default Config instance and overrides values from the
// environment (that's a shortcut)
func NewConfigFromEnv() *Config {
	return NewDefaultConfig().OverrideFromEnv()
}
