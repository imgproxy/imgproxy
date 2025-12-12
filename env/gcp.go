package env

import (
	"context"
	"errors"
	"fmt"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
)

var (
	IMGPROXY_ENV_GCP_SECRET_ID         = String("IMGPROXY_ENV_GCP_SECRET_ID")
	IMGPROXY_ENV_GCP_SECRET_VERSION_ID = String("IMGPROXY_ENV_GCP_SECRET_VERSION_ID")
	IMGPROXY_ENV_GCP_SECRET_PROJECT_ID = String("IMGPROXY_ENV_GCP_SECRET_PROJECT_ID")
	IMGPROXY_ENV_GCP_KEY               = String("IMGPROXY_ENV_GCP_KEY")
)

func loadGCPSecret(ctx context.Context) error {
	var secretID, secretVersion, secretProject, secretKey string

	IMGPROXY_ENV_GCP_SECRET_ID.Parse(&secretID)
	IMGPROXY_ENV_GCP_SECRET_VERSION_ID.Parse(&secretVersion)
	IMGPROXY_ENV_GCP_SECRET_PROJECT_ID.Parse(&secretProject)
	IMGPROXY_ENV_GCP_KEY.Parse(&secretKey)

	if len(secretID) == 0 {
		return nil
	}

	if len(secretVersion) == 0 {
		secretVersion = "latest"
	}

	var (
		client *secretmanager.Client
		err    error
	)

	ctx, ctxcancel := context.WithTimeout(ctx, time.Minute)
	defer ctxcancel()

	opts := []option.ClientOption{}

	if len(secretKey) > 0 {
		opts = append(opts, option.WithCredentialsJSON([]byte(secretKey)))
	}

	client, err = secretmanager.NewClient(ctx, opts...)

	if err != nil {
		return fmt.Errorf("can't create Google Cloud Secret Manager client: %w", err)
	}

	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", secretProject, secretID, secretVersion),
	}

	resp, err := client.AccessSecretVersion(ctx, &req)
	if err != nil {
		return fmt.Errorf("can't get Google Cloud Secret Manager secret: %w", err)
	}

	payload := resp.GetPayload()
	if payload == nil {
		return errors.New("can't get Google Cloud Secret Manager secret: payload is empty")
	}

	data := payload.GetData()

	if len(data) == 0 {
		return nil
	}

	return unmarshalEnv(string(data), "GCP Secret Manager")
}
