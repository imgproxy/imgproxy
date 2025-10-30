package azure

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
	"github.com/imgproxy/imgproxy/v3/storage/common"
	"github.com/imgproxy/imgproxy/v3/storage/response"
)

// Storage represents Azure Storage
type Storage struct {
	config *Config
	client *azblob.Client
}

// New creates a new Azure Storage instance
func New(config *Config, trans *http.Transport) (*Storage, error) {
	var (
		client                 *azblob.Client
		sharedKeyCredential    *azblob.SharedKeyCredential
		defaultAzureCredential *azidentity.DefaultAzureCredential
		err                    error
	)

	if err = config.Validate(); err != nil {
		return nil, err
	}

	endpoint := config.Endpoint
	if len(endpoint) == 0 {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", config.Name)
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	opts := azblob.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Transport: &http.Client{Transport: trans},
		},
	}

	if len(config.Key) > 0 {
		sharedKeyCredential, err = azblob.NewSharedKeyCredential(config.Name, config.Key)
		if err != nil {
			return nil, err
		}

		client, err = azblob.NewClientWithSharedKeyCredential(endpointURL.String(), sharedKeyCredential, &opts)
	} else {
		defaultAzureCredential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}

		client, err = azblob.NewClient(endpointURL.String(), defaultAzureCredential, &opts)
	}

	if err != nil {
		return nil, err
	}

	return &Storage{config, client}, nil
}

// GetObject retrieves an object from Azure cloud
func (s *Storage) GetObject(
	ctx context.Context,
	reqHeader http.Header,
	container, key, _ string,
) (*response.Object, error) {
	// If either container or object name is empty, return 404
	if len(container) == 0 || len(key) == 0 {
		return response.NewNotFound(
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
		return response.NewInvalidRange(), nil
	}

	// Open the object
	result, err := s.client.DownloadStream(ctx, container, key, opts)
	if err != nil {
		if azError, ok := err.(*azcore.ResponseError); !ok || azError.StatusCode < 100 || azError.StatusCode == 301 {
			return nil, err
		} else {
			return response.NewError(azError.StatusCode, azError.Error()), nil
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

		return response.NewNotModified(header), nil
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
		return response.NewPartialContent(header, result.Body), nil
	}

	return response.NewOK(header, result.Body), nil
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
