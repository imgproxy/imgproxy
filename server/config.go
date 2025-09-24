package server

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/server/responsewriter"
)

type EnvDesc struct {
	Name        string
	Description string // is not used programmatically, most likely not needed
	Format      string
}

var (
	network = EnvDesc{
		Name:        "IMGPROXY_NETWORK",
		Description: "Network type (tcp, unix)",
		Format:      "tcp|udp|unix",
	}

	bind = EnvDesc{
		Name:        "IMGPROXY_BIND",
		Description: "Address to bind the server to",
		Format:      "address:port",
	}

	maxClients = EnvDesc{
		Name:        "IMGPROXY_MAX_CLIENTS",
		Description: "Maximum number of concurrent clients",
		Format:      "number >= 0",
	}

	readRequestTimeout = EnvDesc{
		Name:        "IMGPROXY_READ_REQUEST_TIMEOUT",
		Description: "Timeout for reading requests",
		Format:      "number >= 0, seconds",
	}
)

// Getenv returns the value of the env variable
func (d *EnvDesc) Lookup() (string, bool) {
	return os.LookupEnv(d.Name)
}

// WarnParseError logs a warning when an env var fails to parse
func (d *EnvDesc) WarnParseError(err error, value any) {
	v, _ := d.Lookup()

	slog.Warn(
		"failed to parse env var, using default value",
		"name", d.Name,
		"format", d.Format,
		"value", v,
		"error", err,
		"default", value,
	)
}

func (d *EnvDesc) Errorf(msg string, args ...any) error {
	return fmt.Errorf(
		"invalid %s value (format: %s): %s",
		d.Name,
		d.Format,
		fmt.Sprintf(msg, args...),
	)
}

// configurators.Int
func Int(i *int, desc *EnvDesc) {
	env, ok := desc.Lookup()
	if !ok {
		return
	}

	value, err := strconv.Atoi(env)
	if err != nil {
		desc.WarnParseError(err, *i)
		return
	}
	*i = value
}

// configurators.Duration (in seconds)
func Duration(d *time.Duration, desc *EnvDesc) {
	env, ok := desc.Lookup()
	if !ok {
		return
	}

	value, err := strconv.Atoi(env)
	if err != nil {
		desc.WarnParseError(err, *d)
		return
	}
	*d = time.Duration(value) * time.Second
}

// configurators.String
func String(s *string, desc *EnvDesc) {
	if env, ok := desc.Lookup(); ok {
		// No warning here: empty string is a valid value, it has no format
		*s = env
	}
}

// Config represents HTTP server config
type Config struct {
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

	String(&config.Network, &network)
	String(&c.Bind, &bind)
	c.Bind = config.Bind
	c.PathPrefix = config.PathPrefix
	Int(&c.MaxClients, &maxClients)
	Duration(&c.ReadRequestTimeout, &readRequestTimeout)
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
		return bind.Errorf("should not be empty")
	}

	if c.MaxClients < 0 {
		return maxClients.Errorf("current value: %v", c.MaxClients)
	}

	if c.ReadRequestTimeout <= 0 {
		return readRequestTimeout.Errorf("current value: %v", c.ReadRequestTimeout)
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
