package azure

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ctxreader"
)

type transport struct {
	svc *azblob.Client
}

func New() (http.RoundTripper, error) {
	var (
		client                 *azblob.Client
		defaultAzureCredential *azidentity.DefaultAzureCredential
		err                    error
		sharedKeyCredential    *azblob.SharedKeyCredential
	)

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

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	container := req.URL.Host
	key := req.URL.Path

	header := make(http.Header)

	result, err := t.svc.DownloadStream(req.Context(), container, strings.TrimPrefix(key, "/"), nil)
	if err != nil {
		if azError, ok := err.(*azcore.ResponseError); !ok || azError.StatusCode < 100 || azError.StatusCode == 301 {
			return nil, err
		} else {
			return &http.Response{
				StatusCode:    azError.StatusCode,
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
	}

	if config.ETagEnabled {
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

	return &http.Response{
		StatusCode:    http.StatusOK,
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
