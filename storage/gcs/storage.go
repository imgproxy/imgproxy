package gcs

import (
	"context"
	"log/slog"
	"net/http"

	gcs "cloud.google.com/go/storage"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "error creating GCS transport")
	}

	httpClient := &http.Client{Transport: htrans}
	opts = append(opts, option.WithHTTPClient(httpClient))

	client, err = gcs.NewClient(context.Background(), opts...)

	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("Can't create GCS client"))
	}

	return &Storage{config, client}, nil
}
