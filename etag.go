package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

var notModifiedErr = newError(304, "Not modified", "Not modified")

func calcETag(ctx context.Context) []byte {
	footprint := sha256.Sum256(getImageData(ctx).Bytes())

	hash := sha256.New()
	hash.Write(footprint[:])
	hash.Write([]byte(version))
	binary.Write(hash, binary.LittleEndian, conf)
	binary.Write(hash, binary.LittleEndian, *getProcessingOptions(ctx))

	return []byte(fmt.Sprintf("%x", hash.Sum(nil)))
}
