package s3

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_S3_REGION                    = env.Describe("IMGPROXY_S3_REGION", "string")
	IMGPROXY_S3_ENDPOINT                  = env.Describe("IMGPROXY_S3_ENDPOINT", "string")
	IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE   = env.Describe("IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE", "boolean")
	IMGPROXY_S3_ASSUME_ROLE_ARN           = env.Describe("IMGPROXY_S3_ASSUME_ROLE_ARN", "string")
	IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID   = env.Describe("IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID", "string")
	IMGPROXY_S3_DECRYPTION_CLIENT_ENABLED = env.Describe("IMGPROXY_S3_DECRYPTION_CLIENT_ENABLED", "boolean")
)

// Config holds the configuration for S3 transport
type Config struct {
	Region                  string // AWS region for S3 (default: "")
	Endpoint                string // Custom endpoint for S3 (default: "")
	EndpointUsePathStyle    bool   // Use path-style URLs for S3 (default: true)
	AssumeRoleArn           string // ARN for assuming an AWS role (default: "")
	AssumeRoleExternalID    string // External ID for assuming an AWS role (default: "")
	DecryptionClientEnabled bool   // Enables S3 decryption client (default: false)
}

// NewDefaultConfig returns a new default configuration for S3 transport
func NewDefaultConfig() Config {
	return Config{
		Region:                  "",
		Endpoint:                "",
		EndpointUsePathStyle:    true,
		AssumeRoleArn:           "",
		AssumeRoleExternalID:    "",
		DecryptionClientEnabled: false,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Region, IMGPROXY_S3_REGION),
		env.String(&c.Endpoint, IMGPROXY_S3_ENDPOINT),
		env.Bool(&c.EndpointUsePathStyle, IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE),
		env.String(&c.AssumeRoleArn, IMGPROXY_S3_ASSUME_ROLE_ARN),
		env.String(&c.AssumeRoleExternalID, IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID),
		env.Bool(&c.DecryptionClientEnabled, IMGPROXY_S3_DECRYPTION_CLIENT_ENABLED),
	)

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
