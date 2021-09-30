package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/imgproxy/imgproxy/v3/config"
)

var (
	ErrInvalidSignature         = errors.New("Invalid signature")
	ErrInvalidSignatureEncoding = errors.New("Invalid signature encoding")
)

func VerifySignature(signature, path string) error {
	if len(config.Keys) == 0 || len(config.Salts) == 0 {
		return nil
	}

	messageMAC, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return ErrInvalidSignatureEncoding
	}

	for i := 0; i < len(config.Keys); i++ {
		if hmac.Equal(messageMAC, signatureFor(path, config.Keys[i], config.Salts[i], config.SignatureSize)) {
			return nil
		}
	}

	return ErrInvalidSignature
}

func signatureFor(str string, key, salt []byte, signatureSize int) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	mac.Write([]byte(str))
	expectedMAC := mac.Sum(nil)
	if signatureSize < 32 {
		return expectedMAC[:signatureSize]
	}
	return expectedMAC
}
