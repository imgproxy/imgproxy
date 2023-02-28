package loadenv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/DarthSim/godotenv"
	"google.golang.org/api/option"
)

func loadGCPSecret() error {
	secretID := os.Getenv("IMGPROXY_ENV_GCP_SECRET_ID")
	secretVersion := os.Getenv("IMGPROXY_ENV_GCP_SECRET_VERSION_ID")
	secretProject := os.Getenv("IMGPROXY_ENV_GCP_SECRET_PROJECT_ID")
	secretKey := os.Getenv("IMGPROXY_ENV_GCP_KEY")

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

	ctx, ctxcancel := context.WithTimeout(context.Background(), time.Minute)
	defer ctxcancel()

	opts := []option.ClientOption{}

	if len(secretKey) > 0 {
		opts = append(opts, option.WithCredentialsJSON([]byte(secretKey)))
	}

	client, err = secretmanager.NewClient(ctx, opts...)

	if err != nil {
		return fmt.Errorf("Can't create Google Cloud Secret Manager client: %s", err)
	}

	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", secretProject, secretID, secretVersion),
	}

	resp, err := client.AccessSecretVersion(ctx, &req)
	if err != nil {
		return fmt.Errorf("Can't get Google Cloud Secret Manager secret: %s", err)
	}

	payload := resp.GetPayload()
	if payload == nil {
		return errors.New("Can't get Google Cloud Secret Manager secret: payload is empty")
	}

	data := payload.GetData()

	if len(data) == 0 {
		return nil
	}

	envmap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return fmt.Errorf("Can't parse config from Google Cloud Secrets Manager: %s", err)
	}

	for k, v := range envmap {
		if err = os.Setenv(k, v); err != nil {
			return fmt.Errorf("Can't set %s env variable from Google Cloud Secrets Manager: %s", k, err)
		}
	}

	return nil
}
