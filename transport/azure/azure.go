package azure

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/httprange"
)

type transport struct {
	serviceURL *azblob.ServiceURL
}

func New() (http.RoundTripper, error) {
	credential, err := azblob.NewSharedKeyCredential(config.ABSName, config.ABSKey)
	if err != nil {
		return nil, err
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	endpoint := config.ABSEndpoint
	if len(endpoint) == 0 {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", config.ABSName)
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	serviceURL := azblob.NewServiceURL(*endpointURL, pipeline)

	return transport{&serviceURL}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	containerURL := t.serviceURL.NewContainerURL(strings.ToLower(req.URL.Host))
	blobURL := containerURL.NewBlockBlobURL(strings.TrimPrefix(req.URL.Path, "/"))

	start, end, err := httprange.Parse(req.Header.Get("Range"))
	if err != nil {
		return httprange.InvalidHTTPRangeResponse(req), nil
	}

	length := end - start + 1
	if end <= 0 {
		length = azblob.CountToEnd
	}

	get, err := blobURL.Download(req.Context(), start, length, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	if config.ETagEnabled && start == 0 && end == azblob.CountToEnd {
		etag := string(get.ETag())

		if etag == req.Header.Get("If-None-Match") {
			if body := get.Response().Body; body != nil {
				get.Response().Body.Close()
			}

			return &http.Response{
				StatusCode:    http.StatusNotModified,
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header:        make(http.Header),
				ContentLength: 0,
				Body:          nil,
				Close:         false,
				Request:       req,
			}, nil
		}
	}

	return get.Response(), nil
}
