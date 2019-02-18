package internal

import (
	"encoding/base64"
	"errors"
)

func deobfuscate(in string, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be zero length")
	}

	decoded, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(decoded))
	for i, c := range decoded {
		out[i] = c ^ key[i%len(key)]
	}

	return out, nil
}

func obfuscate(in, key []byte) (string, error) {
	if len(key) == 0 {
		return "", errors.New("key cannot be zero length")
	}

	out := make([]byte, len(in))
	for i, c := range in {
		out[i] = c ^ key[i%len(key)]
	}

	return base64.StdEncoding.EncodeToString(out), nil
}
