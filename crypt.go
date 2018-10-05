package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var (
	errInvalidToken         = errors.New("Invalid token")
	errInvalidTokenEncoding = errors.New("Invalid token encoding")
)

func validatePath(token, path string) error {
	messageMAC, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return errInvalidTokenEncoding
	}

	mac := hmac.New(sha256.New, conf.Key)
	mac.Write(conf.Salt)
	mac.Write([]byte(path))
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(messageMAC, expectedMAC) {
		return errInvalidToken
	}

	return nil
}
