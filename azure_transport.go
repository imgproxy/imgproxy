package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type azureTransport struct {
	serviceURL *azblob.ServiceURL
}

func newAzureTransport() (http.RoundTripper, error) {
	credential, err := azblob.NewSharedKeyCredential(conf.ABSName, conf.ABSKey)
	if err != nil {
		return nil, err
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	endpoint := conf.ABSEndpoint
	if len(endpoint) == 0 {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", conf.ABSName)
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	serviceURL := azblob.NewServiceURL(*endpointURL, pipeline)

	return azureTransport{&serviceURL}, nil
}

func (t azureTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	containerURL := t.serviceURL.NewContainerURL(strings.ToLower(req.URL.Host))
	blobURL := containerURL.NewBlockBlobURL(strings.TrimPrefix(req.URL.Path, "/"))

	get, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	return get.Response(), nil
}
