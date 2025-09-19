package env

import (
	"strconv"
	"time"
)

// String parses an environment variable as a string
func String(value string) (string, error) {
	return value, nil
}

// Int parses an environment variable as an integer
func Int(value string) (int, error) {
	return strconv.Atoi(value)
}

// Duration parses an environment variable as a time.Duration in seconds
func DurationSec(value string) (time.Duration, error) {
	sec, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return time.Duration(sec) * time.Second, nil
}

// func Float(i *float64, name string) {
// 	if env, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
// 		*i = env
// 	}
// }

// func MegaInt(f *int, name string) {
// 	if env, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
// 		*f = int(env * 1000000)
// 	}
// }

// func String(s *string, name string) {
// 	if env := os.Getenv(name); len(env) > 0 {
// 		*s = env
// 	}
// }

// func StringSliceSep(s *[]string, name, sep string) {
// 	if env := os.Getenv(name); len(env) > 0 {
// 		parts := strings.Split(env, sep)

// 		for i, p := range parts {
// 			parts[i] = strings.TrimSpace(p)
// 		}

// 		*s = parts

// 		return
// 	}

// 	*s = []string{}
// }
