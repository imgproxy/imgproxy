# Configuration

imgproxy is [Twelve-Factor-App](https://12factor.net/)-ready and can be configured using `ENV` variables.

## URL signature

imgproxy allows URLs to be signed with a key and a salt. This feature is disabled by default, but is _highly_ recommended to be enabled in production. To enable URL signature checking, define the key/salt pair:

* `IMGPROXY_KEY`: hex-encoded key
* `IMGPROXY_SALT`: hex-encoded salt
* `IMGPROXY_SIGNATURE_SIZE`: number of bytes to use for signature before encoding to Base64. Default: 32

You can specify multiple key/salt pairs by dividing the keys and salts with a comma (`,`). imgproxy will check URL signatures with each pair. This is useful when you need to change key/salt pairs in your application while incurring zero downtime.

You can also specify file paths using the command line by referencing a separate file containing hex-coded keys and salts line by line:

```bash
imgproxy -keypath /path/to/file/with/key -saltpath /path/to/file/with/salt
```

If you need a random key/salt pair really fast, as an example, you can quickly generate one using the following snippet:

```bash
echo $(xxd -g 2 -l 64 -p /dev/random | tr -d '\n')
```

## Server

* `IMGPROXY_BIND`: the address and port or Unix socket to listen to. Default: `:8080`
* `IMGPROXY_NETWORK`: the network to use. Known networks are `tcp`, `tcp4`, `tcp6`, `unix`, and `unixpacket`. Default: `tcp`
* `IMGPROXY_READ_TIMEOUT`: the maximum duration (in seconds) for reading the entire image request, including the body. Default: `10`
* `IMGPROXY_WRITE_TIMEOUT`: the maximum duration (in seconds) for writing the response. Default: `10`
* `IMGPROXY_KEEP_ALIVE_TIMEOUT`: the maximum duration (in seconds) to wait for the next request before closing the connection. When set to `0`, keep-alive is disabled. Default: `10`
* `IMGPROXY_DOWNLOAD_TIMEOUT`: the maximum duration (in seconds) for downloading the source image. Default: `5`
* `IMGPROXY_CONCURRENCY`: the maximum number of image requests to be processed simultaneously. Default: the number of CPU cores multiplied by two
* `IMGPROXY_MAX_CLIENTS`: the maximum number of simultaneous active connections. Default: `IMGPROXY_CONCURRENCY * 10`
* `IMGPROXY_TTL`: a duration (in seconds) sent via the `Expires` and `Cache-Control: max-age` HTTP headers. Default: `3600` (1 hour)
* `IMGPROXY_CACHE_CONTROL_PASSTHROUGH`: when `true` and the source image response contains the `Expires` or `Cache-Control` headers, reuse those headers. Default: false
* `IMGPROXY_SET_CANONICAL_HEADER`: when `true` and the source image has an `http` or `https` scheme, set a `rel="canonical"` HTTP header to the value of the source image URL. More details [here](https://developers.google.com/search/docs/advanced/crawling/consolidate-duplicate-urls#rel-canonical-header-method). Default: `false`
* `IMGPROXY_SO_REUSEPORT`: when `true`, enables `SO_REUSEPORT` socket option (currently only available on Linux and macOS);
* `IMGPROXY_PATH_PREFIX`: the URL path prefix. Example: when set to `/abc/def`, the imgproxy URL will be `/abc/def/%signature/%processing_options/%source_url`. Default: blank
* `IMGPROXY_USER_AGENT`: the User-Agent header that will be sent with the source image request. Default: `imgproxy/%current_version`
* `IMGPROXY_USE_ETAG`: when set to `true`, enables using the [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) HTTP header for HTTP cache control. Default: `false`
* `IMGPROXY_ETAG_BUSTER`: change this to change ETags for all the images. Default: blank
* `IMGPROXY_CUSTOM_REQUEST_HEADERS`: <i class='badge badge-pro'></i> list of custom headers that imgproxy will send while requesting the source image, divided by `\;` (can be redefined by `IMGPROXY_CUSTOM_HEADERS_SEPARATOR`). Example: `X-MyHeader1=Lorem\;X-MyHeader2=Ipsum`
* `IMGPROXY_CUSTOM_RESPONSE_HEADERS`: <i class='badge badge-pro'></i> a list of custom response headers, separated by `\;` (can be redefined by `IMGPROXY_CUSTOM_HEADERS_SEPARATOR`). Example: `X-MyHeader1=Lorem\;X-MyHeader2=Ipsum`
* `IMGPROXY_CUSTOM_HEADERS_SEPARATOR`: <i class='badge badge-pro'></i> a string that will be used as a custom header separator. Default: `\;`
* `IMGPROXY_ENABLE_DEBUG_HEADERS`: when set to `true`, imgproxy will add debug headers to the response. Default: `false`. The following headers will be added:
  * `X-Origin-Content-Length`: the size of the source image
  * `X-Origin-Width`: the width of the source image
  * `X-Origin-Height`: the height of the source image
* `IMGPROXY_SERVER_NAME`: <i class='badge badge-pro'></i> the `Server` header value. Default: `imgproxy`

## Security

imgproxy protects you from so-called image bombs. Here's how you can specify the maximum image resolution which you consider reasonable:

* `IMGPROXY_MAX_SRC_RESOLUTION`: the maximum resolution of the source image, in megapixels. Images with larger actual size will be rejected. Default: `16.8`
* `IMGPROXY_MAX_SRC_FILE_SIZE`: the maximum size of the source image, in bytes. Images with larger file size will be rejected. When set to `0`, file size check is disabled. Default: `0`

imgproxy can process animated images (GIF, WebP), but since this operation is pretty memory heavy, only one frame is processed by default. You can increase the maximum animation frames that can be processed number of with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum number of animated image frames that may be processed. Default: `1`

**📝Note:** imgproxy summarizes all frame resolutions while checking the source image resolution.

To check if the source image is SVG, imgproxy reads some amount of bytes; by default it reads a maximum of 32KB. However, you can change this value using the following variable:

* `IMGPROXY_MAX_SVG_CHECK_BYTES`: the maximum number of bytes imgproxy will read to recognize SVG files. If imgproxy is unable to recognize your SVG, try increasing this number. Default: `32768` (32KB)

Requests to some image sources may go through too many redirects or enter an infinite loop. You can limit the number of allowed redirects:

* `IMGPROXY_MAX_REDIRECTS`: the max number of redirects imgproxy can follow while requesting the source image

You can also specify a secret key to enable authorization with the HTTP `Authorization` header for use in production environments:

* `IMGPROXY_SECRET`: the authorization token. If specified, the HTTP request should contain the `Authorization: Bearer %secret%` header.

imgproxy does not send CORS headers by default. CORS will need to be allowed by uisng the following variable:

* `IMGPROXY_ALLOW_ORIGIN`: when specified, enables CORS headers with the provided origin. CORS headers are disabled by default.

You can limit allowed source URLs with the following variable:

* `IMGPROXY_ALLOWED_SOURCES`: a whitelist of source image URL prefixes divided by comma. Wildcards can be included with `*` to match all characters except `/`. When blank, imgproxy allows all source image URLs. Example: `s3://,https://*.example.com/,local://`. Default: blank

**⚠️Warning:** Be careful when using this config to limit source URL hosts, and always add a trailing slash after the host.
* Bad: `http://example.com`
* Good: `http://example.com/`
If the trailing slash is absent, `http://example.com@baddomain.com` would be a permissable URL, however, the request would be made to `baddomain.com`.

When using imgproxy in a development environment, it can be useful to ignore SSL verification:

* `IMGPROXY_IGNORE_SSL_VERIFICATION`: when true, disables SSL verification, so imgproxy can be used in a development environment with self-signed SSL certificates.

Also you may want imgproxy to respond with the same error message that it writes to the log:

* `IMGPROXY_DEVELOPMENT_ERRORS_MODE`: when true, imgproxy will respond with detailed error messages. Not recommended for production because some errors may contain stack traces.

## Cookies

imgproxy can pass cookies in image requests. This can be activated with `IMGPROXY_COOKIE_PASSTHROUGH`. Unfortunately the `Cookie` header doesn't contain information about which URLs these cookies are applicable to, so imgproxy can only assume (or must be told).

When cookie forwarding is activated, by default, imgproxy assumes the scope of the cookies to be all URLs with the same hostname/port and request scheme as given by the headers `X-Forwarded-Host`, `X-Forwarded-Port`, `X-Forwarded-Scheme` or `Host`. To change that use `IMGPROXY_COOKIE_BASE_URL`.

* `IMGPROXY_COOKIE_PASSTHROUGH`: when `true`, incoming cookies will be passed through the image request if they are applicable for the image URL. Default: `false`

* `IMGPROXY_COOKIE_BASE_URL`: when set, assume that cookies have the scope of this URL for an incoming request (instead of using request headers). If the cookies are applicable to the image URL too, they will be passed along in the image request.


## Compression

* `IMGPROXY_QUALITY`: the default quality of the resultant image, percentage. Default: `80`
* `IMGPROXY_FORMAT_QUALITY`: default quality of the resulting image per format, separated by commas. Example: `jpeg=70,avif=40,webp=60`. When a value for the resulting format is not set, the `IMGPROXY_QUALITY` value is used. Default: `avif=50`

### Advanced JPEG compression

* `IMGPROXY_JPEG_PROGRESSIVE`: when true, enables progressive JPEG compression. Default: `false`
* `IMGPROXY_JPEG_NO_SUBSAMPLE`: <i class='badge badge-pro'></i> when true, chrominance subsampling is disabled. This will improve quality at the cost of larger file size. Default: `false`
* `IMGPROXY_JPEG_TRELLIS_QUANT`: <i class='badge badge-pro'></i> when true, enables trellis quantisation for each 8x8 block. Reduces file size but increases compression time. Default: `false`
* `IMGPROXY_JPEG_OVERSHOOT_DERINGING`: <i class='badge badge-pro'></i> when true, enables overshooting of samples with extreme values. Overshooting may reduce ringing artifacts from compression, in particular in areas where black text appears on a white background. Default: `false`
* `IMGPROXY_JPEG_OPTIMIZE_SCANS`: <i class='badge badge-pro'></i> when true, splits the spectrum of DCT coefficients into separate scans. Reduces file size but increases compression time. Requires `IMGPROXY_JPEG_PROGRESSIVE` to be true. Default: `false`
* `IMGPROXY_JPEG_QUANT_TABLE`: <i class='badge badge-pro'></i> quantization table to use. Supported values are:
  * `0`: Table from JPEG Annex K (default)
  * `1`: Flat table
  * `2`: Table tuned for MSSIM on Kodak image set
  * `3`: Table from ImageMagick by N. Robidoux
  * `4`: Table tuned for PSNR-HVS-M on Kodak image set
  * `5`: Table from Relevance of Human Vision to JPEG-DCT Compression (1992)
  * `6`: Table from DCTune Perceptual Optimization of Compressed Dental X-Rays (1997)
  * `7`: Table from A Visual Detection Model for DCT Coefficient Quantization (1993)
  * `8`: Table from An Improved Detection Model for DCT Coefficient Quantization (1993)

**📝Note:** `IMGPROXY_JPEG_TRELLIS_QUANT`, `IMGPROXY_JPEG_OVERSHOOT_DERINGING`, `IMGPROXY_JPEG_OPTIMIZE_SCANS`, and `IMGPROXY_JPEG_QUANT_TABLE` require libvips to be built with [MozJPEG](https://github.com/mozilla/mozjpeg) since standard libjpeg doesn't support those optimizations.

### Advanced PNG compression

* `IMGPROXY_PNG_INTERLACED`: when true, enables interlaced PNG compression. Default: `false`
* `IMGPROXY_PNG_QUANTIZE`: when true, enables PNG quantization. libvips should be built with [Quantizr](https://github.com/DarthSim/quantizr) or libimagequant support. Default: `false`
* `IMGPROXY_PNG_QUANTIZATION_COLORS`: maximum number of quantization palette entries. Should be between 2 and 256. Default: 256

<!-- ### Advanced GIF compression

* `IMGPROXY_GIF_OPTIMIZE_FRAMES`: <i class='badge badge-pro'></i> when true, enables GIF frame optimization. This may produce a smaller result, but may increase compression time.
* `IMGPROXY_GIF_OPTIMIZE_TRANSPARENCY`: <i class='badge badge-pro'></i> when true, enables GIF transparency optimization. This may produce a smaller result, but may also increase compression time. -->

### Advanced AVIF compression

* `IMGPROXY_AVIF_SPEED`: controls the CPU effort spent improving compression. The lowest speed is at 0 and the fastest is at 8. Default: `5`

### Autoquality

imgproxy can calculate the quality of the resulting image based on selected metric. Read more in the [Autoquality](autoquality.md) guide.

**⚠️Warning:** Autoquality requires the image to be saved several times. Use it only when you prefer the resulting size and quality over the speed.

* `IMGPROXY_AUTOQUALITY_METHOD`: <i class='badge badge-pro'></i> the method of quality calculation. Default: `none`
* `IMGPROXY_AUTOQUALITY_TARGET`: <i class='badge badge-pro'></i> desired value of the autoquality method metric. Default: 0.02
* `IMGPROXY_AUTOQUALITY_MIN`: <i class='badge badge-pro'></i> minimal quality imgproxy can use. Default: 70
* `IMGPROXY_AUTOQUALITY_FORMAT_MIN`: <i class='badge badge-pro'></i> the minimal quality imgproxy can use per format, comma divided. Example: `jpeg=70,avif=40,webp=60`. When value for the resulting format is not set, `IMGPROXY_AUTOQUALITY_MIN` value is used. Default: `avif=40`
* `IMGPROXY_AUTOQUALITY_MAX`: <i class='badge badge-pro'></i> the maximum quality imgproxy can use. Default: 80
* `IMGPROXY_AUTOQUALITY_FORMAT_MAX`: <i class='badge badge-pro'></i> the maximum quality imgproxy can use per format, comma divided. Example: `jpeg=70,avif=40,webp=60`. When a value for the resulting format is not set, the `IMGPROXY_AUTOQUALITY_MAX` value is used. Default: `avif=50`
* `IMGPROXY_AUTOQUALITY_ALLOWED_ERROR`: <i class='badge badge-pro'></i> the allowed `IMGPROXY_AUTOQUALITY_TARGET` error. Applicable only to `dssim` and `ml` methods. Default: 0.001
* `IMGPROXY_AUTOQUALITY_MAX_RESOLUTION`: <i class='badge badge-pro'></i> when this value is greater then zero and the resultant resolution exceeds the value, autoquality won't be used. Default: 0
* `IMGPROXY_AUTOQUALITY_JPEG_NET`: <i class='badge badge-pro'></i> the path to the neural network for JPEG.
* `IMGPROXY_AUTOQUALITY_WEBP_NET`: <i class='badge badge-pro'></i> the path to the neural network for WebP.
* `IMGPROXY_AUTOQUALITY_AVIF_NET`: <i class='badge badge-pro'></i> the path to the neural network for AVIF.

## AVIF/WebP support detection

imgproxy can use the `Accept` HTTP header to detect if the browser supports AVIF or WebP and use it as the default format. This feature is disabled by default and can be enabled by the following options:

* `IMGPROXY_ENABLE_WEBP_DETECTION`: enables WebP support detection. When the file extension is omitted in the imgproxy URL and browser supports WebP, imgproxy will use it as the resulting format.
* `IMGPROXY_ENFORCE_WEBP`: enables WebP support detection and enforces WebP usage. If the browser supports WebP, it will be used as resulting format even if another extension is specified in the imgproxy URL.
* `IMGPROXY_ENABLE_AVIF_DETECTION`: enables AVIF support detection. When the file extension is omitted in the imgproxy URL and browser supports AVIF, imgproxy will use it as the resulting format.
* `IMGPROXY_ENFORCE_AVIF`: enables AVIF support detection and enforces AVIF usage. If the browser supports AVIF, it will be used as resulting format even if another extension is specified in the imgproxy URL.

**📝Note:** imgproxy prefers AVIF over WebP. This means that if both AVIF and WebP detection/enforcement are enabled and the browser supports both of them, AVIF will be used.

**📝Note:** If both the source and the requested image formats support animation and AVIF detection/enforcement is enabled, AVIF won't be used as AVIF sequence is not supported yet.

**📝Note:** When AVIF/WebP support detection is enabled, please take care to configure your CDN or caching proxy to take the `Accept` HTTP header into account while caching.

**⚠️Warning:** Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Accept` HTTP headers. Keep this in mind when configuring your production caching setup.

## Client Hints support

imgproxy can use the `Width`, `Viewport-Width` or `DPR` HTTP headers to determine default width and DPR options using Client Hints. This feature is disabled by default and can be enabled by the following option:

* `IMGPROXY_ENABLE_CLIENT_HINTS`: enables Client Hints support to determine default width and DPR options. Read more details [here](https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints) about Client Hints.

**⚠️Warning:** Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Width`, `Viewport-Width` or `DPR` HTTP headers. Keep this in mind when configuring your production caching setup.

## Video thumbnails

imgproxy Pro can extract specific video frames to create thumbnails. This feature is disabled by default, but can be enabled with `IMGPROXY_ENABLE_VIDEO_THUMBNAILS`.

* `IMGPROXY_ENABLE_VIDEO_THUMBNAILS`: <i class='badge badge-pro'></i> when true, enables video thumbnail generation. Default: `false`
* `IMGPROXY_VIDEO_THUMBNAIL_SECOND`: <i class='badge badge-pro'></i> the timestamp of the frame (in seconds) that will be used for a thumbnail. Default: 1
* `IMGPROXY_VIDEO_THUMBNAIL_PROBE_SIZE`: <i class='badge badge-pro'></i> the maximum amount of bytes used to determine the format. Lower values can decrease memory usage but can produce inaccurate data, or even lead to errors. Default: 5000000
* `IMGPROXY_VIDEO_THUMBNAIL_MAX_ANALYZE_DURATION`: <i class='badge badge-pro'></i> the maximum number of milliseconds used to get the stream info. Lower values can decrease memory usage but can produce inaccurate data, or even lead to errors. When set to 0, the heuristic is used. Default: 0

**⚠️Warning:** Though using `IMGPROXY_VIDEO_THUMBNAIL_PROBE_SIZE` and `IMGPROXY_VIDEO_THUMBNAIL_MAX_ANALYZE_DURATION` can lower the memory footprint of video thumbnail generation, they should be used in production only when you know what you're doing.

## Watermark

* `IMGPROXY_WATERMARK_DATA`: Base64-encoded image data. You can easily calculate it with `base64 tmp/watermark.png | tr -d '\n'`.
* `IMGPROXY_WATERMARK_PATH`: the path to the locally stored image
* `IMGPROXY_WATERMARK_URL`: the watermark image URL
* `IMGPROXY_WATERMARK_OPACITY`: the watermark's base opacity
* `IMGPROXY_WATERMARKS_CACHE_SIZE`: <i class='badge badge-pro'></i> custom watermarks cache size. When set to `0`, the watermark cache is disabled. 256 watermarks are cached by default.

Read more about watermarks in the [Watermark](watermark.md) guide.

## Unsharpening

imgproxy Pro can apply an unsharpening mask to your images.

* `IMGPROXY_UNSHARPENING_MODE`: <i class='badge badge-pro'></i> controls when an unsharpenning mask should be applied. The following modes are supported:
  * `auto`: _(default)_ apply an unsharpening mask only when an image is downscaled and the `sharpen` option has not been set.
  * `none`: the unsharpening mask is not applied.
  * `always`: always applies the unsharpening mask.
* `IMGPROXY_UNSHARPENING_WEIGHT`: <i class='badge badge-pro'></i> a floating-point number that defines how neighboring pixels will affect the current pixel. The greater the value, the sharper the image. This value should be greater than zero. Default: `1`
* `IMGPROXY_UNSHARPENING_DIVIDOR`: <i class='badge badge-pro'></i> a floating-point number that defines the unsharpening strength. The lesser the value, the sharper the image. This value be greater than zero. Default: `24`

## Object detection

imgproxy can detect objects on the image and use them to perform smart cropping, to blur the detections, or to draw the detections.

* `IMGPROXY_OBJECT_DETECTION_CONFIG`: <i class='badge badge-pro'></i> the path to the neural network config. Default: blank
* `IMGPROXY_OBJECT_DETECTION_WEIGHTS`: <i class='badge badge-pro'></i> the path to the neural network weights. Default: blank
* `IMGPROXY_OBJECT_DETECTION_CLASSES`: <i class='badge badge-pro'></i> the path to the text file with the classes names, one per line. Default: blank
* `IMGPROXY_OBJECT_DETECTION_NET_SIZE`: <i class='badge badge-pro'></i> the size of the neural network input. The width and the heights of the inputs should be the same, so this config value should be a single number. Default: 416
* `IMGPROXY_OBJECT_DETECTION_CONFIDENCE_THRESHOLD`: <i class='badge badge-pro'></i> detections with confidences below this value will be discarded. Default: 0.2
* `IMGPROXY_OBJECT_DETECTION_NMS_THRESHOLD`: <i class='badge badge-pro'></i> non-max supression threshold. Don't change this if you don't know what you're doing. Default: 0.4

## Fallback image

You can set up a fallback image that will be used in case imgproxy is unable to fetch the requested one. Use one of the following variables:

* `IMGPROXY_FALLBACK_IMAGE_DATA`: Base64-encoded image data. You can easily calculate it with `base64 tmp/fallback.png | tr -d '\n'`.
* `IMGPROXY_FALLBACK_IMAGE_PATH`: the path to the locally stored image
* `IMGPROXY_FALLBACK_IMAGE_URL`: the fallback image URL
* `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE`: the HTTP code for the fallback image response. When set to zero, imgproxy will respond with the usual HTTP code. Default: `200`
* `IMGPROXY_FALLBACK_IMAGE_TTL`: a duration (in seconds) sent via the `Expires` and `Cache-Control: max-age` HTTP headers when a fallback image was used. When blank or `0`, the value from `IMGPROXY_TTL` is used.
* `IMGPROXY_FALLBACK_IMAGES_CACHE_SIZE`: <i class='badge badge-pro'></i> the size of custom fallback images cache. When set to `0`, the fallback image cache is disabled. 256 fallback images are cached by default.

## Skip processing

You can configure imgproxy to skip processing of some formats:

* `IMGPROXY_SKIP_PROCESSING_FORMATS`: a list of formats that imgproxy shouldn't process, comma divided.

**📝Note:** Processing can only be skipped when the requested format is the same as the source format.

**📝Note:** Video thumbnail processing can't be skipped.

## Presets

Read more about imgproxy presets in the [Presets](presets.md) guide.

There are two ways to define presets:

#### Using an environment variable

* `IMGPROXY_PRESETS`: a set of preset definitions, comma divided. Example: `default=resizing_type:fill/enlarge:1,sharp=sharpen:0.7,blurry=blur:2`. Default: blank

#### Using a command line argument

```bash
imgproxy -presets /path/to/file/with/presets
```

This file should contain preset definitions, one per line. Lines starting with `#` are treated as comments. Example:

```
default=resizing_type:fill/enlarge:1

# Sharpen the image to make it look better
sharp=sharpen:0.7

# Blur the image to hide details
blurry=blur:2
```

### Using only presets

imgproxy can be switched into "presets-only mode". In this mode, imgproxy accepts only `preset` option arguments as processing options. Example: `http://imgproxy.example.com/unsafe/thumbnail:blurry:watermarked/plain/http://example.com/images/curiosity.jpg@png`

* `IMGPROXY_ONLY_PRESETS`: disables all URL formats and enables presets-only mode.

## Serving local files

imgproxy can serve your local images, but this feature is disabled by default. To enable it, specify your local filesystem root:

* `IMGPROXY_LOCAL_FILESYSTEM_ROOT`: the root of the local filesystem. Keep this empty to disable local file serving.

Check out the [Serving local files](serving_local_files.md) guide to learn more.

## Serving files from Amazon S3

imgproxy can process files from Amazon S3 buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_S3` to `true`:

* `IMGPROXY_USE_S3`: when `true`, enables image fetching from Amazon S3 buckets. Default: `false`
* `IMGPROXY_S3_ENDPOINT`: a custom S3 endpoint to being used by imgproxy

Check out the [Serving files from S3](serving_files_from_s3.md) guide to learn more.

## Serving files from Google Cloud Storage

imgproxy can process files from Google Cloud Storage buckets, but this feature is disabled by default. To enable it, set the value of `IMGPROXY_GCS_KEY` to the content of the Google Cloud JSON key:

* `IMGPROXY_GCS_KEY`: the Google Cloud JSON key. When set, enables image fetching from Google Cloud Storage buckets. Default: blank

Check out the [Serving files from Google Cloud Storage](serving_files_from_google_cloud_storage.md) guide to learn more.

## Serving files from Azure Blob Storage

imgproxy can process files from Azure Blob Storage containers, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_ABS` to `true`:

* `IMGPROXY_USE_ABS`: when `true`, enables image fetching from Azure Blob Storage containers. Default: `false`
* `IMGPROXY_ABS_NAME`: the Azure account name. Default: blank
* `IMGPROXY_ABS_KEY`: the Azure account key. Default: blank
* `IMGPROXY_ABS_ENDPOINT`: the custom Azure Blob Storage endpoint to be used by imgproxy. Default: blank

Check out the [Serving files from Azure Blob Storage](serving_files_from_azure_blob_storage.md) guide to learn more.

## Serving files from OpenStack Object Storage ("Swift")
imgproxy can process files from OpenStack Object Storage, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_SWIFT` to `true`.
* `IMGPROXY_USE_SWIFT`: when `true`, enables image fetching from OpenStack Swift Object Storage. Default: `false`
* `IMGPROXY_SWIFT_USERNAME`: the username for Swift API access. Default: blank
* `IMGPROXY_SWIFT_API_KEY`: the API key for Swift API access. Default: blank
* `IMGPROXY_SWIFT_AUTH_URL`: the Swift Auth URL. Default: blank
* `IMGPROXY_SWIFT_AUTH_VERSION`: the Swift auth version, set to 1, 2 or 3 or leave at 0 for autodetect.
* `IMGPROXY_SWIFT_TENANT`: the tenant name (optional, v2 auth only). Default: blank
* `IMGPROXY_SWIFT_DOMAIN`: the Swift domain name (optional, v3 auth only): Default: blank
* `IMGRPOXY_SWIFT_TIMEOUT_SECONDS`: the data channel timeout in seconds. Default: 60
* `IMGRPOXY_SWIFT_CONNECT_TIMEOUT_SECONDS`: the connect channel timeout in seconds. Default: 10


## New Relic metrics

imgproxy can send its metrics to New Relic. Specify your New Relic license key to activate this feature:

* `IMGPROXY_NEW_RELIC_KEY`: the New Relic license key
* `IMGPROXY_NEW_RELIC_APP_NAME`: a New Relic application name. Default: `imgproxy`

Check out the [New Relic](new_relic.md) guide to learn more.

## Prometheus metrics

imgproxy can collect its metrics for Prometheus. Specify a binding for Prometheus metrics server to activate this feature:

* `IMGPROXY_PROMETHEUS_BIND`: Prometheus metrics server binding. Can't be the same as `IMGPROXY_BIND`. Default: blank
* `IMGPROXY_PROMETHEUS_NAMESPACE`: Namespace (prefix) for imgproxy metrics. Default: blank

Check out the [Prometheus](prometheus.md) guide to learn more.

## Datadog metrics

imgproxy can send its metrics to Datadog:

* `IMGPROXY_DATADOG_ENABLE`: when `true`, enables sending metrics to Datadog. Default: false

Check out the [Datadog](datadog.md) guide to learn more.

## Error reporting

imgproxy can report occurred errors to Bugsnag, Honeybadger and Sentry:

* `IMGPROXY_BUGSNAG_KEY`: Bugsnag API key. When provided, enables error reporting to Bugsnag.
* `IMGPROXY_BUGSNAG_STAGE`: the Bugsnag stage to report to. Default: `production`
* `IMGPROXY_HONEYBADGER_KEY`: the Honeybadger API key. When provided, enables error reporting to Honeybadger.
* `IMGPROXY_HONEYBADGER_ENV`: the Honeybadger env to report to. Default: `production`
* `IMGPROXY_SENTRY_DSN`: Sentry project DSN. When provided, enables error reporting to Sentry.
* `IMGPROXY_SENTRY_ENVIRONMENT`: the Sentry environment to report to. Default: `production`
* `IMGPROXY_SENTRY_RELEASE`: the Sentry release to report to. Default: `imgproxy@{imgproxy version}`
* `IMGPROXY_AIRBRAKE_PROJECT_ID`: an Airbrake project id
* `IMGPROXY_AIRBRAKE_PROJECT_KEY`: an Airbrake project key
* `IMGPROXY_AIRBRAKE_ENVIRONMENT`: the Airbrake environment to report to. Default: `production`
* `IMGPROXY_REPORT_DOWNLOADING_ERRORS`: when `true`, imgproxy will report downloading errors. Default: `true`

## Log

* `IMGPROXY_LOG_FORMAT`: the log format. The following formats are supported:
  * `pretty`: _(default)_ colored human-readable format
  * `structured`: machine-readable format
  * `json`: JSON format
* `IMGPROXY_LOG_LEVEL`: the log level. The following levels are supported `error`, `warn`, `info` and `debug`. Default: `info`

imgproxy can send logs to syslog, but this feature is disabled by default. To enable it, set `IMGPROXY_SYSLOG_ENABLE` to `true`:

* `IMGPROXY_SYSLOG_ENABLE`: when `true`, enables sending logs to syslog.
* `IMGPROXY_SYSLOG_LEVEL`: the maximum log level to send to syslog. Known levels are: `crit`, `error`, `warning` and `info`. Default: `info`
* `IMGPROXY_SYSLOG_NETWORK`: the network that will be used to connect to syslog. When blank, the local syslog server will be used. Known networks are `tcp`, `tcp4`, `tcp6`, `udp`, `udp4`, `udp6`, `ip`, `ip4`, `ip6`, `unix`, `unixgram` and `unixpacket`. Default: blank
* `IMGPROXY_SYSLOG_ADDRESS`: the address of the syslog service. Not used if `IMGPROXY_SYSLOG_NETWORK` is blank. Default: blank
* `IMGPROXY_SYSLOG_TAG`: the specific syslog tag. Default: `imgproxy`

**📝Note:** imgproxy always uses structured log format for syslog.

## Memory usage tweaks

**⚠️Warning:** We highly recommended reading the [Memory usage tweaks](memory_usage_tweaks.md) guide before changing these settings.

* `IMGPROXY_DOWNLOAD_BUFFER_SIZE`: the initial size (in bytes) of a single download buffer. When set to zero, initializes empty download buffers. Default: `0`
* `IMGPROXY_GZIP_BUFFER_SIZE`: the initial size (in bytes) of a single GZip buffer. When zero, initializes empty GZip buffers. This makess sense only when GZip compression is enabled. Default: `0`
* `IMGPROXY_FREE_MEMORY_INTERVAL`: the interval (in seconds) at which unused memory will be returned to the OS. Default: `10`
* `IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD`: the number of buffers that should be returned to a pool before calibration. Default: `1024`

## Miscellaneous

* `IMGPROXY_BASE_URL`: a base URL prefix that will be added to each requested image URL. For example, if the base URL is `http://example.com/images` and `/path/to/image.png` is requested, imgproxy will download the source image from `http://example.com/images/path/to/image.png`. If the image URL already contains the prefix, it won't be added. Default: blank
* `IMGPROXY_USE_LINEAR_COLORSPACE`: when `true`, imgproxy will process images in linear colorspace. This will slow down processing. Note that images won't be fully processed in linear colorspace while shrink-on-load is enabled (see below).
* `IMGPROXY_DISABLE_SHRINK_ON_LOAD`: when `true`, disables shrink-on-load for JPEGs and WebP files. Allows processing the entire image in linear colorspace but dramatically slows down resizing and increases memory usage when working with large images.
* `IMGPROXY_STRIP_METADATA`: when `true`, imgproxy will strip all metadata (EXIF, IPTC, etc.) from JPEG and WebP output images. Default: `true`
* `IMGPROXY_STRIP_COLOR_PROFILE`: when `true`, imgproxy will transform the embedded color profile (ICC) to sRGB and remove it from the image. Otherwise, imgproxy will try to keep it as is. Default: `true`
* `IMGPROXY_AUTO_ROTATE`: when `true`, imgproxy will automatically rotate images based on the EXIF Orientation parameter (if available in the image meta data). The orientation tag will be removed from the image in all cases. Default: `true`
* `IMGPROXY_HEALTH_CHECK_MESSAGE`: <i class='badge badge-pro'></i> the content of the health check response. Default: `imgproxy is running`
* `IMGPROXY_HEALTH_CHECK_PATH`: an additional path of the health check. Default: blank
