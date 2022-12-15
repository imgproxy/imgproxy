package azure

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/httprange"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ctxreader"
)

type transport struct {
	client *azblob.Client
}

func New() (http.RoundTripper, error) {
	var (
		client                 *azblob.Client
		defaultAzureCredential *azidentity.DefaultAzureCredential
		err                    error
		sharedKeyCredential    *azblob.SharedKeyCredential
	)

	if config.ABSName == "" {
		return nil, errors.New("IMGPROXY_ABS_NAME must be set")
	}

	endpoint := config.ABSEndpoint
	if len(endpoint) == 0 {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", config.ABSName)
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if config.ABSKey != "" {
		sharedKeyCredential, err = azblob.NewSharedKeyCredential(config.ABSName, config.ABSKey)
		if err != nil {
			return nil, err
		}

		client, err = azblob.NewClientWithSharedKeyCredential(endpointURL.String(), sharedKeyCredential, nil)
	} else {
		defaultAzureCredential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}

		client, err = azblob.NewClient(endpointURL.String(), defaultAzureCredential, nil)
	}

	if err != nil {
		return nil, err
	}

	return transport{client}, nil
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	container := req.URL.Host
	key := req.URL.Path

	var statusCode = http.StatusOK

	header := make(http.Header)
	opts := &blob.DownloadStreamOptions{}

	if r := req.Header.Get("Range"); len(r) != 0 {
		start, end, err := httprange.Parse(r)
		if err != nil {
			return httprange.InvalidHTTPRangeResponse(req), err
		}

		if end != 0 {
			length := end - start + 1
			if end <= 0 {
				length = blockblob.CountToEnd
			}

			opts.Range = blob.HTTPRange{
				Offset: start,
				Count:  length,
			}
		}

		statusCode = http.StatusPartialContent
	}

	result, err := t.client.DownloadStream(req.Context(), container, strings.TrimPrefix(key, "/"), opts)
	if err != nil {
		if azError, ok := err.(*azcore.ResponseError); !ok || azError.StatusCode < 100 || azError.StatusCode == 301 {
			return nil, err
		} else {
			body := strings.NewReader(azError.Error())
			return &http.Response{
				StatusCode:    azError.StatusCode,
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header:        header,
				ContentLength: int64(body.Len()),
				Body:          io.NopCloser(body),
				Close:         false,
				Request:       req,
			}, nil
		}
	}

	if config.ETagEnabled && result.ETag != nil {
		azETag := string(*result.ETag)
		header.Set("ETag", azETag)

		if etag := req.Header.Get("If-None-Match"); len(etag) > 0 && azETag == etag {
			return &http.Response{
				StatusCode:    http.StatusNotModified,
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header:        header,
				ContentLength: 0,
				Body:          nil,
				Close:         false,
				Request:       req,
			}, nil
		}
	}

	header.Set("Accept-Ranges", "bytes")

	if result.ContentLength != nil {
		header.Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))
	}

	if result.ContentType != nil {
		header.Set("Content-Type", *result.ContentType)
	}

	if result.ContentRange != nil {
		header.Set("Content-Range", *result.ContentRange)
	}

	if result.CacheControl != nil {
		header.Set("Cache-Control", *result.CacheControl)
	}

	return &http.Response{
		StatusCode:    statusCode,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: *result.ContentLength,
		Body:          ctxreader.New(req.Context(), result.Body, true),
		Close:         true,
		Request:       req,
	}, nil
}
