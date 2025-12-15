package transport

import (
	"errors"
	"os"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	azure "github.com/imgproxy/imgproxy/v3/storage/abs"
	"github.com/imgproxy/imgproxy/v3/storage/fs"
	"github.com/imgproxy/imgproxy/v3/storage/gcs"
	"github.com/imgproxy/imgproxy/v3/storage/s3"
	"github.com/imgproxy/imgproxy/v3/storage/swift"
)

var (
	IMGPROXY_USE_ABS                    = env.Bool("IMGPROXY_USE_ABS")
	IMGPROXY_USE_GCS                    = env.Bool("IMGPROXY_USE_GCS")
	IMGPROXY_USE_S3                     = env.Bool("IMGPROXY_USE_S3")
	IMGPROXY_USE_SWIFT                  = env.Bool("IMGPROXY_USE_SWIFT")
	IMGPROXY_SOURCE_URL_QUERY_SEPARATOR = env.String("IMGPROXY_SOURCE_URL_QUERY_SEPARATOR")

	fsDesc = fs.ConfigDesc{
		Root: env.String("IMGPROXY_LOCAL_FILESYSTEM_ROOT"),
	}

	absConfigDesc = azure.ConfigDesc{
		Name:           env.String("IMGPROXY_ABS_NAME"),
		Endpoint:       env.String("IMGPROXY_ABS_ENDPOINT"),
		Key:            env.String("IMGPROXY_ABS_KEY"),
		AllowedBuckets: env.StringSlice("IMGPROXY_ABS_ALLOWED_BUCKETS"),
		DeniedBuckets:  env.StringSlice("IMGPROXY_ABS_DENIED_BUCKETS"),
	}

	gcsConfigDesc = gcs.ConfigDesc{
		Key:            env.String("IMGPROXY_GCS_KEY"),
		Endpoint:       env.String("IMGPROXY_GCS_ENDPOINT"),
		AllowedBuckets: env.StringSlice("IMGPROXY_GCS_ALLOWED_BUCKETS"),
		DeniedBuckets:  env.StringSlice("IMGPROXY_GCS_DENIED_BUCKETS"),
	}

	s3ConfigDesc = s3.ConfigDesc{
		Region:                  env.String("IMGPROXY_S3_REGION"),
		Endpoint:                env.String("IMGPROXY_S3_ENDPOINT"),
		EndpointUsePathStyle:    env.Bool("IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE"),
		AssumeRoleArn:           env.String("IMGPROXY_S3_ASSUME_ROLE_ARN"),
		AssumeRoleExternalID:    env.String("IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID"),
		DecryptionClientEnabled: env.Bool("IMGPROXY_S3_DECRYPTION_CLIENT_ENABLED"),
		AllowedBuckets:          env.StringSlice("IMGPROXY_S3_ALLOWED_BUCKETS"),
		DeniedBuckets:           env.StringSlice("IMGPROXY_S3_DENIED_BUCKETS"),
	}

	swiftConfigDesc = swift.ConfigDesc{
		Username:       env.String("IMGPROXY_SWIFT_USERNAME"),
		APIKey:         env.String("IMGPROXY_SWIFT_API_KEY"),
		AuthURL:        env.String("IMGPROXY_SWIFT_AUTH_URL"),
		Domain:         env.String("IMGPROXY_SWIFT_DOMAIN"),
		Tenant:         env.String("IMGPROXY_SWIFT_TENANT"),
		AuthVersion:    env.Int("IMGPROXY_SWIFT_AUTH_VERSION"),
		ConnectTimeout: env.Duration("IMGPROXY_SWIFT_CONNECT_TIMEOUT_SECONDS"),
		Timeout:        env.Duration("IMGPROXY_SWIFT_TIMEOUT_SECONDS"),
		AllowedBuckets: env.StringSlice("IMGPROXY_SWIFT_ALLOWED_BUCKETS"),
		DeniedBuckets:  env.StringSlice("IMGPROXY_SWIFT_DENIED_BUCKETS"),
	}
)

// Config represents configuration of the transport package
type Config struct {
	HTTP generichttp.Config

	Local fs.Config

	ABSEnabled bool
	ABS        azure.Config

	GCSEnabled bool
	GCS        gcs.Config

	S3Enabled bool
	S3        s3.Config

	SwiftEnabled bool
	Swift        swift.Config

	// query string separator (see docs). Unfortunately, we'll have to pass this
	// to each transport which needs it as the consturctor parameter. Otherwise,
	// we would have to add it to each transport config struct.
	SourceURLQuerySeparator string
}

// NewDefaultConfig returns a new default transport configuration
func NewDefaultConfig() Config {
	return Config{
		HTTP:                    generichttp.NewDefaultConfig(),
		Local:                   fs.NewDefaultConfig(),
		ABSEnabled:              false,
		ABS:                     azure.NewDefaultConfig(),
		GCSEnabled:              false,
		GCS:                     gcs.NewDefaultConfig(),
		S3Enabled:               false,
		S3:                      s3.NewDefaultConfig(),
		SwiftEnabled:            false,
		Swift:                   swift.NewDefaultConfig(),
		SourceURLQuerySeparator: "?", // default is ?, but can be overridden with empty
	}
}

// LoadConfigFromEnv loads transport configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	_, genericErr := generichttp.LoadConfigFromEnv(&c.HTTP)
	_, localErr := fs.LoadConfigFromEnv(fsDesc, &c.Local)
	_, absErr := azure.LoadConfigFromEnv(absConfigDesc, &c.ABS)
	_, gcsErr := gcs.LoadConfigFromEnv(gcsConfigDesc, &c.GCS)
	_, s3Err := s3.LoadConfigFromEnv(s3ConfigDesc, &c.S3)
	_, swiftErr := swift.LoadConfigFromEnv(swiftConfigDesc, &c.Swift)

	err := errors.Join(
		IMGPROXY_USE_ABS.Parse(&c.ABSEnabled),
		IMGPROXY_USE_GCS.Parse(&c.GCSEnabled),
		IMGPROXY_USE_S3.Parse(&c.S3Enabled),
		IMGPROXY_USE_SWIFT.Parse(&c.SwiftEnabled),
		genericErr,
		localErr,
		absErr,
		gcsErr,
		s3Err,
		swiftErr,
	)

	// empty value is a valid value for this separator, we can't rely on env.String,
	// which skips empty values
	if s, ok := os.LookupEnv(IMGPROXY_SOURCE_URL_QUERY_SEPARATOR.Name); ok {
		c.SourceURLQuerySeparator = s
	}

	return c, err
}

func (c *Config) Validate() error {
	return nil
}
