package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	key := "943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881"
	salt := "520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5"

	var keyBin, saltBin []byte
	var err error

	if keyBin, err = hex.DecodeString(key); err != nil {
		log.Fatal("Key expected to be hex-encoded string")
	}

	if saltBin, err = hex.DecodeString(salt); err != nil {
		log.Fatal("Salt expected to be hex-encoded string")
	}

	resize := "fill"
	width := 300
	height := 300
	gravity := "no"
	enlarge := 1
	extension := "png"

	url := "http://img.example.com/pretty/image.jpg"
	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(url))

	path := fmt.Sprintf("/rs:%s:%d:%d:%d/g:%s/%s.%s", resize, width, height, enlarge, gravity, encodedURL, extension)

	mac := hmac.New(sha256.New, keyBin)
	mac.Write(saltBin)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	fmt.Printf("/%s%s", signature, path)
}
