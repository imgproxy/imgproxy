package loadenv

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DarthSim/godotenv"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func loadAWSSecret() error {
	secretID := os.Getenv("IMGPROXY_ENV_AWS_SECRET_ID")
	secretVersionID := os.Getenv("IMGPROXY_ENV_AWS_SECRET_VERSION_ID")
	secretVersionStage := os.Getenv("IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE")
	secretRegion := os.Getenv("IMGPROXY_ENV_AWS_SECRET_REGION")

	if len(secretID) == 0 {
		return nil
	}

	conf, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return fmt.Errorf("can't load AWS Secrets Manager config: %s", err)
	}

	if len(secretRegion) != 0 {
		conf.Region = secretRegion
	}

	if len(conf.Region) == 0 {
		conf.Region = "us-west-1"
	}

	client := secretsmanager.NewFromConfig(conf)

	input := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretID)}
	if len(secretVersionID) > 0 {
		input.VersionId = aws.String(secretVersionID)
	} else if len(secretVersionStage) > 0 {
		input.VersionStage = aws.String(secretVersionStage)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := client.GetSecretValue(ctx, &input)
	if err != nil {
		return fmt.Errorf("Can't retrieve config from AWS Secrets Manager: %s", err)
	}

	if output.SecretString == nil {
		return nil
	}

	envmap, err := godotenv.Unmarshal(*output.SecretString)
	if err != nil {
		return fmt.Errorf("Can't parse config from AWS Secrets Manager: %s", err)
	}

	for k, v := range envmap {
		if err = os.Setenv(k, v); err != nil {
			return fmt.Errorf("Can't set %s env variable from AWS Secrets Manager: %s", k, err)
		}
	}

	return nil
}

func loadAWSSystemManagerParams() error {
	paramsPath := os.Getenv("IMGPROXY_ENV_AWS_SSM_PARAMETERS_PATH")
	paramsRegion := os.Getenv("IMGPROXY_ENV_AWS_SSM_PARAMETERS_REGION")

	if len(paramsPath) == 0 {
		return nil
	}

	conf, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return fmt.Errorf("can't load AWS SSM config: %s", err)
	}

	if len(paramsRegion) != 0 {
		conf.Region = paramsRegion
	}

	if len(conf.Region) == 0 {
		conf.Region = "us-west-1"
	}

	client := ssm.NewFromConfig(conf)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			return fmt.Errorf("Can't retrieve parameters from AWS SSM: %s", err)
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
				return fmt.Errorf("Can't set %s env variable from AWS SSM: %s", env, err)
			}
		}

		if nextToken = output.NextToken; nextToken == nil {
			break
		}
	}

	return nil
}
