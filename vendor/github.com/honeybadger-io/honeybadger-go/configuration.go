package honeybadger

import (
	"log"
	"os"
	"strconv"
	"time"
)

// The Logger interface is implemented by the standard log package and requires
// a limited subset of the interface implemented by log.Logger.
type Logger interface {
	Printf(format string, v ...interface{})
}

// Configuration manages the configuration for the client.
type Configuration struct {
	APIKey          string
	Root            string
	Env             string
	Hostname        string
	Endpoint        string
	Timeout         time.Duration
	Logger          Logger
	Backend         Backend
}

func (c1 *Configuration) update(c2 *Configuration) *Configuration {
	if c2.APIKey != "" {
		c1.APIKey = c2.APIKey
	}
	if c2.Root != "" {
		c1.Root = c2.Root
	}
	if c2.Env != "" {
		c1.Env = c2.Env
	}
	if c2.Hostname != "" {
		c1.Hostname = c2.Hostname
	}
	if c2.Endpoint != "" {
		c1.Endpoint = c2.Endpoint
	}
	if c2.Timeout > 0 {
		c1.Timeout = c2.Timeout
	}
	if c2.Logger != nil {
		c1.Logger = c2.Logger
	}
	if c2.Backend != nil {
		c1.Backend = c2.Backend
	}
	return c1
}

func newConfig(c Configuration) *Configuration {
	config := &Configuration{
		APIKey:          getEnv("HONEYBADGER_API_KEY"),
		Root:            getPWD(),
		Env:             getEnv("HONEYBADGER_ENV"),
		Hostname:        getHostname(),
		Endpoint:        getEnv("HONEYBADGER_ENDPOINT", "https://api.honeybadger.io"),
		Timeout:         getTimeout(),
		Logger:          log.New(os.Stderr, "[honeybadger] ", log.Flags()),
	}
	config.update(&c)

	if config.Backend == nil {
		config.Backend = newServerBackend(config)
	}

	return config
}

func getTimeout() time.Duration {
	if env := getEnv("HONEYBADGER_TIMEOUT"); env != "" {
		if ns, err := strconv.ParseInt(env, 10, 64); err == nil {
			return time.Duration(ns)
		}
	}
	return 3 * time.Second
}

func getEnv(key string, fallback ...string) (val string) {
	val = os.Getenv(key)
	if val == "" && len(fallback) > 0 {
		return fallback[0]
	}
	return
}

func getHostname() (hostname string) {
	if val, err := os.Hostname(); err == nil {
		hostname = val
	}
	return getEnv("HONEYBADGER_HOSTNAME", hostname)
}

func getPWD() (pwd string) {
	if val, err := os.Getwd(); err == nil {
		pwd = val
	}
	return getEnv("HONEYBADGER_ROOT", pwd)
}
