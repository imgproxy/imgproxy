# Configuration

imgproxy is [Twelve-Factor-App](https://12factor.net/)-ready and can be configured using `ENV` variables.

### URL signature

imgproxy allows URLs to be signed with a key and salt. This feature is disabled by default, but it is _highly_ recommended to enable it in production. To enable URL signature checking, define the key/salt pair:

* `IMGPROXY_KEY`: hex-encoded key;
* `IMGPROXY_SALT`: hex-encoded salt;
* `IMGPROXY_SIGNATURE_SIZE`: number of bytes to use for signature before encoding to Base64. Default: 32;

You can specify multiple key/salt pairs by dividing keys and salts with comma (`,`). imgproxy will check URL signatures with each pair. Useful when you need to change key/salt pair in your application with zero downtime.

You can also specify paths to files with a hex-encoded keys and salts, one by line (useful in a development environment):

```bash
$ imgproxy -keypath /path/to/file/with/key -saltpath /path/to/file/with/salt
```

If you need a random key/salt pair real fast, you can quickly generate it using, for example, the following snippet:

```bash
$ echo $(xxd -g 2 -l 64 -p /dev/random | tr -d '\n')
```

### Server

* `IMGPROXY_BIND`: TCP address and port to listen on. Default: `:8080`;
* `IMGPROXY_READ_TIMEOUT`: the maximum duration (in seconds) for reading the entire image request, including the body. Default: `10`;
* `IMGPROXY_WRITE_TIMEOUT`: the maximum duration (in seconds) for writing the response. Default: `10`;
* `IMGPROXY_KEEP_ALIVE_TIMEOUT`: the maximum duration (in seconds) to wait for the next request before closing the connection. When set to `0`, keep-alive is disabled. Default: `10`;
* `IMGPROXY_DOWNLOAD_TIMEOUT`: the maximum duration (in seconds) for downloading the source image. Default: `5`;
* `IMGPROXY_CONCURRENCY`: the maximum number of image requests to be processed simultaneously. Default: number of CPU cores times two;
* `IMGPROXY_MAX_CLIENTS`: the maximum number of simultaneous active connections. Default: `IMGPROXY_CONCURRENCY * 10`;
* `IMGPROXY_TTL`: duration (in seconds) sent in `Expires` and `Cache-Control: max-age` HTTP headers. Default: `3600` (1 hour);
* `IMGPROXY_SO_REUSEPORT`: when `true`, enables `SO_REUSEPORT` socket option (currently on linux and darwin only);
* `IMGPROXY_USER_AGENT`: User-Agent header that will be sent with source image request. Default: `imgproxy/%current_version`;
* `IMGPROXY_USE_ETAG`: when `true`, enables using [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) HTTP header for HTTP cache control. Default: false;

### Security

imgproxy protects you from so-called image bombs. Here is how you can specify maximum image resolution which you consider reasonable:

* `IMGPROXY_MAX_SRC_RESOLUTION`: the maximum resolution of the source image, in megapixels. Images with larger actual size will be rejected. Default: `16.8`;
* `IMGPROXY_MAX_SRC_FILE_SIZE`: the maximum size of the source image, in bytes. Images with larger file size will be rejected. When `0`, file size check is disabled. Default: `0`;

imgproxy can process animated images (GIF, WebP), but since this operation is pretty heavy, only one frame is processed by default. You can increase the maximum of animation frames to process with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum of animated image frames to being processed. Default: `1`.

**Note:** imgproxy summarizes all frames resolutions while checking source image resolution.

You can also specify a secret to enable authorization with the HTTP `Authorization` header for use in production environments:

* `IMGPROXY_SECRET`: the authorization token. If specified, the HTTP request should contain the `Authorization: Bearer %secret%` header;

imgproxy does not send CORS headers by default. Specify allowed origin to enable CORS headers:

* `IMGPROXY_ALLOW_ORIGIN`: when set, enables CORS headers with provided origin. CORS headers are disabled by default.

When you use imgproxy in a development environment, it can be useful to ignore SSL verification:

* `IMGPROXY_IGNORE_SSL_VERIFICATION`: when true, disables SSL verification, so imgproxy can be used in a development environment with self-signed SSL certificates.

Also you may want imgproxy to respond with the same error message that it writes to the log:

* `IMGPROXY_DEVELOPMENT_ERRORS_MODE`: when true, imgproxy will respond with detailed error messages. Not recommended for production because some errors may contain stack trace.

### Compression

* `IMGPROXY_QUALITY`: default quality of the resulting image, percentage. Default: `80`;
* `IMGPROXY_GZIP_COMPRESSION`: GZip compression level. Default: `5`;
* `IMGPROXY_JPEG_PROGRESSIVE` : when true, enables progressive JPEG compression. Default: false;
* `IMGPROXY_PNG_INTERLACED`: when true, enables interlaced PNG compression. Default: false;
* `IMGPROXY_PNG_QUANTIZE`: when true, enables PNG quantization. libvips should be built with libimagequant support. Default: false;
* `IMGPROXY_PNG_QUANTIZATION_COLORS`: maximum number of quantization palette entries. Should be between 2 and 256. Default: 256;

## WebP support detection

imgproxy can use the `Accept` HTTP header to detect if the browser supports WebP and use it as the default format. This feature is disabled by default and can be enabled by the following options:

* `IMGPROXY_ENABLE_WEBP_DETECTION`: enables WebP support detection. When the file extension is omitted in the imgproxy URL and browser supports WebP, imgproxy will use it as the resulting format;
* `IMGPROXY_ENFORCE_WEBP`: enables WebP support detection and enforces WebP usage. If the browser supports WebP, it will be used as resulting format even if another extension is specified in the imgproxy URL.

When WebP support detection is enabled, please take care to configure your CDN or caching proxy to take the `Accept` HTTP header into account while caching.

**Warning**: Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Accept` HTTP headers. Have this in mind when configuring your production caching setup.

## Client Hints support

imgproxy can use the `Width`, `Viewport-Width` or `DPR` HTTP headers to determine default width and DPR options using Client Hints. This feature is disabled by default and can be enabled by the following option:

* `IMGPROXY_ENABLE_CLIENT_HINTS`: enables Client Hints support to determine default width and DPR options. Read [here](https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints) details about Client Hints.

**Warning**: Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Width`, `Viewport-Width` or `DPR` HTTP headers. Have this in mind when configuring your production caching setup.

### Watermark

* `IMGPROXY_WATERMARK_DATA`: Base64-encoded image data. You can easily calculate it with `base64 tmp/watermark.png | tr -d '\n'`;
* `IMGPROXY_WATERMARK_PATH`: path to the locally stored image;
* `IMGPROXY_WATERMARK_URL`: watermark image URL;
* `IMGPROXY_WATERMARK_OPACITY`: watermark base opacity.

Read more about watermarks in the [Watermark](./watermark.md) guide.

### Presets

Read about imgproxy presets in the [Presets](./presets.md) guide.

There are two ways to define presets:

##### Using an environment variable

* `IMGPROXY_PRESETS`: set of preset definitions, comma-divided. Example: `default=resizing_type:fill/enlarge:1,sharp=sharpen:0.7,blurry=blur:2`. Default: blank.

##### Using a command line argument

```bash
$ imgproxy -presets /path/to/file/with/presets
```

The file should contain preset definitions, one per line. Lines starting with `#` are treated as comments. Example:

```
default=resizing_type:fill/enlarge:1

# Sharpen the image to make it look better
sharp=sharpen:0.7

# Blur the image to hide details
blurry=blur:2
```

#### Using only presets

imgproxy can be switched into "presets-only mode". In this mode, imgproxy accepts only `preset` option arguments as processing options. Example: `http://imgproxy.example.com/unsafe/thumbnail:blurry:watermarked/plain/http://example.com/images/curiosity.jpg@png`

* `IMGPROXY_ONLY_PRESETS`: disable all URL formats and enable presets-only mode.

### Serving local files

imgproxy can serve your local images, but this feature is disabled by default. To enable it, specify your local filesystem root:

* `IMGPROXY_LOCAL_FILESYSTEM_ROOT`: the root of the local filesystem. Keep empty to disable serving of local files.

Check out the [Serving local files](./serving_local_files.md) guide to learn more.

### Serving files from Amazon S3

imgproxy can process files from Amazon S3 buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_S3` to `true`:

* `IMGPROXY_USE_S3`: when `true`, enables image fetching from Amazon S3 buckets. Default: false;
* `IMGPROXY_S3_ENDPOINT`: custom S3 endpoint to being used by imgproxy.

Check out the [Serving files from S3](./serving_files_from_s3.md) guide to learn more.

### Serving files from Google Cloud Storage

imgproxy can process files from Google Cloud Storage buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_GCS_KEY` to the content of Google Cloud JSON key:

* `IMGPROXY_GCS_KEY`: Google Cloud JSON key. When set, enables image fetching from Google Cloud Storage buckets. Default: blank.

Check out the [Serving files from Google Cloud Storage](./serving_files_from_google_cloud_storage.md) guide to learn more.

### New Relic metrics

imgproxy can send its metrics to New Relic. Specify your New Relic license key to activate this feature:

* `IMGPROXY_NEW_RELIC_KEY`: New Relic license key;
* `IMGPROXY_NEW_RELIC_APP_NAME`: New Relic application name. Default: `imgproxy`.

Check out the [New Relic](./new_relic.md) guide to learn more.

### Prometheus metrics

imgproxy can collect its metrics for Prometheus. Specify binding for Prometheus metrics server to activate this feature:

* `IMGPROXY_PROMETHEUS_BIND`: Prometheus metrics server binding. Can't be the same as `IMGPROXY_BIND`. Default: blank.

Check out the [Prometheus](./prometheus.md) guide to learn more.

### Error reporting

imgproxy can report occurred errors to Bugsnag, Honeybadger and Sentry:

* `IMGPROXY_BUGSNAG_KEY`: Bugsnag API key. When provided, enables error reporting to Bugsnag;
* `IMGPROXY_BUGSNAG_STAGE`: Bugsnag stage to report to. Default: `production`;
* `IMGPROXY_HONEYBADGER_KEY`: Honeybadger API key. When provided, enables error reporting to Honeybadger;
* `IMGPROXY_HONEYBADGER_ENV`: Honeybadger env to report to. Default: `production`.
* `IMGPROXY_SENTRY_DSN`: Sentry project DSN. When provided, enables error reporting to Sentry;
* `IMGPROXY_SENTRY_ENVIRONMENT`: Sentry environment to report to. Default: `production`.
* `IMGPROXY_SENTRY_RELEASE`: Sentry release to report to. Default: `imgproxy/{imgproxy version}`.

### Syslog

imgproxy can send logs to syslog, but this feature is disabled by default. To enable it, set `IMGPROXY_SYSLOG_ENABLE` to `true`:

* `IMGPROXY_SYSLOG_ENABLE`: when `true`, enables sending logs to syslog;
* `IMGPROXY_SYSLOG_LEVEL`: maximum log level to send to syslog. Known levels are: `crit`, `error`, `warning` and `notice`. Default: `notice`;
* `IMGPROXY_SYSLOG_NETWORK`: network that will be used to connect to syslog. When blank, the local syslog server will be used. Known networks are `tcp`, `tcp4`, `tcp6`, `udp`, `udp4`, `udp6`, `ip`, `ip4`, `ip6`, `unix`, `unixgram` and `unixpacket`. Default: blank;
* `IMGPROXY_SYSLOG_ADDRESS`: address of the syslog service. Not used if `IMGPROXY_SYSLOG_NETWORK` is blank. Default: blank;
* `IMGPROXY_SYSLOG_TAG`: specific syslogtag. Default: `imgproxy`;

### Memory usage tweaks

**Warning:** It's highly recommended to read [Memory usage tweaks](./memory_usage_tweaks.md) guide before changing this settings.

* `IMGPROXY_DOWNLOAD_BUFFER_SIZE`: the initial size (in bytes) of a single download buffer. When zero, initializes empty download buffers. Default: `0`;
* `IMGPROXY_GZIP_BUFFER_SIZE`: the initial size (in bytes) of a single GZip buffer. When zero, initializes empty GZip buffers. Makes sense only when GZip compression is enabled. Default: `0`;
* `IMGPROXY_FREE_MEMORY_INTERVAL`: the interval (in seconds) at which unused memory will be returned to the OS. Default: `10`;
* `IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD`: the number of buffers that should be returned to a pool before calibration. Default: `1024`.

### Miscellaneous

* `IMGPROXY_BASE_URL`: base URL prefix that will be added to every requested image URL. For example, if the base URL is `http://example.com/images` and `/path/to/image.png` is requested, imgproxy will download the source image from `http://example.com/images/path/to/image.png`. Default: blank.
* `IMGPROXY_USE_LINEAR_COLORSPACE`: when `true`, imgproxy will process images in linear colorspace. This will slow down processing. Note that images won't be fully processed in linear colorspace while shrink-on-load is enabled (see below).
* `IMGPROXY_DISABLE_SHRINK_ON_LOAD`: when `true`, disables shrink-on-load for JPEG and WebP. Allows to process the whole image in linear colorspace but dramatically slows down resizing and increases memory usage when working with large images.
