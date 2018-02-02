package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net/http"
)

// check whether client's ETag matches current response body.
// - if the IMGPROXY_USE_ETAG env var is unset, this function always returns false
// - if the IMGPROXY_USE_ETAG is set, the function calculates current ETag and compare it
//   with another ETag value provided by client
// Note that the calculated ETag is saved to outcoming response with "ETag" header.
func isETagMatching(b []byte, po *processingOptions, rw *http.ResponseWriter, r *http.Request) bool {

	if conf.ETagEnabled {

		// calculate current ETag value using sha1 hashing function
		currentEtagValue := calculateHashSumFor(b, po)
		(*rw).Header().Set("ETag", currentEtagValue)
		return currentEtagValue == r.Header.Get("If-None-Match")
	}

	return false
}

// function calculates the SHA checksum for the current image and current Processing Options.
// Principal is very simple: if an original image is the same and PO are the same, then
// the checksum must be always identical. But if PO has some different parameters, the
// checksum must be different event if the orinal image matches
func calculateHashSumFor(b []byte, po *processingOptions) string {
	hash := sha1.New()

	hash.Write(b)

	binary.Write(hash, binary.LittleEndian, po.enlarge)
	binary.Write(hash, binary.LittleEndian, po.format)
	binary.Write(hash, binary.LittleEndian, po.gravity)
	binary.Write(hash, binary.LittleEndian, po.height)
	binary.Write(hash, binary.LittleEndian, po.width)
	binary.Write(hash, binary.LittleEndian, po.resize)

	hash.Write([]byte(conf.Salt))

	return fmt.Sprintf("%x", hash.Sum(nil))
}
