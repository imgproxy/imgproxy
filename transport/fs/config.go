package fs

import (
	"errors"
	"fmt"
	"os"

	"github.com/imgproxy/imgproxy/v3/config"
)

// Config holds the configuration for local file system transport
type Config struct {
	Root string // Root directory for the local file system transport
}

// NewDefaultConfig returns a new default configuration for local file system transport
func NewDefaultConfig() *Config {
	return &Config{
		Root: "",
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	if c == nil {
		c = NewDefaultConfig()
	}

	c.Root = config.LocalFileSystemRoot

	return c, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Root == "" {
		return errors.New("local file system root shold not be blank")
	}

	stat, err := os.Stat(c.Root)
	if err != nil {
		return fmt.Errorf("cannot use local directory: %s", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("cannot use local directory: not a directory")
	}

	if c.Root == "/" {
		// Warning: exposing root is unsafe
		// TODO: Move this somewhere to the instance checks (?)
		fmt.Println("Warning: Exposing root via IMGPROXY_LOCAL_FILESYSTEM_ROOT is unsafe")
	}

	return nil
}
