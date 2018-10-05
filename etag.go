package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"sync"
)

var errNotModified = newError(304, "Not modified", "Not modified")

type eTagCalc struct {
	hash hash.Hash
	enc  *json.Encoder
}

var eTagCalcPool = sync.Pool{
	New: func() interface{} {
		h := sha256.New()

		enc := json.NewEncoder(h)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "")

		return &eTagCalc{h, enc}
	},
}

func calcETag(ctx context.Context) []byte {
	c := eTagCalcPool.Get().(*eTagCalc)
	defer eTagCalcPool.Put(c)

	c.hash.Reset()
	c.hash.Write(getImageData(ctx).Bytes())
	footprint := c.hash.Sum(nil)

	c.hash.Reset()
	c.hash.Write(footprint)
	c.hash.Write([]byte(version))
	c.enc.Encode(conf)
	c.enc.Encode(getProcessingOptions(ctx))

	etag := make([]byte, 64)
	hex.Encode(etag, c.hash.Sum(nil))

	return etag
}
