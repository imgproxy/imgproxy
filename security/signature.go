package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/imgproxy/imgproxy/v3/config"
)

func VerifySignature(signature, path string) error {
	if len(config.Keys) == 0 || len(config.Salts) == 0 {
		return nil
	}

	for _, s := range config.TrustedSignatures {
		if s == signature {
			return nil
		}
	}

	messageMAC, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return newSignatureError("Invalid signature encoding")
	}

	for i := 0; i < len(config.Keys); i++ {
		if hmac.Equal(messageMAC, signatureFor(path, config.Keys[i], config.Salts[i], config.SignatureSize)) {
			return nil
		}
	}

	return newSignatureError("Invalid signature")
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
