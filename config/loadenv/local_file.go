package loadenv

import (
	"fmt"
	"os"

	"github.com/DarthSim/godotenv"
)

func loadLocalFile() error {
	path := os.Getenv("IMGPROXY_ENV_LOCAL_FILE_PATH")

	if len(path) == 0 {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Can't read loacal environment file: %s", err)
	}

	if len(data) == 0 {
		return nil
	}

	envmap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return fmt.Errorf("Can't parse config from local file: %s", err)
	}

	for k, v := range envmap {
		if err = os.Setenv(k, v); err != nil {
			return fmt.Errorf("Can't set %s env variable from local file: %s", k, err)
		}
	}

	return nil
}
