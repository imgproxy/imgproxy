package main

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatureFor(t *testing.T) {
	oldConf := conf
	defer func() { conf = oldConf }()

	base64Signature := func(x string) string { return base64.RawURLEncoding.EncodeToString(signatureFor(x)) }
	conf.Key = []byte("test-key")
	conf.Salt = []byte("test-salt")
	assert.Equal(t, "dtLwhdnPPiu_epMl1LrzheLpvHas-4mwvY6L3Z8WwlY", base64Signature("asd"))
	assert.Equal(t, "8x1xvzxVqZ3Uz3kEC8gVvBfU0dfU1vKv0Gho8m3Ysgw", base64Signature("qwe"))
	conf.SignatureSize = 8
	assert.Equal(t, "dtLwhdnPPis", base64Signature("asd"))
	assert.Equal(t, "8x1xvzxVqZ0", base64Signature("qwe"))
}
