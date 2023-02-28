package loadenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/DarthSim/godotenv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func loadAWSSecret() error {
	secretID := os.Getenv("IMGPROXY_ENV_AWS_SECRET_ID")
	secretVersionID := os.Getenv("IMGPROXY_ENV_AWS_SECRET_VERSION_ID")
	secretVersionStage := os.Getenv("IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE")
	secretRegion := os.Getenv("IMGPROXY_ENV_AWS_SECRET_REGION")

	if len(secretID) == 0 {
		return nil
	}

	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("Can't create AWS Secrets Manager session: %s", err)
	}

	conf := aws.NewConfig()

	if len(secretRegion) != 0 {
		conf.Region = aws.String(secretRegion)
	}

	svc := secretsmanager.New(sess, conf)

	input := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretID)}
	if len(secretVersionID) > 0 {
		input.VersionId = aws.String(secretVersionID)
	} else if len(secretVersionStage) > 0 {
		input.VersionStage = aws.String(secretVersionStage)
	}

	output, err := svc.GetSecretValue(&input)
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

	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("Can't create AWS SSM session: %s", err)
	}

	conf := aws.NewConfig()

	if len(paramsRegion) != 0 {
		conf.Region = aws.String(paramsRegion)
	}

	svc := ssm.New(sess, conf)

	input := ssm.GetParametersByPathInput{
		Path:           aws.String(paramsPath),
		WithDecryption: aws.Bool(true),
	}

	output, err := svc.GetParametersByPath(&input)
	if err != nil {
		return fmt.Errorf("Can't retrieve parameters from AWS SSM: %s", err)
	}

	for _, p := range output.Parameters {
		if p == nil || p.Name == nil || p.Value == nil {
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

	return nil
}
