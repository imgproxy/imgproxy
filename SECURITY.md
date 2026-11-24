# Security Policy
## Supported Versions

We provide security updates for the latest minor version of imgproxy.

We perform daily vulnerability scans of imgproxy's Docker images and upgrade imgproxy's dependencies on every release. However, we do not consider vulnerabilities in imgproxy's dependencies as vulnerabilities in imgproxy itself.

## Reporting a Vulnerability

The recommended way to report a security vulnerability is [GitHub Security Advisories](https://github.com/imgproxy/imgproxy/security). This allows us to track security issues efficiently and simplifies CVE ID assignment. Alternatively, you can report security issues to [security@imgproxy.net](mailto:security@imgproxy.net).

**Please do not report** potential security issues caused by misconfigurations or usage errors. In particular:

- "`IMGPROXY_ALLOWED_SOURCES` doesn't check redirect URLs." `IMGPROXY_ALLOWED_SOURCES` is a list of trusted source patterns. If it matches sources with open redirects, it's a configuration issue, not a bug in imgproxy itself.
- "The `raw` processing option allows fetching arbitrary unprocessed files." This is what the option is designed to do. If this is an unwanted behavior, the user is expected to either use signed URLs or restrict allowed processing options via `IMGPROXY_ALLOWED_PROCESSING_OPTIONS`.
