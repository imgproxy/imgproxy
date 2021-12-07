# Configuration

imgproxy is [Twelve-Factor-App](https://12factor.net/)-ready and can be configured using `ENV` variables.

## URL signature

imgproxy allows URLs to be signed with a key and salt. This feature is disabled by default, but it is _highly_ recommended to enable it in production. To enable URL signature checking, define the key/salt pair:

* `IMGPROXY_KEY`: hex-encoded key;
* `IMGPROXY_SALT`: hex-encoded salt;
* `IMGPROXY_SIGNATURE_SIZE`: number of bytes to use for signature before encoding to Base64. Default: 32;

You can specify multiple key/salt pairs by dividing keys and salts with comma (`,`). imgproxy will check URL signatures with each pair. Useful when you need to change key/salt pair in your application with zero downtime.

You can also specify paths to files with a hex-encoded keys and salts, one by line (useful in a development environment):

```bash
imgproxy -keypath /path/to/file/with/key -saltpath /path/to/file/with/salt
```

If you need a random key/salt pair real fast, you can quickly generate it using, for example, the following snippet:

```bash
echo $(xxd -g 2 -l 64 -p /dev/random | tr -d '\n')
```

## Server

* `IMGPROXY_BIND`: address and port or Unix socket to listen on. Default: `:8080`;
* `IMGPROXY_NETWORK`: network to use. Known networks are `tcp`, `tcp4`, `tcp6`, `unix`, and `unixpacket`. Default: `tcp`;
* `IMGPROXY_READ_TIMEOUT`: the maximum duration (in seconds) for reading the entire image request, including the body. Default: `10`;
* `IMGPROXY_WRITE_TIMEOUT`: the maximum duration (in seconds) for writing the response. Default: `10`;
* `IMGPROXY_KEEP_ALIVE_TIMEOUT`: the maximum duration (in seconds) to wait for the next request before closing the connection. When set to `0`, keep-alive is disabled. Default: `10`;
* `IMGPROXY_DOWNLOAD_TIMEOUT`: the maximum duration (in seconds) for downloading the source image. Default: `5`;
* `IMGPROXY_CONCURRENCY`: the maximum number of image requests to be processed simultaneously. Default: number of CPU cores times two;
* `IMGPROXY_MAX_CLIENTS`: the maximum number of simultaneous active connections. Default: `IMGPROXY_CONCURRENCY * 10`;
* `IMGPROXY_TTL`: duration (in seconds) sent in `Expires` and `Cache-Control: max-age` HTTP headers. Default: `3600` (1 hour);
* `IMGPROXY_CACHE_CONTROL_PASSTHROUGH`: when `true` and source image response contains `Expires` or `Cache-Control` headers, reuse those headers. Default: false;
* `IMGPROXY_SET_CANONICAL_HEADER`: when `true` and the source image has `http` or `https` scheme, set `rel="canonical"` HTTP header to the value of the source image URL. More details [here](https://developers.google.com/search/docs/advanced/crawling/consolidate-duplicate-urls#rel-canonical-header-method). Default: false;
* `IMGPROXY_SO_REUSEPORT`: when `true`, enables `SO_REUSEPORT` socket option (currently on linux and darwin only);
* `IMGPROXY_PATH_PREFIX`: URL path prefix. Example: when set to `/abc/def`, imgproxy URL will be `/abc/def/%signature/%processing_options/%source_url`. Default: blank.
* `IMGPROXY_USER_AGENT`: User-Agent header that will be sent with source image request. Default: `imgproxy/%current_version`;
* `IMGPROXY_USE_ETAG`: when `true`, enables using [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) HTTP header for HTTP cache control. Default: false;
* `IMGPROXY_ETAG_BUSTER`: change this to change ETags for all the images. Default: blank.
* `IMGPROXY_CUSTOM_REQUEST_HEADERS`: <i class='badge badge-pro'></i> list of custom headers that imgproxy will send while requesting the source image, divided by `\;` (can be redefined by `IMGPROXY_CUSTOM_HEADERS_SEPARATOR`). Example: `X-MyHeader1=Lorem\;X-MyHeader2=Ipsum`;
* `IMGPROXY_CUSTOM_RESPONSE_HEADERS`: <i class='badge badge-pro'></i> list of custom response headers, divided by `\;` (can be redefined by `IMGPROXY_CUSTOM_HEADERS_SEPARATOR`). Example: `X-MyHeader1=Lorem\;X-MyHeader2=Ipsum`;
* `IMGPROXY_CUSTOM_HEADERS_SEPARATOR`: <i class='badge badge-pro'></i> string that will be used as a custom headers separator. Default: `\;`;
* `IMGPROXY_ENABLE_DEBUG_HEADERS`: when `true`, imgproxy will add debug headers to the response. Default: `false`. The following headers will be added:
  * `X-Origin-Content-Length`: size of the source image.
  * `X-Origin-Width`: width of the source image.
  * `X-Origin-Height`: height of the source image.

## Security

imgproxy protects you from so-called image bombs. Here is how you can specify maximum image resolution which you consider reasonable:

* `IMGPROXY_MAX_SRC_RESOLUTION`: the maximum resolution of the source image, in megapixels. Images with larger actual size will be rejected. Default: `16.8`;
* `IMGPROXY_MAX_SRC_FILE_SIZE`: the maximum size of the source image, in bytes. Images with larger file size will be rejected. When `0`, file size check is disabled. Default: `0`;

imgproxy can process animated images (GIF, WebP), but since this operation is pretty heavy, only one frame is processed by default. You can increase the maximum of animation frames to process with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum of animated image frames to being processed. Default: `1`.

**📝Note:** imgproxy summarizes all frames resolutions while checking source image resolution.

imgproxy reads some amount of bytes to check if the source image is SVG. By default it reads maximum of 32KB, but you can change this:

* `IMGPROXY_MAX_SVG_CHECK_BYTES`: the maximum number of bytes imgproxy will read to recognize SVG. If imgproxy can't recognize your SVG, try to increase this number. Default: `32768` (32KB)

You can also specify a secret to enable authorization with the HTTP `Authorization` header for use in production environments:

* `IMGPROXY_SECRET`: the authorization token. If specified, the HTTP request should contain the `Authorization: Bearer %secret%` header;

imgproxy does not send CORS headers by default. Specify allowed origin to enable CORS headers:

* `IMGPROXY_ALLOW_ORIGIN`: when set, enables CORS headers with provided origin. CORS headers are disabled by default.

You can limit allowed source URLs:

* `IMGPROXY_ALLOWED_SOURCES`: whitelist of source image URLs prefixes divided by comma. Wildcards can be included with `*` to match all characters except `/`. When blank, imgproxy allows all source image URLs. Example: `s3://,https://*.example.com/,local://`. Default: blank.

**⚠️Warning:** Be careful when using this config to limit source URL hosts, and always add a trailing slash after the host. Bad: `http://example.com`, good: `http://example.com/`. If you don't add a trailing slash, `http://example.com@baddomain.com` will be an allowed URL but the request will be made to `baddomain.com`.

When you use imgproxy in a development environment, it can be useful to ignore SSL verification:

* `IMGPROXY_IGNORE_SSL_VERIFICATION`: when true, disables SSL verification, so imgproxy can be used in a development environment with self-signed SSL certificates.

Also you may want imgproxy to respond with the same error message that it writes to the log:

* `IMGPROXY_DEVELOPMENT_ERRORS_MODE`: when true, imgproxy will respond with detailed error messages. Not recommended for production because some errors may contain stack trace.

## Cookies

imgproxy can pass through cookies in image requests. This can be activated with `IMGPROXY_COOKIE_PASSTHROUGH`. Unfortunately a `Cookie` header doesn't contain information for which URLs these cookies are applicable, so imgproxy can only assume (or must be told).

When cookie forwarding is activated, imgproxy by default assumes the scope of the cookies to be all URLs with the same hostname/port and request scheme as given by the headers `X-Forwarded-Host`, `X-Forwarded-Port`, `X-Forwarded-Scheme` or `Host`. To change that use `IMGPROXY_COOKIE_BASE_URL`.

* `IMGPROXY_COOKIE_PASSTHROUGH`: when `true`, incoming cookies will be passed through to the image request if they are applicable for the image URL. Default: false;

* `IMGPROXY_COOKIE_BASE_URL`: when set, assume that cookies have a scope of this URL for the incoming request (instead of using the request headers). If the cookies are applicable to the image URL too, they will be passed along in the image request.


## Compression

* `IMGPROXY_QUALITY`: default quality of the resulting image, percentage. Default: `80`;
* `IMGPROXY_FORMAT_QUALITY`: default quality of the resulting image per format, comma divided. Example: `jpeg=70,avif=40,webp=60`. When value for the resulting format is not set, `IMGPROXY_QUALITY` value is used. Default: `avif=50`.

### Advanced JPEG compression

* `IMGPROXY_JPEG_PROGRESSIVE`: when true, enables progressive JPEG compression. Default: false;
* `IMGPROXY_JPEG_NO_SUBSAMPLE`: <i class='badge badge-pro'></i> when true, chrominance subsampling is disabled. This will improve quality at the cost of larger file size. Default: false;
* `IMGPROXY_JPEG_TRELLIS_QUANT`: <i class='badge badge-pro'></i> when true, enables trellis quantisation for each 8x8 block. Reduces file size but increases compression time. Default: false;
* `IMGPROXY_JPEG_OVERSHOOT_DERINGING`: <i class='badge badge-pro'></i> when true, enables overshooting of samples with extreme values. Overshooting may reduce ringing artifacts from compression, in particular in areas where black text appears on a white background. Default: false;
* `IMGPROXY_JPEG_OPTIMIZE_SCANS`: <i class='badge badge-pro'></i> when true, split the spectrum of DCT coefficients into separate scans. Reduces file size but increases compression time. Requires `IMGPROXY_JPEG_PROGRESSIVE` to be true. Default: false;
* `IMGPROXY_JPEG_QUANT_TABLE`: <i class='badge badge-pro'></i> quantization table to use. Supported values are:
  * `0`: Table from JPEG Annex K (default);
  * `1`: Flat table;
  * `2`: Table tuned for MSSIM on Kodak image set;
  * `3`: Table from ImageMagick by N. Robidoux;
  * `4`: Table tuned for PSNR-HVS-M on Kodak image set;
  * `5`: Table from Relevance of Human Vision to JPEG-DCT Compression (1992);
  * `6`: Table from DCTune Perceptual Optimization of Compressed Dental X-Rays (1997);
  * `7`: Table from A Visual Detection Model for DCT Coefficient Quantization (1993);
  * `8`: Table from An Improved Detection Model for DCT Coefficient Quantization (1993).

**📝Note:** `IMGPROXY_JPEG_TRELLIS_QUANT`, `IMGPROXY_JPEG_OVERSHOOT_DERINGING`, `IMGPROXY_JPEG_OPTIMIZE_SCANS`, and `IMGPROXY_JPEG_QUANT_TABLE` require libvips to be built with [MozJPEG](https://github.com/mozilla/mozjpeg) since standard libjpeg doesn't support those optimizations.

### Advanced PNG compression

* `IMGPROXY_PNG_INTERLACED`: when true, enables interlaced PNG compression. Default: false;
* `IMGPROXY_PNG_QUANTIZE`: when true, enables PNG quantization. libvips should be built with [Quantizr](https://github.com/DarthSim/quantizr) or libimagequant support. Default: false;
* `IMGPROXY_PNG_QUANTIZATION_COLORS`: maximum number of quantization palette entries. Should be between 2 and 256. Default: 256;

### Advanced GIF compression

* `IMGPROXY_GIF_OPTIMIZE_FRAMES`: <i class='badge badge-pro'></i> when true, enables GIF frames optimization. This may produce a smaller result, but may increase compression time.
* `IMGPROXY_GIF_OPTIMIZE_TRANSPARENCY`: <i class='badge badge-pro'></i> when true, enables GIF transparency optimization. This may produce a smaller result, but may increase compression time.

### Advanced AVIF compression

* `IMGPROXY_AVIF_SPEED`: controls the CPU effort spent improving compression. 0 slowest - 8 fastest. Default: `5`;

### Autoquality

imgproxy can calculate the quality of the resulting image based on selected metric. Read more in the [Autoquality](autoquality.md) guide.

**⚠️Warning:** Autoquality requires the image to be saved several times. Use it only when you prefer the resulting size and quality over the speed.

* `IMGPROXY_AUTOQUALITY_METHOD`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> the method of quality calculation. Default: `none`.
* `IMGPROXY_AUTOQUALITY_TARGET`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> desired value of the autoquality method metric. Default: 0.02.
* `IMGPROXY_AUTOQUALITY_MIN`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> minimal quality imgproxy can use. Default: 70.
* `IMGPROXY_AUTOQUALITY_FORMAT_MIN`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> minimal quality imgproxy can use per format, comma divided. Example: `jpeg=70,avif=40,webp=60`. When value for the resulting format is not set, `IMGPROXY_AUTOQUALITY_MIN` value is used. Default: `avif=40`.
* `IMGPROXY_AUTOQUALITY_MAX`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> maximal quality imgproxy can use. Default: 80.
* `IMGPROXY_AUTOQUALITY_FORMAT_MAX`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> maximal quality imgproxy can use per format, comma divided. Example: `jpeg=70,avif=40,webp=60`. When value for the resulting format is not set, `IMGPROXY_AUTOQUALITY_MAX` value is used. Default: `avif=50`.
* `IMGPROXY_AUTOQUALITY_ALLOWED_ERROR`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> allowed `IMGPROXY_AUTOQUALITY_TARGET` error. Applicable only to `dssim` and `ml` methods. Default: 0.001.
* `IMGPROXY_AUTOQUALITY_MAX_RESOLUTION`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> when value is greater then zero and the result resolution exceeds the value, autoquality won't be used. Default: 0.
* `IMGPROXY_AUTOQUALITY_JPEG_NET`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> path to the neural network for JPEG.
* `IMGPROXY_AUTOQUALITY_WEBP_NET`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> path to the neural network for WebP.
* `IMGPROXY_AUTOQUALITY_AVIF_NET`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> path to the neural network for AVIF.

## AVIF/WebP support detection

imgproxy can use the `Accept` HTTP header to detect if the browser supports AVIF or WebP and use it as the default format. This feature is disabled by default and can be enabled by the following options:

* `IMGPROXY_ENABLE_WEBP_DETECTION`: enables WebP support detection. When the file extension is omitted in the imgproxy URL and browser supports WebP, imgproxy will use it as the resulting format;
* `IMGPROXY_ENFORCE_WEBP`: enables WebP support detection and enforces WebP usage. If the browser supports WebP, it will be used as resulting format even if another extension is specified in the imgproxy URL.
* `IMGPROXY_ENABLE_AVIF_DETECTION`: enables AVIF support detection. When the file extension is omitted in the imgproxy URL and browser supports AVIF, imgproxy will use it as the resulting format;
* `IMGPROXY_ENFORCE_AVIF`: enables AVIF support detection and enforces AVIF usage. If the browser supports AVIF, it will be used as resulting format even if another extension is specified in the imgproxy URL.

**📝Note:** imgproxy prefers AVIF over WebP. This means that if both AVIF and WebP detection/enforcement are enabled and the browser supports both of them, AVIF will be used.

**📝Note:** If both the source and the requested image formats support animation and AVIF detection/enforcement is enabled, AVIF won't be used as AVIF sequence is not supported yet.

**📝Note:** When AVIF/WebP support detection is enabled, please take care to configure your CDN or caching proxy to take the `Accept` HTTP header into account while caching.

**⚠️Warning:** Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Accept` HTTP headers. Have this in mind when configuring your production caching setup.

## Client Hints support

imgproxy can use the `Width`, `Viewport-Width` or `DPR` HTTP headers to determine default width and DPR options using Client Hints. This feature is disabled by default and can be enabled by the following option:

* `IMGPROXY_ENABLE_CLIENT_HINTS`: enables Client Hints support to determine default width and DPR options. Read [here](https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints) details about Client Hints.

**⚠️Warning:** Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Width`, `Viewport-Width` or `DPR` HTTP headers. Have this in mind when configuring your production caching setup.

## Video thumbnails

imgproxy Pro can extract specific frames of videos to create thumbnails. The feature is disabled by default, but can be enabled with `IMGPROXY_ENABLE_VIDEO_THUMBNAILS`.

* `IMGPROXY_ENABLE_VIDEO_THUMBNAILS`: <i class='badge badge-pro'></i> then true, enables video thumbnails generation. Default: false;
* `IMGPROXY_VIDEO_THUMBNAIL_SECOND`: <i class='badge badge-pro'></i> the timestamp of the frame in seconds that will be used for a thumbnail. Default: 1.
* `IMGPROXY_VIDEO_THUMBNAIL_PROBE_SIZE`: <i class='badge badge-pro'></i> the maximum amount of bytes used to determine the format. Lower values can decrease memory usage but can produce inaccurate data or even lead to errors. Default: 5000000.
* `IMGPROXY_VIDEO_THUMBNAIL_MAX_ANALYZE_DURATION`: <i class='badge badge-pro'></i> the maximum of milliseconds used to get the stream info. Low values can decrease memory usage but can produce inaccurate data or even lead to errors. When set to 0, the heuristic is used. Default: 0.

**⚠️Warning:** Though using `IMGPROXY_VIDEO_THUMBNAIL_PROBE_SIZE` and `IMGPROXY_VIDEO_THUMBNAIL_MAX_ANALYZE_DURATION` can lower the memory footprint of video thumbnails generation, you should use them in production only when you know what are you doing.

## Watermark

* `IMGPROXY_WATERMARK_DATA`: Base64-encoded image data. You can easily calculate it with `base64 tmp/watermark.png | tr -d '\n'`;
* `IMGPROXY_WATERMARK_PATH`: path to the locally stored image;
* `IMGPROXY_WATERMARK_URL`: watermark image URL;
* `IMGPROXY_WATERMARK_OPACITY`: watermark base opacity;
* `IMGPROXY_WATERMARKS_CACHE_SIZE`: <i class='badge badge-pro'></i> size of custom watermarks cache. When set to `0`, watermarks cache is disabled. By default 256 watermarks are cached.

Read more about watermarks in the [Watermark](watermark.md) guide.

## Unsharpening

imgproxy Pro can apply unsharpening mask to your images.

* `IMGPROXY_UNSHARPENING_MODE`: <i class='badge badge-pro'></i> controls when unsharpenning mask should be applied. The following modes are supported:
  * `auto`: _(default)_ apply unsharpening mask only when image is downscaled and `sharpen` option is not set.
  * `none`: don't apply the unsharpening mask.
  * `always`: always apply the unsharpening mask.
* `IMGPROXY_UNSHARPENING_WEIGHT`: <i class='badge badge-pro'></i> a floating-point number that defines how neighbor pixels will affect the current pixel. Greater the value - sharper the image. Should be greater than zero. Default: `1`.
* `IMGPROXY_UNSHARPENING_DIVIDOR`: <i class='badge badge-pro'></i> a floating-point number that defines the unsharpening strength. Lesser the value - sharper the image. Should be greater than zero. Default: `24`.

## Object detection

imgproxy can detect objects on the image and use them for smart crop, bluring the detections, or drawing the detections.

* `IMGPROXY_OBJECT_DETECTION_CONFIG`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> path to the neural network config. Default: blank.
* `IMGPROXY_OBJECT_DETECTION_WEIGHTS`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> path to the neural network weights. Default: blank.
* `IMGPROXY_OBJECT_DETECTION_CLASSES`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> path to the text file with the classes names, one by line. Default: blank.
* `IMGPROXY_OBJECT_DETECTION_NET_SIZE`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> the size of the neural network input. The width and the heights of the inputs should be the same, so this config value should be a single number. Default: 416.
* `IMGPROXY_OBJECT_DETECTION_CONFIDENCE_THRESHOLD`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> the detections with confidences below this value will be discarded. Default: 0.2.
* `IMGPROXY_OBJECT_DETECTION_NMS_THRESHOLD`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> non max supression threshold. Don't change this if you don't know what you're doing. Default: 0.4.

## Fallback image

You can set up a fallback image that will be used in case imgproxy can't fetch the requested one. Use one of the following variables:

* `IMGPROXY_FALLBACK_IMAGE_DATA`: Base64-encoded image data. You can easily calculate it with `base64 tmp/fallback.png | tr -d '\n'`;
* `IMGPROXY_FALLBACK_IMAGE_PATH`: path to the locally stored image;
* `IMGPROXY_FALLBACK_IMAGE_URL`: fallback image URL.
* `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE`: <i class='badge badge-v3'></i> HTTP code for the fallback image response. When set to zero, imgproxy will respond with the usual HTTP code. Default: `200`.
* `IMGPROXY_FALLBACK_IMAGES_CACHE_SIZE`: <i class='badge badge-pro'></i> <i class='badge badge-v3'></i> size of custom fallback images cache. When set to `0`, fallback images cache is disabled. By default 256 fallback images are cached.

## Skip processing

You can configure imgproxy to skip processing of some formats:

* `IMGPROXY_SKIP_PROCESSING_FORMATS`: list of formats that imgproxy shouldn't process, comma-divided.

**📝Note:** Processing can be skipped only when the requested format is the same as the source format.

**📝Note:** Video thumbnail processing can't be skipped.

## Presets

Read about imgproxy presets in the [Presets](presets.md) guide.

There are two ways to define presets:

#### Using an environment variable

* `IMGPROXY_PRESETS`: set of preset definitions, comma-divided. Example: `default=resizing_type:fill/enlarge:1,sharp=sharpen:0.7,blurry=blur:2`. Default: blank.

#### Using a command line argument

```bash
imgproxy -presets /path/to/file/with/presets
```

The file should contain preset definitions, one per line. Lines starting with `#` are treated as comments. Example:

```
default=resizing_type:fill/enlarge:1

# Sharpen the image to make it look better
sharp=sharpen:0.7

# Blur the image to hide details
blurry=blur:2
```

### Using only presets

imgproxy can be switched into "presets-only mode". In this mode, imgproxy accepts only `preset` option arguments as processing options. Example: `http://imgproxy.example.com/unsafe/thumbnail:blurry:watermarked/plain/http://example.com/images/curiosity.jpg@png`

* `IMGPROXY_ONLY_PRESETS`: disable all URL formats and enable presets-only mode.

## Serving local files

imgproxy can serve your local images, but this feature is disabled by default. To enable it, specify your local filesystem root:

* `IMGPROXY_LOCAL_FILESYSTEM_ROOT`: the root of the local filesystem. Keep empty to disable serving of local files.

Check out the [Serving local files](serving_local_files.md) guide to learn more.

## Serving files from Amazon S3

imgproxy can process files from Amazon S3 buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_S3` to `true`:

* `IMGPROXY_USE_S3`: when `true`, enables image fetching from Amazon S3 buckets. Default: false;
* `IMGPROXY_S3_ENDPOINT`: custom S3 endpoint to being used by imgproxy.

Check out the [Serving files from S3](serving_files_from_s3.md) guide to learn more.

## Serving files from Google Cloud Storage

imgproxy can process files from Google Cloud Storage buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_GCS_KEY` to the content of Google Cloud JSON key:

* `IMGPROXY_GCS_KEY`: Google Cloud JSON key. When set, enables image fetching from Google Cloud Storage buckets. Default: blank.

Check out the [Serving files from Google Cloud Storage](serving_files_from_google_cloud_storage.md) guide to learn more.

## Serving files from Azure Blob Storage

imgproxy can process files from Azure Blob Storage containers, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_ABS` to `true`:

* `IMGPROXY_USE_ABS`: when `true`, enables image fetching from Azure Blob Storage containers. Default: false;
* `IMGPROXY_ABS_NAME`: Azure account name. Default: blank;
* `IMGPROXY_ABS_KEY`: Azure account key. Default: blank;
* `IMGPROXY_ABS_ENDPOINT`: custom Azure Blob Storage endpoint to being used by imgproxy. Default: blank.

Check out the [Serving files from Azure Blob Storage](serving_files_from_azure_blob_storage.md) guide to learn more.

## New Relic metrics

imgproxy can send its metrics to New Relic. Specify your New Relic license key to activate this feature:

* `IMGPROXY_NEW_RELIC_KEY`: New Relic license key;
* `IMGPROXY_NEW_RELIC_APP_NAME`: New Relic application name. Default: `imgproxy`.

Check out the [New Relic](new_relic.md) guide to learn more.

## Prometheus metrics

imgproxy can collect its metrics for Prometheus. Specify binding for Prometheus metrics server to activate this feature:

* `IMGPROXY_PROMETHEUS_BIND`: Prometheus metrics server binding. Can't be the same as `IMGPROXY_BIND`. Default: blank.
* `IMGPROXY_PROMETHEUS_NAMESPACE`: Namespace (prefix) for imgproxy metrics. Default: blank.

Check out the [Prometheus](prometheus.md) guide to learn more.

## Datadog metrics

imgproxy can send its metrics to Datadog:

* `IMGPROXY_DATADOG_ENABLE`: <i class='badge badge-v3'></i> when `true`, enables sending metrics to Datadog. Default: false;

Check out the [Datadog](datadog.md) guide to learn more.

## Error reporting

imgproxy can report occurred errors to Bugsnag, Honeybadger and Sentry:

* `IMGPROXY_BUGSNAG_KEY`: Bugsnag API key. When provided, enables error reporting to Bugsnag;
* `IMGPROXY_BUGSNAG_STAGE`: Bugsnag stage to report to. Default: `production`;
* `IMGPROXY_HONEYBADGER_KEY`: Honeybadger API key. When provided, enables error reporting to Honeybadger;
* `IMGPROXY_HONEYBADGER_ENV`: Honeybadger env to report to. Default: `production`;
* `IMGPROXY_SENTRY_DSN`: Sentry project DSN. When provided, enables error reporting to Sentry;
* `IMGPROXY_SENTRY_ENVIRONMENT`: Sentry environment to report to. Default: `production`;
* `IMGPROXY_SENTRY_RELEASE`: Sentry release to report to. Default: `imgproxy@{imgproxy version}`;
* `IMGPROXY_AIRBRAKE_PROJECT_ID`: Airbrake project id;
* `IMGPROXY_AIRBRAKE_PROJECT_KEY`: Airbrake project key;
* `IMGPROXY_AIRBRAKE_ENVIRONMENT`: Airbrake environment to report to. Default: `production`;
* `IMGPROXY_REPORT_DOWNLOADING_ERRORS`: when `true`, imgproxy will report downloading errors. Default: `true`.

## Log

* `IMGPROXY_LOG_FORMAT`: the log format. The following formats are supported:
  * `pretty`: _(default)_ colored human-readable format;
  * `structured`: machine-readable format;
  * `json`: JSON format;
* `IMGPROXY_LOG_LEVEL`: the log level. The following levels are supported `error`, `warn`, `info` and `debug`. Default: `info`;

imgproxy can send logs to syslog, but this feature is disabled by default. To enable it, set `IMGPROXY_SYSLOG_ENABLE` to `true`:

* `IMGPROXY_SYSLOG_ENABLE`: when `true`, enables sending logs to syslog;
* `IMGPROXY_SYSLOG_LEVEL`: maximum log level to send to syslog. Known levels are: `crit`, `error`, `warning` and `info`. Default: `info`;
* `IMGPROXY_SYSLOG_NETWORK`: network that will be used to connect to syslog. When blank, the local syslog server will be used. Known networks are `tcp`, `tcp4`, `tcp6`, `udp`, `udp4`, `udp6`, `ip`, `ip4`, `ip6`, `unix`, `unixgram` and `unixpacket`. Default: blank;
* `IMGPROXY_SYSLOG_ADDRESS`: address of the syslog service. Not used if `IMGPROXY_SYSLOG_NETWORK` is blank. Default: blank;
* `IMGPROXY_SYSLOG_TAG`: specific syslog tag. Default: `imgproxy`;

**📝Note:** imgproxy always uses structured log format for syslog.

## Memory usage tweaks

**⚠️Warning:** It's highly recommended to read [Memory usage tweaks](memory_usage_tweaks.md) guide before changing this settings.

* `IMGPROXY_DOWNLOAD_BUFFER_SIZE`: the initial size (in bytes) of a single download buffer. When zero, initializes empty download buffers. Default: `0`;
* `IMGPROXY_GZIP_BUFFER_SIZE`: the initial size (in bytes) of a single GZip buffer. When zero, initializes empty GZip buffers. Makes sense only when GZip compression is enabled. Default: `0`;
* `IMGPROXY_FREE_MEMORY_INTERVAL`: the interval (in seconds) at which unused memory will be returned to the OS. Default: `10`;
* `IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD`: the number of buffers that should be returned to a pool before calibration. Default: `1024`.

## Miscellaneous

* `IMGPROXY_BASE_URL`: base URL prefix that will be added to every requested image URL. For example, if the base URL is `http://example.com/images` and `/path/to/image.png` is requested, imgproxy will download the source image from `http://example.com/images/path/to/image.png`. If the image URL already contains the prefix, it won't be added. Default: blank.
* `IMGPROXY_USE_LINEAR_COLORSPACE`: when `true`, imgproxy will process images in linear colorspace. This will slow down processing. Note that images won't be fully processed in linear colorspace while shrink-on-load is enabled (see below).
* `IMGPROXY_DISABLE_SHRINK_ON_LOAD`: when `true`, disables shrink-on-load for JPEG and WebP. Allows to process the whole image in linear colorspace but dramatically slows down resizing and increases memory usage when working with large images.
* `IMGPROXY_STRIP_METADATA`: when `true`, imgproxy will strip all metadata (EXIF, IPTC, etc.) from JPEG and WebP output images. Default: `true`.
* `IMGPROXY_STRIP_COLOR_PROFILE`: when `true`, imgproxy will transform the embedded color profile (ICC) to sRGB and remove it from the image. Otherwise, imgproxy will try to keep it as is. Default: `true`.
* `IMGPROXY_AUTO_ROTATE`: when `true`, imgproxy will auto rotate images based on the EXIF Orientation parameter (if available in the image meta data). The orientation tag will be removed from the image anyway. Default: `true`.
