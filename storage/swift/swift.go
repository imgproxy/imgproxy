package swift

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ncw/swift/v2"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage/common"
	"github.com/imgproxy/imgproxy/v3/storage/response"
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

// GetObject retrieves an object from Swift storage.
func (s *Storage) GetObject(
	ctx context.Context, reqHeader http.Header, bucket, name, _ string,
) (*response.Object, error) {
	// If either bucket or object key is empty, return 404
	if len(bucket) == 0 || len(name) == 0 {
		return response.NewNotFound(
			"invalid Swift URL: bucket name or object name are empty",
		), nil
	}

	// Check if access to the container is allowed
	if !common.IsBucketAllowed(bucket, s.config.AllowedBuckets, s.config.DeniedBuckets) {
		return nil, fmt.Errorf("access to the Swift bucket %s is denied", bucket)
	}

	// Copy if-modified-since, if-none-match and range headers from
	// the original request. They act as the parameters for this storage.
	h := make(swift.Headers)

	for _, k := range []string{
		httpheaders.Range,           // Range for partial requests
		httpheaders.IfNoneMatch,     // If-None-Match for caching
		httpheaders.IfModifiedSince, // If-Modified-Since for caching
	} {
		v := reqHeader.Get(k)
		if len(v) > 0 {
			h[k] = v
		}
	}

	// Fetch the object from Swift
	object, objectHeaders, err := s.connection.ObjectOpen(ctx, bucket, name, false, h)

	// Convert Swift response headers to normal headers (if any)
	header := make(http.Header)
	for k, v := range objectHeaders {
		header.Set(k, v)
	}

	if err != nil {
		// Handle not found errors gracefully
		if errors.Is(err, swift.ObjectNotFound) || errors.Is(err, swift.ContainerNotFound) {
			return response.NewNotFound(err.Error()), nil
		}

		// Same for NotModified
		if errors.Is(err, swift.NotModified) {
			return response.NewNotModified(header), nil
		}

		return nil, fmt.Errorf("error opening swift object: %v", err)
	}

	// Range header: means partial content
	partial := len(reqHeader.Get(httpheaders.Range)) > 0

	// By default, Swift storage handles this.
	// Just in case, let's double check.
	if !partial && common.IsNotModified(reqHeader, header) {
		object.Close()
		return response.NewNotModified(header), nil
	}

	return response.NewOK(header, object), nil
}
