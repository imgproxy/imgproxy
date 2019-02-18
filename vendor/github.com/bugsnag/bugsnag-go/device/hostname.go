package device

import "os"

var hostname string

// GetHostname returns the hostname of the current device. Caches the hostname
// between calls to ensure this is performant. Returns a blank string in case
// that the hostname cannot be identified.
func GetHostname() string {
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	return hostname
}
