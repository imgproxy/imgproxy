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
	IMGPROXY_ENV_GCP_SECRET_ID         = Describe("IMGPROXY_ENV_GCP_SECRET_ID", "string")
	IMGPROXY_ENV_GCP_SECRET_VERSION_ID = Describe("IMGPROXY_ENV_GCP_SECRET_VERSION_ID", "string")
	IMGPROXY_ENV_GCP_SECRET_PROJECT_ID = Describe("IMGPROXY_ENV_GCP_SECRET_PROJECT_ID", "string")
	IMGPROXY_ENV_GCP_KEY               = Describe("IMGPROXY_ENV_GCP_KEY", "JSON string")
)

func loadGCPSecret(ctx context.Context) error {
	var secretID, secretVersion, secretProject, secretKey string

	String(&secretID, IMGPROXY_ENV_GCP_SECRET_ID)
	String(&secretVersion, IMGPROXY_ENV_GCP_SECRET_VERSION_ID)
	String(&secretProject, IMGPROXY_ENV_GCP_SECRET_PROJECT_ID)
	String(&secretKey, IMGPROXY_ENV_GCP_KEY)

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
		return fmt.Errorf("can't create Google Cloud Secret Manager client: %s", err)
	}

	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", secretProject, secretID, secretVersion),
	}

	resp, err := client.AccessSecretVersion(ctx, &req)
	if err != nil {
		return fmt.Errorf("can't get Google Cloud Secret Manager secret: %s", err)
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
