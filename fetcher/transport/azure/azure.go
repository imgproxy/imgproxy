package azure

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/common"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/notmodified"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
)

type transport struct {
	client      *azblob.Client
	qsSeparator string
}

func New(config *Config, trans *http.Transport, sep string) (http.RoundTripper, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var (
		client                 *azblob.Client
		sharedKeyCredential    *azblob.SharedKeyCredential
		defaultAzureCredential *azidentity.DefaultAzureCredential
		err                    error
	)

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

	return transport{client, sep}, nil
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	container, key, _ := common.GetBucketAndKey(req.URL, t.qsSeparator)

	if len(container) == 0 || len(key) == 0 {
		body := strings.NewReader("Invalid ABS URL: container name or object key is empty")
		return &http.Response{
			StatusCode:    http.StatusNotFound,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Header:        http.Header{httpheaders.ContentType: {"text/plain"}},
			ContentLength: int64(body.Len()),
			Body:          io.NopCloser(body),
			Close:         false,
			Request:       req,
		}, nil
	}

	statusCode := http.StatusOK

	header := make(http.Header)
	opts := &blob.DownloadStreamOptions{}

	if r := req.Header.Get(httpheaders.Range); len(r) != 0 {
		start, end, err := httprange.Parse(r)
		if err != nil {
			return httprange.InvalidHTTPRangeResponse(req), nil
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

	result, err := t.client.DownloadStream(req.Context(), container, key, opts)
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
				Header:        http.Header{"Content-Type": {"text/plain"}},
				ContentLength: int64(body.Len()),
				Body:          io.NopCloser(body),
				Close:         false,
				Request:       req,
			}, nil
		}
	}

	if result.ETag != nil {
		etag := string(*result.ETag)
		header.Set(httpheaders.Etag, etag)
	}

	if result.LastModified != nil {
		lastModified := result.LastModified.Format(http.TimeFormat)
		header.Set(httpheaders.LastModified, lastModified)
	}

	if resp := notmodified.Response(req, header); resp != nil {
		if result.Body != nil {
			result.Body.Close()
		}
		return resp, nil
	}

	header.Set(httpheaders.AcceptRanges, "bytes")

	contentLength := int64(0)
	if result.ContentLength != nil {
		contentLength = *result.ContentLength
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

	return &http.Response{
		StatusCode:    statusCode,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: contentLength,
		Body:          result.Body,
		Close:         true,
		Request:       req,
	}, nil
}
