package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"sync"

	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/version"
)

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

func calcETag(ctx context.Context, imgdata *imagedata.ImageData, po *options.ProcessingOptions) string {
	c := eTagCalcPool.Get().(*eTagCalc)
	defer eTagCalcPool.Put(c)

	c.hash.Reset()
	c.hash.Write(imgdata.Data)
	footprint := c.hash.Sum(nil)

	c.hash.Reset()
	c.hash.Write(footprint)
	c.hash.Write([]byte(version.Version()))
	c.enc.Encode(po)

	return hex.EncodeToString(c.hash.Sum(nil))
}
