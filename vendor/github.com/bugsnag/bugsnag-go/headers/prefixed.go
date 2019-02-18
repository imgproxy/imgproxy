package headers

import "time"

//PrefixedHeaders returns a map of Content-Type and the 'Bugsnag-' headers for
//API key, payload version, and the time at which the request is being sent.
func PrefixedHeaders(apiKey, payloadVersion string) map[string]string {
	return map[string]string{
		"Content-Type":            "application/json",
		"Bugsnag-Api-Key":         apiKey,
		"Bugsnag-Payload-Version": payloadVersion,
		"Bugsnag-Sent-At":         time.Now().UTC().Format(time.RFC3339),
	}
}
