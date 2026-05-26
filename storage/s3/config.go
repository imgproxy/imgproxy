package s3

import (
	"errors"

	"github.com/imgproxy/imgproxy/v4/ensure"
	"github.com/imgproxy/imgproxy/v4/env"
)

// ConfigDesc holds the configuration descriptions for S3 storage
type ConfigDesc struct {
	Region                  env.StringVar
	Endpoint                env.StringVar
	EndpointUsePathStyle    env.BoolVar
	AssumeRoleArn           env.StringVar
	AssumeRoleExternalID    env.StringVar
	DecryptionClientEnabled env.BoolVar
	AllowedBuckets          env.StringSliceVar
	DeniedBuckets           env.StringSliceVar
	AccessPoints            env.StringMapVar
}

// Config holds the configuration for S3 transport
type Config struct {
	Region                  string            // AWS region for S3 (default: "")
	Endpoint                string            // Custom endpoint for S3 (default: "")
	EndpointUsePathStyle    bool              // Use path-style URLs for S3 (default: true)
	AssumeRoleArn           string            // ARN for assuming an AWS role (default: "")
	AssumeRoleExternalID    string            // External ID for assuming an AWS role (default: "")
	DecryptionClientEnabled bool              // Enables S3 decryption client (default: false)
	AllowedBuckets          []string          // List of allowed buckets (containers)
	DeniedBuckets           []string          // List of denied buckets (containers)
	AccessPoints            map[string]string // Map of bucket names to access point ARNs
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
		AllowedBuckets:          nil,
		DeniedBuckets:           nil,
		AccessPoints:            nil,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		desc.Region.Parse(&c.Region),
		desc.Endpoint.Parse(&c.Endpoint),
		desc.EndpointUsePathStyle.Parse(&c.EndpointUsePathStyle),
		desc.AssumeRoleArn.Parse(&c.AssumeRoleArn),
		desc.AssumeRoleExternalID.Parse(&c.AssumeRoleExternalID),
		desc.DecryptionClientEnabled.Parse(&c.DecryptionClientEnabled),
		desc.AllowedBuckets.Parse(&c.AllowedBuckets),
		desc.DeniedBuckets.Parse(&c.DeniedBuckets),
		desc.AccessPoints.Parse(&c.AccessPoints),
	)

	c.desc = desc

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	for name, arn := range c.AccessPoints {
		if arn == "" {
			return c.desc.AccessPoints.Errorf("access point %q ARN is empty", name)
		}
	}

	return nil
}
