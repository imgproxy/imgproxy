package gcs

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/storage/common"
	"github.com/imgproxy/imgproxy/v3/storage/response"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"
	htransport "google.golang.org/api/transport/http"
)

// Storage represents Google Cloud Storage implementation
type Storage struct {
	config *Config
	client *storage.Client
}

// New creates a new Storage instance.
func New(
	config *Config,
	trans *http.Transport,
	auth bool, // use authentication, should be false in tests
) (*Storage, error) {
	var client *storage.Client

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

	if !auth {
		slog.Warn("GCS storage: authentication disabled")
		opts = append(opts, option.WithoutAuthentication())
	}

	htrans, err := htransport.NewTransport(context.TODO(), trans, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCS transport")
	}

	httpClient := &http.Client{Transport: htrans}
	opts = append(opts, option.WithHTTPClient(httpClient))

	client, err = storage.NewClient(context.Background(), opts...)

	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("Can't create GCS client"))
	}

	return &Storage{config, client}, nil
}

// GetObject retrieves an object from Azure cloud
func (s *Storage) GetObject(
	ctx context.Context,
	reqHeader http.Header,
	bucket, key, query string,
) (*response.Object, error) {
	// If either bucket or object key is empty, return 404
	if len(bucket) == 0 || len(key) == 0 {
		return response.NewNotFound(
			"invalid GCS Storage URL: bucket name or object key are empty",
		), nil
	}

	// Check if access to the bucket is allowed
	if !common.IsBucketAllowed(bucket, s.config.AllowedBuckets, s.config.DeniedBuckets) {
		return nil, fmt.Errorf("access to the GCS bucket %s is denied", bucket)
	}

	var (
		reader *storage.Reader
		size   int64
	)

	bkt := s.client.Bucket(bucket)
	obj := bkt.Object(key)

	if g, err := strconv.ParseInt(query, 10, 64); err == nil && g > 0 {
		obj = obj.Generation(g)
	}

	header := make(http.Header)

	// Try respond with partial: if that was a partial request,
	// we either return error or Object
	if r, err := s.tryRespondWithPartial(ctx, obj, reqHeader, header); r != nil || err != nil {
		return r, err
	}

	attrs, aerr := obj.Attrs(ctx)
	if aerr != nil {
		return handleError(aerr)
	}
	header.Set(httpheaders.Etag, attrs.Etag)
	header.Set(httpheaders.LastModified, attrs.Updated.Format(http.TimeFormat))

	if common.IsNotModified(reqHeader, header) {
		return response.NewNotModified(header), nil
	}

	var err error
	reader, err = obj.NewReader(ctx)
	if err != nil {
		return handleError(err)
	}

	size = reader.Attrs.Size
	setHeadersFromReader(header, reader, size)

	return response.NewOK(header, reader), nil
}

// tryRespondWithPartial tries to respond with a partial object
// if the Range header is set.
func (s *Storage) tryRespondWithPartial(
	ctx context.Context,
	obj *storage.ObjectHandle,
	reqHeader http.Header,
	header http.Header,
) (*response.Object, error) {
	r := reqHeader.Get(httpheaders.Range)
	if len(r) == 0 {
		return nil, nil
	}

	start, end, err := httprange.Parse(r)
	if err != nil {
		return response.NewInvalidRange(), nil
	}

	if end == 0 {
		return nil, nil
	}

	length := end - start + 1
	if end < 0 {
		length = -1
	}

	reader, err := obj.NewRangeReader(ctx, start, length)
	if err != nil {
		return nil, err
	}

	if end < 0 || end >= reader.Attrs.Size {
		end = reader.Attrs.Size - 1
	}

	size := end - reader.Attrs.StartOffset + 1

	header.Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", reader.Attrs.StartOffset, end, reader.Attrs.Size))
	setHeadersFromReader(header, reader, size)

	return response.NewPartialContent(header, reader), nil
}

func handleError(err error) (*response.Object, error) {
	if err != storage.ErrBucketNotExist && err != storage.ErrObjectNotExist {
		return nil, err
	}

	return response.NewNotFound(err.Error()), nil
}

func setHeadersFromReader(header http.Header, reader *storage.Reader, size int64) {
	header.Set(httpheaders.AcceptRanges, "bytes")
	header.Set(httpheaders.ContentLength, strconv.Itoa(int(size)))
	header.Set(httpheaders.ContentType, reader.Attrs.ContentType)
	header.Set(httpheaders.CacheControl, reader.Attrs.CacheControl)
}
