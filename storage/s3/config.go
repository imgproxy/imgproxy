package s3

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions for S3 storage
type ConfigDesc struct {
	Region                  env.Desc
	Endpoint                env.Desc
	EndpointUsePathStyle    env.Desc
	AssumeRoleArn           env.Desc
	AssumeRoleExternalID    env.Desc
	DecryptionClientEnabled env.Desc
	AllowedBuckets          env.Desc
	DeniedBuckets           env.Desc
}

// Config holds the configuration for S3 transport
type Config struct {
	Region                  string   // AWS region for S3 (default: "")
	Endpoint                string   // Custom endpoint for S3 (default: "")
	EndpointUsePathStyle    bool     // Use path-style URLs for S3 (default: true)
	AssumeRoleArn           string   // ARN for assuming an AWS role (default: "")
	AssumeRoleExternalID    string   // External ID for assuming an AWS role (default: "")
	DecryptionClientEnabled bool     // Enables S3 decryption client (default: false)
	ReadOnly                bool     // Read-only access
	AllowedBuckets          []string // List of allowed buckets (containers)
	DeniedBuckets           []string // List of denied buckets (containers)
	desc                    ConfigDesc
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
		ReadOnly:                true,
		AllowedBuckets:          nil,
		DeniedBuckets:           nil,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Region, desc.Region),
		env.String(&c.Endpoint, desc.Endpoint),
		env.Bool(&c.EndpointUsePathStyle, desc.EndpointUsePathStyle),
		env.String(&c.AssumeRoleArn, desc.AssumeRoleArn),
		env.String(&c.AssumeRoleExternalID, desc.AssumeRoleExternalID),
		env.Bool(&c.DecryptionClientEnabled, desc.DecryptionClientEnabled),
		env.StringSlice(&c.AllowedBuckets, desc.AllowedBuckets),
		env.StringSlice(&c.DeniedBuckets, desc.DeniedBuckets),
	)

	c.desc = desc

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
