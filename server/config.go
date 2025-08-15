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
	KeepAliveTimeout      time.Duration // Timeout for keep-alive connections
	GracefulTimeout       time.Duration // Timeout for graceful shutdown
	CORSAllowOrigin       string        // CORS allowed origin
	Secret                string        // Secret for authorization
	DevelopmentErrorsMode bool          // Enable development mode for detailed error messages
	SocketReusePort       bool          // Enable SO_REUSEPORT socket option
	HealthCheckPath       string        // Health check path from config
}

// NewConfigFromEnv creates a new Config instance from environment variables
func NewConfigFromEnv() *Config {
	return &Config{
		Network:               config.Network,
		Bind:                  config.Bind,
		PathPrefix:            config.PathPrefix,
		MaxClients:            config.MaxClients,
		ReadRequestTimeout:    time.Duration(config.ReadRequestTimeout) * time.Second,
		KeepAliveTimeout:      time.Duration(config.KeepAliveTimeout) * time.Second,
		GracefulTimeout:       gracefulTimeout,
		CORSAllowOrigin:       config.AllowOrigin,
		Secret:                config.Secret,
		DevelopmentErrorsMode: config.DevelopmentErrorsMode,
		SocketReusePort:       config.SoReuseport,
		HealthCheckPath:       config.HealthCheckPath,
	}
}
