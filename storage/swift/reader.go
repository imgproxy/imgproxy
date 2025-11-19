package swift

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ncw/swift/v2"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/common"
)

// GetObject retrieves an object from Swift storage.
func (s *Storage) GetObject(
	ctx context.Context, reqHeader http.Header, bucket, name, _ string,
) (*storage.ObjectReader, error) {
	// If either bucket or object key is empty, return 404
	if len(bucket) == 0 || len(name) == 0 {
		return storage.NewObjectNotFound(
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
	obj, objectHeaders, err := s.connection.ObjectOpen(ctx, bucket, name, false, h)

	// Convert Swift response headers to normal headers (if any)
	header := make(http.Header)
	for k, v := range objectHeaders {
		header.Set(k, v)
	}

	if err != nil {
		// Handle not found errors gracefully
		if errors.Is(err, swift.ObjectNotFound) || errors.Is(err, swift.ContainerNotFound) {
			return storage.NewObjectNotFound(err.Error()), nil
		}

		// Same for NotModified
		if errors.Is(err, swift.NotModified) {
			return storage.NewObjectNotModified(header), nil
		}

		return nil, fmt.Errorf("error opening swift object: %v", err)
	}

	// Range header: means partial content
	partial := len(reqHeader.Get(httpheaders.Range)) > 0

	// By default, Swift storage handles this.
	// Just in case, let's double check.
	if !partial && common.IsNotModified(reqHeader, header) {
		obj.Close()
		return storage.NewObjectNotModified(header), nil
	}

	return storage.NewObjectOK(header, obj), nil
}
