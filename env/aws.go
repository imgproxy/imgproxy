package env

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	// defaultAWSRegion represents default AWS region for all configuration
	defaultAWSRegion = "us-west-1"
)

var (
	IMGPROXY_ENV_AWS_SECRET_ID             = String("IMGPROXY_ENV_AWS_SECRET_ID")
	IMGPROXY_ENV_AWS_SECRET_VERSION_ID     = String("IMGPROXY_ENV_AWS_SECRET_VERSION_ID")
	IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE  = String("IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE")
	IMGPROXY_ENV_AWS_SECRET_REGION         = String("IMGPROXY_ENV_AWS_SECRET_REGION")
	IMGPROXY_ENV_AWS_SSM_PARAMETERS_PATH   = String("IMGPROXY_ENV_AWS_SSM_PARAMETERS_PATH")
	IMGPROXY_ENV_AWS_SSM_PARAMETERS_REGION = String("IMGPROXY_ENV_AWS_SSM_PARAMETERS_REGION")
)

func loadAWSSecret(ctx context.Context) error {
	var secretID, secretVersionID, secretVersionStage, secretRegion string

	IMGPROXY_ENV_AWS_SECRET_ID.Parse(&secretID)
	IMGPROXY_ENV_AWS_SECRET_VERSION_ID.Parse(&secretVersionID)
	IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE.Parse(&secretVersionStage)
	IMGPROXY_ENV_AWS_SECRET_REGION.Parse(&secretRegion)

	// No secret ID, no aws
	if len(secretID) == 0 {
		return nil
	}

	// Let's form AWS default config
	conf, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("can't load AWS Secrets Manager config: %w", err)
	}

	if len(secretRegion) > 0 {
		conf.Region = secretRegion
	}

	if len(conf.Region) == 0 {
		conf.Region = defaultAWSRegion
	}

	// Let's create secrets manager client
	client := secretsmanager.NewFromConfig(conf)

	input := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretID)}
	if len(secretVersionID) > 0 {
		input.VersionId = aws.String(secretVersionID)
	} else if len(secretVersionStage) > 0 {
		input.VersionStage = aws.String(secretVersionStage)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := client.GetSecretValue(ctx, &input)
	if err != nil {
		return fmt.Errorf("can't retrieve config from AWS Secrets Manager: %w", err)
	}

	// No secret string, failed to initialize secrets manager, return
	if output.SecretString == nil {
		return nil
	}

	return unmarshalEnv(*output.SecretString, "AWS Secrets Manager")
}

// loadAWSSystemManagerParams loads environment variables from AWS System Manager
func loadAWSSystemManagerParams(ctx context.Context) error {
	var paramsPath, paramsRegion string

	IMGPROXY_ENV_AWS_SSM_PARAMETERS_PATH.Parse(&paramsPath)
	IMGPROXY_ENV_AWS_SSM_PARAMETERS_REGION.Parse(&paramsRegion)

	// Path is not set: can't use SSM
	if len(paramsPath) == 0 {
		return nil
	}

	conf, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("can't load AWS SSM config: %w", err)
	}

	conf.Region = defaultAWSRegion
	if len(paramsRegion) != 0 {
		conf.Region = paramsRegion
	}

	// Let's create SSM client
	client := ssm.NewFromConfig(conf)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var nextToken *string

	for {
		input := ssm.GetParametersByPathInput{
			Path:           aws.String(paramsPath),
			WithDecryption: aws.Bool(true),
			NextToken:      nextToken,
		}

		output, err := client.GetParametersByPath(ctx, &input)
		if err != nil {
			return fmt.Errorf("can't retrieve parameters from AWS SSM: %w", err)
		}

		for _, p := range output.Parameters {
			if p.Name == nil || p.Value == nil {
				continue
			}

			if p.DataType == nil || *p.DataType != "text" {
				continue
			}

			name := *p.Name

			env := strings.ReplaceAll(
				strings.TrimPrefix(strings.TrimPrefix(name, paramsPath), "/"),
				"/", "_",
			)

			if err = os.Setenv(env, *p.Value); err != nil {
				return fmt.Errorf("can't set %s env variable from AWS SSM: %w", env, err)
			}
		}

		if nextToken = output.NextToken; nextToken == nil {
			break
		}
	}

	return nil
}
