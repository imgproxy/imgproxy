package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net/http"
)

// checks whether client's ETag matches current response body.
// - if the IMGPROXY_USE_ETAG env var is unset, this function always returns false
// - if the IMGPROXY_USE_ETAG is set to "true", the function calculates current ETag and compares it
//   with another ETag value provided by a client request
// Note that the calculated ETag value is saved to outcoming response with "ETag" header.
func isETagMatching(b []byte, po *processingOptions, rw *http.ResponseWriter, r *http.Request) bool {

	if !conf.ETagEnabled {
		return false
	}

	// calculate current ETag value using sha1 hashing function
	currentEtagValue := calculateHashSumFor(b, po)
	(*rw).Header().Set("ETag", currentEtagValue)
	return currentEtagValue == r.Header.Get("If-None-Match")
}

// the function calculates the SHA checksum for the current image and current Processing Options.
// The principal is very simple: if an original image is the same and POs are the same, then
// the checksum must be always identical. But if PO has some different parameters, the
// checksum must be different even if original images match
func calculateHashSumFor(b []byte, po *processingOptions) string {
	hash := sha1.New()

	// SHA checksum consists of image, porcessing options and random value, generated on the server startup.
	hash.Write(b)
	binary.Write(hash, binary.LittleEndian, *po)
	hash.Write(conf.RandomValue)

	return fmt.Sprintf("%x", hash.Sum(nil))
}
