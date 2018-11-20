package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var (
	errInvalidSignature         = errors.New("Invalid signature")
	errInvalidSignatureEncoding = errors.New("Invalid signature encoding")
)

type securityKey []byte

func validatePath(signature, path string) error {
	messageMAC, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return errInvalidSignatureEncoding
	}

	for i := 0; i < len(conf.Keys); i++ {
		if hmac.Equal(messageMAC, signatureFor(path, i)) {
			return nil
		}
	}

	return errInvalidSignature
}

func signatureFor(str string, pairInd int) []byte {
	mac := hmac.New(sha256.New, conf.Keys[pairInd])
	mac.Write(conf.Salts[pairInd])
	mac.Write([]byte(str))
	expectedMAC := mac.Sum(nil)
	if conf.SignatureSize < 32 {
		return expectedMAC[:conf.SignatureSize]
	}
	return expectedMAC
}
