package abs

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/common"
)

// GetObject retrieves an object from Azure cloud
func (s *Storage) GetObject(
	ctx context.Context,
	reqHeader http.Header,
	container, key, _ string,
) (*storage.ObjectReader, error) {
	// If either container or object name is empty, return 404
	if len(container) == 0 || len(key) == 0 {
		return storage.NewObjectNotFound(
			"invalid Azure Storage URL: container name or object key are empty",
		), nil
	}

	// Check if access to the container is allowed
	if !common.IsBucketAllowed(container, s.config.AllowedBuckets, s.config.DeniedBuckets) {
		return nil, fmt.Errorf("access to the Azure Storage container %s is denied", container)
	}

	header := make(http.Header)
	opts := &blob.DownloadStreamOptions{}

	// Check if this is partial request
	partial, err := parseRangeHeader(opts, reqHeader)
	if err != nil {
		return storage.NewObjectInvalidRange(), nil //nolint:nilerr
	}

	// Open the object
	result, err := s.client.DownloadStream(ctx, container, key, opts)
	if err != nil {
		//nolint:errorlint
		azError, ok := err.(*azcore.ResponseError)
		if !ok || azError.StatusCode < 100 || azError.StatusCode == http.StatusMovedPermanently {
			return nil, err
		} else {
			return storage.NewObjectError(azError.StatusCode, azError.Error()), nil
		}
	}

	// Pass through etag and last modified
	if result.ETag != nil {
		etag := string(*result.ETag)
		header.Set(httpheaders.Etag, etag)
	}

	if result.LastModified != nil {
		lastModified := result.LastModified.Format(http.TimeFormat)
		header.Set(httpheaders.LastModified, lastModified)
	}

	// Break early if response was not modified
	if !partial && common.IsNotModified(reqHeader, header) {
		if result.Body != nil {
			result.Body.Close()
		}

		return storage.NewObjectNotModified(header), nil
	}

	// Pass through important headers
	header.Set(httpheaders.AcceptRanges, "bytes")

	if result.ContentLength != nil {
		header.Set(httpheaders.ContentLength, strconv.FormatInt(*result.ContentLength, 10))
	}

	if result.ContentType != nil {
		header.Set(httpheaders.ContentType, *result.ContentType)
	}

	if result.ContentRange != nil {
		header.Set(httpheaders.ContentRange, *result.ContentRange)
	}

	if result.CacheControl != nil {
		header.Set(httpheaders.CacheControl, *result.CacheControl)
	}

	// If the request was partial, let's respond with partial
	if partial {
		return storage.NewObjectPartialContent(header, result.Body), nil
	}

	return storage.NewObjectOK(header, result.Body), nil
}

func parseRangeHeader(opts *blob.DownloadStreamOptions, reqHeader http.Header) (bool, error) {
	r := reqHeader.Get(httpheaders.Range)
	if len(r) == 0 {
		return false, nil
	}

	start, end, err := httprange.Parse(r)
	if err != nil {
		return false, err
	}

	if end == 0 {
		return false, nil
	}

	length := end - start + 1
	if end <= 0 {
		length = blockblob.CountToEnd
	}

	opts.Range = blob.HTTPRange{
		Offset: start,
		Count:  length,
	}

	return true, nil
}
