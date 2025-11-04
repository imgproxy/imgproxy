package fs

import "net/http"

// Storage represents fs file storage
type Storage struct {
	fs     http.Dir
	config *Config
}

// New creates a new Storage instance.
func New(config *Config) (*Storage, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Storage{config: config, fs: http.Dir(config.Root)}, nil
}
