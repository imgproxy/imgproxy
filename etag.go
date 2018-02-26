package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
)

var notModifiedErr = newError(304, "Not modified", "Not modified")

func calcETag(b []byte, po *processingOptions) string {
	footprint := sha1.Sum(b)

	hash := sha1.New()
	hash.Write(footprint[:])
	binary.Write(hash, binary.LittleEndian, *po)
	hash.Write(conf.ETagSignature)

	return fmt.Sprintf("%x", hash.Sum(nil))
}
