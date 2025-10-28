package common

import (
	"slices"
)

// IsBucketAllowed checks if the provided bucket is allowed based on allowed and denied buckets lists.
func IsBucketAllowed(bucket string, allowedBuckets, deniedBuckets []string) bool {
	if len(allowedBuckets) > 0 && !slices.Contains(allowedBuckets, bucket) {
		return false
	}

	if slices.Contains(deniedBuckets, bucket) {
		return false
	}

	return true
}
