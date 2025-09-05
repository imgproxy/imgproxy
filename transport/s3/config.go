package s3

import "github.com/imgproxy/imgproxy/v3/config"

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
func NewDefaultConfig() *Config {
	return &Config{
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
	c.Region = config.S3Region
	c.Endpoint = config.S3Endpoint
	c.EndpointUsePathStyle = config.S3EndpointUsePathStyle
	c.AssumeRoleArn = config.S3AssumeRoleArn
	c.AssumeRoleExternalID = config.S3AssumeRoleExternalID
	c.DecryptionClientEnabled = config.S3DecryptionClientEnabled

	return c, nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
