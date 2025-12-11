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
	PORT                             = env.Describe("PORT", "port")
	IMGPROXY_NETWORK                 = env.Describe("IMGPROXY_NETWORK", "tcp|tcp4|tcp6|udp|udp4|udp6|unix|unixgram|unixpacket") //nolint:lll
	IMGPROXY_BIND                    = env.Describe("IMGPROXY_BIND", "address:port, path to unix socket, etc")
	IMGPROXY_PATH_PREFIX             = env.Describe("IMGPROXY_PATH_PREFIX", "string")
	IMGPROXY_MAX_CLIENTS             = env.Describe("IMGPROXY_MAX_CLIENTS", "number, 0 means unlimited")
	IMGPROXY_TIMEOUT                 = env.Describe("IMGPROXY_TIMEOUT", "seconds > 0")
	IMGPROXY_READ_REQUEST_TIMEOUT    = env.Describe("IMGPROXY_READ_REQUEST_TIMEOUT", "seconds > 0")
	IMGPROXY_KEEP_ALIVE_TIMEOUT      = env.Describe("IMGPROXY_KEEP_ALIVE_TIMEOUT", "seconds >= 0")
	IMGPROXY_GRACEFUL_STOP_TIMEOUT   = env.Describe("IMGPROXY_GRACEFUL_STOP_TIMEOUT", "seconds >= 0")
	IMGPROXY_ALLOW_ORIGIN            = env.Describe("IMGPROXY_ALLOW_ORIGIN", "string")
	IMGPROXY_SECRET                  = env.Describe("IMGPROXY_SECRET", "string")
	IMGPROXY_DEVELOPMENT_ERRORS_MODE = env.Describe("IMGPROXY_DEVELOPMENT_ERRORS_MODE", "boolean")
	IMGPROXY_SO_REUSEPORT            = env.Describe("IMGPROXY_SO_REUSEPORT", "boolean")
	IMGPROXY_HEALTH_CHECK_PATH       = env.Describe("IMGPROXY_HEALTH_CHECK_PATH", "string")
	IMGPROXY_FREE_MEMORY_INTERVAL    = env.Describe("IMGPROXY_FREE_MEMORY_INTERVAL", "seconds >= 0")
	IMGPROXY_LOG_MEM_STATS           = env.Describe("IMGPROXY_LOG_MEM_STATS", "boolean")
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

	var port string
	if err := env.String(&port, PORT); err != nil {
		return nil, err
	}

	if len(port) > 0 {
		c.Bind = fmt.Sprintf(":%s", port)
	}

	_, rwErr := responsewriter.LoadConfigFromEnv(&c.ResponseWriter)

	err := errors.Join(
		rwErr,
		env.String(&c.Network, IMGPROXY_NETWORK),
		env.String(&c.Bind, IMGPROXY_BIND),
		env.URLPath(&c.PathPrefix, IMGPROXY_PATH_PREFIX),
		env.Int(&c.MaxClients, IMGPROXY_MAX_CLIENTS),
		env.Duration(&c.RequestTimeout, IMGPROXY_TIMEOUT),
		env.Duration(&c.ReadRequestTimeout, IMGPROXY_READ_REQUEST_TIMEOUT),
		env.Duration(&c.KeepAliveTimeout, IMGPROXY_KEEP_ALIVE_TIMEOUT),
		env.Duration(&c.GracefulStopTimeout, IMGPROXY_GRACEFUL_STOP_TIMEOUT),
		env.String(&c.CORSAllowOrigin, IMGPROXY_ALLOW_ORIGIN),
		env.String(&c.Secret, IMGPROXY_SECRET),
		env.Bool(&c.DevelopmentErrorsMode, IMGPROXY_DEVELOPMENT_ERRORS_MODE),
		env.Bool(&c.SocketReusePort, IMGPROXY_SO_REUSEPORT),
		env.URLPath(&c.HealthCheckPath, IMGPROXY_HEALTH_CHECK_PATH),
		env.Duration(&c.FreeMemoryInterval, IMGPROXY_FREE_MEMORY_INTERVAL),
		env.Bool(&c.LogMemStats, IMGPROXY_LOG_MEM_STATS),
	)

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
