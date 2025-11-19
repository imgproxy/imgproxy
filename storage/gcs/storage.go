package gcs

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"
	htransport "google.golang.org/api/transport/http"
)

// Storage represents Google Cloud Storage implementation
type Storage struct {
	config *Config
	client *gcs.Client
}

// New creates a new Storage instance.
func New(
	config *Config,
	trans *http.Transport,
) (*Storage, error) {
	var client *gcs.Client

	if err := config.Validate(); err != nil {
		return nil, err
	}

	opts := []option.ClientOption{
		option.WithScopes(raw.DevstorageReadOnlyScope),
	}

	if !config.ReadOnly {
		opts = append(opts, option.WithScopes(raw.DevstorageReadWriteScope))
	}

	if len(config.Key) > 0 {
		opts = append(opts, option.WithCredentialsJSON([]byte(config.Key)))
	}

	if len(config.Endpoint) > 0 {
		opts = append(opts, option.WithEndpoint(config.Endpoint))
	}

	if config.TestNoAuth {
		slog.Warn("GCS storage: authentication disabled")
		opts = append(opts, option.WithoutAuthentication())
	}

	htrans, err := htransport.NewTransport(context.TODO(), trans, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating GCS transport: %w", err)
	}

	httpClient := &http.Client{Transport: htrans}
	opts = append(opts, option.WithHTTPClient(httpClient))

	client, err = gcs.NewClient(context.Background(), opts...)

	if err != nil {
		return nil, fmt.Errorf("can't create GCS client: %w", err)
	}

	return &Storage{config, client}, nil
}
