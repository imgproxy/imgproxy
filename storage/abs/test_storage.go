package abs

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

// absStorageWrapper wraps the storage and optionally holds a server for cleanup
type absStorageWrapper struct {
	*Storage

	server      *TestServer
	client      *azblob.Client
	shouldClose bool
}

// Server returns the underlying AbsServer
func (w *absStorageWrapper) Server() *TestServer {
	return w.server
}

// Client returns the underlying ABS client for direct API access
func (w *absStorageWrapper) Client() *azblob.Client {
	return w.client
}

// LazySuiteStorage is a sugar alias
type LazySuiteStorage = testutil.LazyObj[*absStorageWrapper]

// NewLazySuiteStorage creates a lazy ABS Storage object for use in test suites
// A new server will be created internally and cleaned up automatically
func NewLazySuiteStorage(
	l testutil.LazySuiteFrom,
) (testutil.LazyObj[*absStorageWrapper], context.CancelFunc) {
	return testutil.NewLazySuiteObj(
		l,
		func() (*absStorageWrapper, error) {
			wrapper := &absStorageWrapper{}

			// Create server internally
			absServer, err := NewAbsServer()
			if err != nil {
				return nil, err
			}
			wrapper.server = absServer
			wrapper.shouldClose = true

			config := NewDefaultConfig()
			config.Endpoint = absServer.URL()
			config.Name = "testaccount"
			config.Key = "dGVzdGtleQ=="

			c := generichttp.NewDefaultConfig()
			c.IgnoreSslVerification = true

			trans, err := generichttp.New(false, &c)
			if err != nil {
				return nil, err
			}

			storage, err := New(&config, trans)
			if err != nil {
				return nil, err
			}

			// Create ABS client for direct API access
			sharedKeyCredential, err := azblob.NewSharedKeyCredential(config.Name, config.Key)
			if err != nil {
				return nil, err
			}

			clientOpts := &azblob.ClientOptions{
				ClientOptions: policy.ClientOptions{
					Transport: &http.Client{Transport: trans},
				},
			}

			client, err := azblob.NewClientWithSharedKeyCredential(absServer.URL(), sharedKeyCredential, clientOpts)
			if err != nil {
				return nil, err
			}

			wrapper.Storage = storage
			wrapper.client = client
			return wrapper, nil
		},
		func(w *absStorageWrapper) error {
			// Clean up internal server if we created it
			if w.shouldClose {
				w.server.Close()
			}
			return nil
		},
	)
}
