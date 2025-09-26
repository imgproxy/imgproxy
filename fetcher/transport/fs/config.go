package fs

import (
	"log/slog"
	"os"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_LOCAL_FILESYSTEM_ROOT = env.Describe("IMGPROXY_LOCAL_FILESYSTEM_ROOT", "path")
)

// Config holds the configuration for local file system transport
type Config struct {
	Root string // Root directory for the local file system transport
}

// NewDefaultConfig returns a new default configuration for local file system transport
func NewDefaultConfig() Config {
	return Config{
		Root: "",
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := env.String(&c.Root, IMGPROXY_LOCAL_FILESYSTEM_ROOT)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	e := IMGPROXY_LOCAL_FILESYSTEM_ROOT

	if c.Root == "" {
		return e.ErrorEmpty()
	}

	stat, err := os.Stat(c.Root)
	if err != nil {
		return e.Errorf("cannot use local directory: %s", err)
	}

	if !stat.IsDir() {
		return e.Errorf("cannot use local directory: not a directory")
	}

	if c.Root == "/" {
		slog.Warn("Exposing root via IMGPROXY_LOCAL_FILESYSTEM_ROOT is unsafe")
	}

	return nil
}
