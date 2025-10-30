package storage

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/storage/response"
)

// Reader represents a generic storage interface, which can read
// objects from a storage backend.
type Reader interface {
	// GetObject retrieves an object from the storage and returns
	// ObjectReader with the result.
	GetObject(
		ctx context.Context,
		reqHeader http.Header,
		bucket, key, query string,
	) (*response.Object, error)
}
