package env

import (
	"context"
	"fmt"
	"os"

	"github.com/DarthSim/godotenv"
)

var (
	IMGPROXY_ENV_LOCAL_FILE_PATH = Describe("IMGPROXY_ENV_LOCAL_FILE_PATH", "path")
)

// Load loads environment variables from various sources
func Load(ctx context.Context) error {
	if err := loadAWSSecret(ctx); err != nil {
		return err
	}

	if err := loadAWSSystemManagerParams(ctx); err != nil {
		return err
	}

	if err := loadGCPSecret(ctx); err != nil {
		return err
	}

	if err := loadLocalFile(); err != nil {
		return err
	}

	return nil
}

// loadLocalFile loads environment variables from a local file if IMGPROXY_ENV_LOCAL_FILE_PATH is set
func loadLocalFile() error {
	var path string

	String(&path, IMGPROXY_ENV_LOCAL_FILE_PATH)

	if len(path) == 0 {
		return nil
	}

	// Read the local environment file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read local environment file: %s", err)
	}

	// If the file is empty, nothing to load
	if len(data) == 0 {
		return nil
	}

	return unmarshalEnv(string(data), "local file")
}

// unmarshalEnv loads environment variables from a string to process environment
func unmarshalEnv(env, source string) error {
	// Parse the secret string as env variables and set them
	envmap, err := godotenv.Unmarshal(env)
	if err != nil {
		return fmt.Errorf("can't parse config from %s: %s", source, err)
	}

	for k, v := range envmap {
		if err = os.Setenv(k, v); err != nil {
			return fmt.Errorf("can't set %s env variable from %s: %s", k, source, err)
		}
	}

	return nil
}
