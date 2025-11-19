package abs

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
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
