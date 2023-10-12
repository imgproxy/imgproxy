# Serving files from OpenStack Object Storage ("Swift")

imgproxy can process images from OpenStack Object Storage, also known as Swift. To use this feature, do the following:

1. Set the `IMGPROXY_USE_SWIFT` environment variable to `true`
2. Configure Swift authentication with the following environment variables
   * `IMGPROXY_SWIFT_USERNAME`: the username for Swift API access. Default: blank
   * `IMGPROXY_SWIFT_API_KEY`: the API key for Swift API access. Default: blank
   * `IMGPROXY_SWIFT_AUTH_URL`: the Swift Auth URL. Default: blank
   * `IMGPROXY_SWIFT_AUTH_VERSION`: the Swift auth version, set to 1, 2 or 3 or leave at 0 for autodetect.
   * `IMGPROXY_SWIFT_TENANT`: the tenant name (optional, v2 auth only). Default: blank
   * `IMGPROXY_SWIFT_DOMAIN`: the Swift domain name (optional, v3 auth only): Default: blank

3. Use `swift://%{container}/%{object_path}` as the source image URL, e.g. an original object storage URL in the format of `/v1/{account}/{container}/{object_path}`, such as `http://127.0.0.1:8080/v1/AUTH_test/images/flowers/rose.jpg`, should be converted to `swift://images/flowers/rose.jpg`.
