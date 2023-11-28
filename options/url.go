package options

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
)

const urlTokenPlain = "plain"
const urlTokenEnc = "enc"

func preprocessURL(u string) string {
	for _, repl := range config.URLReplacements {
		u = repl.Regexp.ReplaceAllString(u, repl.Replacement)
	}

	if len(config.BaseURL) == 0 || strings.HasPrefix(u, config.BaseURL) {
		return u
	}

	return fmt.Sprintf("%s%s", config.BaseURL, u)
}

func decodeBase64URL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "")
	urlParts := strings.Split(encoded, ".")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	imageURL, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(urlParts[0], "="))
	if err != nil {
		return "", "", fmt.Errorf("Invalid url encoding: %s", encoded)
	}

	return preprocessURL(string(imageURL)), format, nil
}

func decodeEncURL(parts []string) (string, string, error) {
	var format string
	var err error

	encoded := strings.Join(parts, "/")
	urlParts := strings.Split(encoded, ".")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	ciphertext, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(urlParts[0], "="))
	if err != nil {
		return "", "", err
	}

	var keyBin []byte
	if keyBin, err = hex.DecodeString(config.EncryptionKey); err != nil {
		return "", "", err
	}

	if len(keyBin) == 0 {
		return "", "", fmt.Errorf("Encrypted url provided but no encryption key set. Set IMGPROXY_SOURCE_URL_ENCRYPTION_KEY")
	}

	c, err := aes.NewCipher(keyBin)
	if err != nil {
		return "", "", err
	}

	decrypted := make([]byte, len(ciphertext[aes.BlockSize:]))
	iv := ciphertext[:aes.BlockSize]
	encrypted := ciphertext[aes.BlockSize:]
	mode := cipher.NewCBCDecrypter(c, iv)
	mode.CryptBlocks(decrypted, encrypted)

	decrypted = bytes.ReplaceAll(decrypted, []byte{3}, []byte{})

	return string(decrypted), format, nil
}

func decodePlainURL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "/")
	urlParts := strings.Split(encoded, "@")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	unescaped, err := url.PathUnescape(urlParts[0])
	if err != nil {
		return "", "", fmt.Errorf("Invalid url encoding: %s", encoded)
	}

	return preprocessURL(unescaped), format, nil
}

func DecodeURL(parts []string) (string, string, error) {
	if len(parts) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if parts[0] == urlTokenEnc && len(parts) > 1 {
		return decodeEncURL(parts[1:])
	}

	if parts[0] == urlTokenPlain && len(parts) > 1 {
		return decodePlainURL(parts[1:])
	}

	return decodeBase64URL(parts)
}
