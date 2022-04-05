# Serving files from OpenStack Object Storage ("Swift")

imgproxy can process images from OpenStack Object Storage, also known as Swift. To use this feature, do the following:

1. Set `IMGPROXY_USE_SWIFT` environment variable to `true`
2. Configure Swift authentication with the following environment variables
   * `IMGPROXY_SWIFT_USERNAME`: the username for Swift API access
   * `IMGPROXY_SWIFT_API_KEY`: the API key for Swift API access
   * `IMGPROXY_SWIFT_AUTH_URL`: the Swift Auth URL
   * `IMGPROXY_SWIFT_AUTH_VERSION`: the Swift auth version, set to 1, 2 or 3 or leave at 0 for autodetect.
   * `IMGPROXY_SWIFT_TENANT`: Name of tenant (optional, v2 auth only)
   * `IMGPROXY_SWIFT_DOMAIN` | Swift domain name (optional, v3 auth only)

3. Use `swift:///%{account}/%{container}/%{object_path}` as the source image URL
