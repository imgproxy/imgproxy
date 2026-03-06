package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/server/responsewriter"
)

var (
	// networks defines allowed network types
	networks = map[string]string{
		"tcp":        "tcp",
		"tcp4":       "tcp4",
		"tcp6":       "tcp6",
		"unix":       "unix",
		"unixpacket": "unixpacket",
	}

	PORT                             = env.Int("PORT")
	IMGPROXY_NETWORK                 = env.Enum("IMGPROXY_NETWORK", networks)
	IMGPROXY_BIND                    = env.String("IMGPROXY_BIND").WithFormat("bind:port")
	IMGPROXY_PATH_PREFIX             = env.String("IMGPROXY_PATH_PREFIX")
	IMGPROXY_MAX_CLIENTS             = env.Int("IMGPROXY_MAX_CLIENTS")
	IMGPROXY_TIMEOUT                 = env.Duration("IMGPROXY_TIMEOUT")
	IMGPROXY_READ_REQUEST_TIMEOUT    = env.Duration("IMGPROXY_READ_REQUEST_TIMEOUT")
	IMGPROXY_KEEP_ALIVE_TIMEOUT      = env.Duration("IMGPROXY_KEEP_ALIVE_TIMEOUT")
	IMGPROXY_GRACEFUL_STOP_TIMEOUT   = env.Duration("IMGPROXY_GRACEFUL_STOP_TIMEOUT")
	IMGPROXY_ALLOW_ORIGIN            = env.String("IMGPROXY_ALLOW_ORIGIN")
	IMGPROXY_SECRET                  = env.String("IMGPROXY_SECRET")
	IMGPROXY_DEVELOPMENT_ERRORS_MODE = env.Bool("IMGPROXY_DEVELOPMENT_ERRORS_MODE")
	IMGPROXY_SO_REUSEPORT            = env.Bool("IMGPROXY_SO_REUSEPORT")
	IMGPROXY_HEALTH_CHECK_PATH       = env.String("IMGPROXY_HEALTH_CHECK_PATH")
	IMGPROXY_FREE_MEMORY_INTERVAL    = env.Duration("IMGPROXY_FREE_MEMORY_INTERVAL")
	IMGPROXY_LOG_MEM_STATS           = env.Bool("IMGPROXY_LOG_MEM_STATS")
)

// Config represents HTTP server config
type Config struct {
	Network               string        // Network type (tcp, unix)
	Bind                  string        // Bind address
	PathPrefix            string        // Path prefix for the server
	MaxClients            int           // Maximum number of concurrent clients
	RequestTimeout        time.Duration // Timeout for requests
	ReadRequestTimeout    time.Duration // Timeout for reading requests
	KeepAliveTimeout      time.Duration // Timeout for keep-alive connections
	GracefulStopTimeout   time.Duration // Timeout for graceful shutdown
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
		RequestTimeout:        10 * time.Second,
		ReadRequestTimeout:    10 * time.Second,
		KeepAliveTimeout:      10 * time.Second,
		GracefulStopTimeout:   20 * time.Second,
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

	_, rwErr := responsewriter.LoadConfigFromEnv(&c.ResponseWriter)

	port := -1
	bind := ""

	err := errors.Join(
		rwErr,
		PORT.Parse(&port),
		IMGPROXY_BIND.Parse(&bind),
		IMGPROXY_NETWORK.Parse(&c.Network),
		IMGPROXY_PATH_PREFIX.Parse(&c.PathPrefix),
		IMGPROXY_MAX_CLIENTS.Parse(&c.MaxClients),
		IMGPROXY_TIMEOUT.Parse(&c.RequestTimeout),
		IMGPROXY_READ_REQUEST_TIMEOUT.Parse(&c.ReadRequestTimeout),
		IMGPROXY_KEEP_ALIVE_TIMEOUT.Parse(&c.KeepAliveTimeout),
		IMGPROXY_GRACEFUL_STOP_TIMEOUT.Parse(&c.GracefulStopTimeout),
		IMGPROXY_ALLOW_ORIGIN.Parse(&c.CORSAllowOrigin),
		IMGPROXY_SECRET.Parse(&c.Secret),
		IMGPROXY_DEVELOPMENT_ERRORS_MODE.Parse(&c.DevelopmentErrorsMode),
		IMGPROXY_SO_REUSEPORT.Parse(&c.SocketReusePort),
		IMGPROXY_HEALTH_CHECK_PATH.Parse(&c.HealthCheckPath),
		IMGPROXY_FREE_MEMORY_INTERVAL.Parse(&c.FreeMemoryInterval),
		IMGPROXY_LOG_MEM_STATS.Parse(&c.LogMemStats),
	)

	switch {
	case len(bind) > 0:
		c.Bind = bind
	case port > -1:
		c.Bind = fmt.Sprintf(":%d", port)
	}

	return c, err
}

// Validate checks that the config values are valid
func (c *Config) Validate() error {
	if len(c.Bind) == 0 {
		return IMGPROXY_BIND.ErrorEmpty()
	}

	if c.MaxClients < 0 {
		return IMGPROXY_MAX_CLIENTS.ErrorNegative()
	}

	if c.RequestTimeout <= 0 {
		return IMGPROXY_TIMEOUT.ErrorZeroOrNegative()
	}

	if c.ReadRequestTimeout <= 0 {
		return IMGPROXY_READ_REQUEST_TIMEOUT.ErrorZeroOrNegative()
	}

	if c.KeepAliveTimeout < 0 {
		return IMGPROXY_KEEP_ALIVE_TIMEOUT.ErrorNegative()
	}

	if c.GracefulStopTimeout < 0 {
		return IMGPROXY_GRACEFUL_STOP_TIMEOUT.ErrorNegative()
	}

	if c.FreeMemoryInterval <= 0 {
		return IMGPROXY_FREE_MEMORY_INTERVAL.ErrorZeroOrNegative()
	}

	return nil
}
