// config.go is just a shortcut for common.Config which helps to
// avoid importing of the `common` package directly.
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
	IMGPROXY_USE_ABS                    = env.Describe("IMGPROXY_USE_ABS", "boolean")
	IMGPROXY_USE_GCS                    = env.Describe("IMGPROXY_GCS_ENABLED", "boolean")
	IMGPROXY_USE_S3                     = env.Describe("IMGPROXY_USE_S3", "boolean")
	IMGPROXY_USE_SWIFT                  = env.Describe("IMGPROXY_USE_SWIFT", "boolean")
	IMGPROXY_SOURCE_URL_QUERY_SEPARATOR = env.Describe("IMGPROXY_SOURCE_URL_QUERY_SEPARATOR", "string")

	fsDesc = fs.ConfigDesc{
		Root: env.Describe("IMGPROXY_LOCAL_FILESYSTEM_ROOT", "path"),
	}

	absConfigDesc = azure.ConfigDesc{
		Name:           env.Describe("IMGPROXY_ABS_NAME", "string"),
		Endpoint:       env.Describe("IMGPROXY_ABS_ENDPOINT", "string"),
		Key:            env.Describe("IMGPROXY_ABS_KEY", "string"),
		AllowedBuckets: env.Describe("IMGPROXY_ABS_ALLOWED_BUCKETS", "comma-separated list"),
		DeniedBuckets:  env.Describe("IMGPROXY_ABS_DENIED_BUCKETS", "comma-separated list"),
	}

	gcsConfigDesc = gcs.ConfigDesc{
		Key:            env.Describe("IMGPROXY_GCS_KEY", "string"),
		Endpoint:       env.Describe("IMGPROXY_GCS_ENDPOINT", "string"),
		AllowedBuckets: env.Describe("IMGPROXY_GCS_ALLOWED_BUCKETS", "comma-separated list"),
		DeniedBuckets:  env.Describe("IMGPROXY_GCS_DENIED_BUCKETS", "comma-separated list"),
	}

	s3ConfigDesc = s3.ConfigDesc{
		Region:                  env.Describe("IMGPROXY_S3_REGION", "string"),
		Endpoint:                env.Describe("IMGPROXY_S3_ENDPOINT", "string"),
		EndpointUsePathStyle:    env.Describe("IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE", "boolean"),
		AssumeRoleArn:           env.Describe("IMGPROXY_S3_ASSUME_ROLE_ARN", "string"),
		AssumeRoleExternalID:    env.Describe("IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID", "string"),
		DecryptionClientEnabled: env.Describe("IMGPROXY_S3_DECRYPTION_CLIENT_ENABLED", "boolean"),
		AllowedBuckets:          env.Describe("IMGPROXY_S3_ALLOWED_BUCKETS", "comma-separated list"),
		DeniedBuckets:           env.Describe("IMGPROXY_S3_DENIED_BUCKETS", "comma-separated list"),
	}

	swiftConfigDesc = swift.ConfigDesc{
		Username:       env.Describe("IMGPROXY_SWIFT_USERNAME", "string"),
		APIKey:         env.Describe("IMGPROXY_SWIFT_API_KEY", "string"),
		AuthURL:        env.Describe("IMGPROXY_SWIFT_AUTH_URL", "string"),
		Domain:         env.Describe("IMGPROXY_SWIFT_DOMAIN", "string"),
		Tenant:         env.Describe("IMGPROXY_SWIFT_TENANT", "string"),
		AuthVersion:    env.Describe("IMGPROXY_SWIFT_AUTH_VERSION", "number"),
		ConnectTimeout: env.Describe("IMGPROXY_SWIFT_CONNECT_TIMEOUT_SECONDS", "number"),
		Timeout:        env.Describe("IMGPROXY_SWIFT_TIMEOUT_SECONDS", "number"),
		AllowedBuckets: env.Describe("IMGPROXY_SWIFT_ALLOWED_BUCKETS", "comma-separated list"),
		DeniedBuckets:  env.Describe("IMGPROXY_SWIFT_DENIED_BUCKETS", "comma-separated list"),
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
		SourceURLQuerySeparator: "?", // default is ?, but can be overriden with empty
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
		env.Bool(&c.ABSEnabled, IMGPROXY_USE_ABS),
		env.Bool(&c.GCSEnabled, IMGPROXY_USE_GCS),
		env.Bool(&c.S3Enabled, IMGPROXY_USE_S3),
		env.Bool(&c.SwiftEnabled, IMGPROXY_USE_SWIFT),
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
