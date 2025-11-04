package swift

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ncw/swift/v2"
)

// Storage implements Openstack Swift storage.
type Storage struct {
	config     *Config
	connection *swift.Connection
}

// New creates a new Swift storage with the provided configuration.
func New(
	ctx context.Context,
	config *Config,
	trans *http.Transport,
) (*Storage, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	c := &swift.Connection{
		UserName:       config.Username,
		ApiKey:         config.APIKey,
		AuthUrl:        config.AuthURL,
		AuthVersion:    config.AuthVersion,
		Domain:         config.Domain, // v3 auth only
		Tenant:         config.Tenant, // v2 auth only
		Timeout:        config.Timeout,
		ConnectTimeout: config.ConnectTimeout,
		Transport:      trans,
	}

	err := c.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("swift authentication failed: %v", err)
	}

	return &Storage{
		config:     config,
		connection: c,
	}, nil
}
