package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

var notModifiedErr = newError(304, "Not modified", "Not modified")

func calcETag(b []byte, po *processingOptions) string {
	footprint := sha256.Sum256(b)

	hash := sha256.New()
	hash.Write(footprint[:])
	hash.Write([]byte(version))
	binary.Write(hash, binary.LittleEndian, conf)
	binary.Write(hash, binary.LittleEndian, *po)

	return fmt.Sprintf("%x", hash.Sum(nil))
}
