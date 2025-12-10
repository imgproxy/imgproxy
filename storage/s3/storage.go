package s3

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	s3Crypto "github.com/aws/amazon-s3-encryption-client-go/v3/client"
	s3CryptoMaterials "github.com/aws/amazon-s3-encryption-client-go/v3/materials"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/storage"
)

// Storage implements S3 Storage
type Storage struct {
	clientOptions []func(*s3.Options)

	defaultClient s3Client
	defaultConfig aws.Config

	clientsByRegion map[string]s3Client
	clientsByBucket map[string]s3Client

	mu sync.RWMutex

	config *Config
}

// New creates a new S3 storage instance
func New(config *Config, trans *http.Transport) (*Storage, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	conf, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("can't load AWS S3 config: %w", err)
	}

	conf.HTTPClient = &http.Client{Transport: trans}

	if len(config.Region) != 0 {
		conf.Region = config.Region
	}

	if len(conf.Region) == 0 {
		conf.Region = "us-west-1"
	}

	if len(config.AssumeRoleArn) != 0 {
		creds := stscreds.NewAssumeRoleProvider(
			sts.NewFromConfig(conf), config.AssumeRoleArn,
			func(o *stscreds.AssumeRoleOptions) {
				if len(config.AssumeRoleExternalID) != 0 {
					o.ExternalID = aws.String(config.AssumeRoleExternalID)
				}
			})
		conf.Credentials = creds
	}

	clientOptions := []func(*s3.Options){
		func(o *s3.Options) {
			o.DisableLogOutputChecksumValidationSkipped = true
		},
	}

	if len(config.Endpoint) != 0 {
		endpoint := config.Endpoint
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "http://" + endpoint
		}
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = config.EndpointUsePathStyle
		})
	}

	client, err := createClient(conf, clientOptions, config)
	if err != nil {
		return nil, fmt.Errorf("can't create S3 client: %w", err)
	}

	return &Storage{
		clientOptions:   clientOptions,
		defaultClient:   client,
		defaultConfig:   conf,
		clientsByRegion: map[string]s3Client{conf.Region: client},
		clientsByBucket: make(map[string]s3Client),
		config:          config,
	}, nil
}

func (t *Storage) getBucketClient(bucket string) s3Client {
	var client s3Client

	func() {
		t.mu.RLock()
		defer t.mu.RUnlock()
		client = t.clientsByBucket[bucket]
	}()

	if client != nil {
		return client
	}

	return t.defaultClient
}

func (t *Storage) createBucketClient(bucket, region string) (s3Client, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check again if someone did this before us
	if client := t.clientsByBucket[bucket]; client != nil {
		return client, nil
	}

	if client := t.clientsByRegion[region]; client != nil {
		t.clientsByBucket[bucket] = client
		return client, nil
	}

	conf := t.defaultConfig.Copy()
	conf.Region = region

	client, err := createClient(conf, t.clientOptions, t.config)
	if err != nil {
		return nil, errctx.Wrap(err, errctx.WithPrefix("can't create regional S3 client"))
	}

	t.clientsByRegion[region] = client
	t.clientsByBucket[bucket] = client

	return client, nil
}

func createClient(conf aws.Config, opts []func(*s3.Options), config *Config) (s3Client, error) {
	client := s3.NewFromConfig(conf, opts...)

	if config.DecryptionClientEnabled {
		kmsClient := kms.NewFromConfig(conf)
		keyring := s3CryptoMaterials.NewKmsDecryptOnlyAnyKeyKeyring(kmsClient)

		cmm, err := s3CryptoMaterials.NewCryptographicMaterialsManager(keyring)
		if err != nil {
			return nil, err
		}

		return s3Crypto.New(client, cmm)
	} else {
		return client, nil
	}
}

func regionFromError(err error) string {
	var rerr *awsHttp.ResponseError
	if !errors.As(err, &rerr) {
		return ""
	}

	if rerr.Response == nil || rerr.Response.StatusCode != http.StatusMovedPermanently {
		return ""
	}

	return rerr.Response.Header.Get("X-Amz-Bucket-Region")
}

func handleError(err error) (*storage.ObjectReader, error) {
	var rerr *awsHttp.ResponseError
	if !errors.As(err, &rerr) {
		return nil, errctx.Wrap(err)
	}

	if rerr.Response == nil || rerr.Response.StatusCode < 100 || rerr.Response.StatusCode == http.StatusMovedPermanently {
		return nil, errctx.Wrap(err)
	}

	return storage.NewObjectError(rerr.Response.StatusCode, err.Error()), nil
}

// callWithClient is a helper function to call S3 client method with automatic region
// error handling
func callWithClient[T any](s *Storage, bucket string, fn func(client s3Client) (*T, error)) (*T, s3Client, error) {
	client := s.getBucketClient(bucket)

	r, err := fn(client)

	if err != nil {
		// Check if the error is the region mismatch error.
		// If so, create a new client with the correct region and retry the request.
		if region := regionFromError(err); len(region) != 0 {
			client, err = s.createBucketClient(bucket, region)
			if err != nil {
				return nil, nil, err
			}

			r, err = fn(client)
		}
	}

	if err != nil {
		return nil, nil, err
	}

	return r, client, nil
}
