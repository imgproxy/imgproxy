package examples

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
)

func pkcs7pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}

func ExcryptSourceURL() {
	key := "1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1"

	var (
		keyBin []byte
		err    error
	)

	if keyBin, err = hex.DecodeString(key); err != nil {
		log.Fatal("Key expected to be hex-encoded string")
	}

	url := "http://img.example.com/pretty/image.jpg"

	c, err := aes.NewCipher(keyBin)
	if err != nil {
		log.Fatal(err)
	}

	data := pkcs7pad([]byte(url), aes.BlockSize)

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]

	// We use a random iv generation, but you'll probably want to use some
	// deterministic method
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal(err)
	}

	mode := cipher.NewCBCEncrypter(c, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)

	encryptedURL := base64.RawURLEncoding.EncodeToString(ciphertext)

	// We don't sign the URL in this example but it is highly recommended to sign
	// imgproxy URLs when imgproxy is being used in production.
	// Signing URLs is especially important when using encrypted source URLs to
	// prevent a padding oracle attack
	fmt.Printf("/unsafe/rs:fit:300:300/enc/%s.jpg", encryptedURL)
}
