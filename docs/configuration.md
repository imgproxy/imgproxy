# Configuration

imgproxy is [Twelve-Factor-App](https://12factor.net/)-ready and can be configured using `ENV` variables.

### URL signature

imgproxy allows URLs to be signed with a key and salt. This feature is disabled by default, but it is _highly_ recommended to enable it in production. To enable URL signature checking, define the key/salt pair:

* `IMGPROXY_KEY`: hex-encoded key;
* `IMGPROXY_SALT`: hex-encoded salt;

You can also specify paths to files with a hex-encoded key and salt (useful in a development environment):

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
* `IMGPROXY_DOWNLOAD_TIMEOUT`: the maximum duration (in seconds) for downloading the source image. Default: `5`;
* `IMGPROXY_CONCURRENCY`: the maximum number of image requests to be processed simultaneously. Default: number of CPU cores times two;
* `IMGPROXY_MAX_CLIENTS`: the maximum number of simultaneous active connections. Default: `IMGPROXY_CONCURRENCY * 10`;
* `IMGPROXY_TTL`: duration (in seconds) sent in `Expires` and `Cache-Control: max-age` HTTP headers. Default: `3600` (1 hour);
* `IMGPROXY_USER_AGENT`: User-Agent header that will be sent with source image request. Default: `imgproxy/%current_version`;
* `IMGPROXY_USE_ETAG`: when `true`, enables using [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) HTTP header for HTTP cache control. Default: false;

### Security

imgproxy protects you from so-called image bombs. Here is how you can specify maximum image dimensions and resolution which you consider reasonable:

* `IMGPROXY_MAX_SRC_DIMENSION`: the maximum dimensions of the source image, in pixels, for both width and height. Images with larger actual size will be rejected. Default: `8192`;
* `IMGPROXY_MAX_SRC_RESOLUTION`: the maximum resolution of the source image, in megapixels. Images with larger actual size will be rejected. Default: `16.8`;

You can also specify a secret to enable authorization with the HTTP `Authorization` header for use in production environments:

* `IMGPROXY_SECRET`: the authorization token. If specified, the HTTP request should contain the `Authorization: Bearer %secret%` header;

imgproxy does not send CORS headers by default. Specify allowed origin to enable CORS headers:

* `IMGPROXY_ALLOW_ORIGIN`: when set, enables CORS headers with provided origin. CORS headers are disabled by default.

When you use imgproxy in a development environment, it can be useful to ignore SSL verification:

* `IMGPROXY_IGNORE_SSL_VERIFICATION`: when true, disables SSL verification, so imgproxy can be used in a development environment with self-signed SSL certificates.

### Compression

* `IMGPROXY_QUALITY`: default quality of the resulting image, percentage. Default: `80`;
* `IMGPROXY_GZIP_COMPRESSION`: GZip compression level. Default: `5`;
* `IMGPROXY_JPEG_PROGRESSIVE` : when true, enables progressive JPEG compression. Default: false;
* `IMGPROXY_PNG_INTERLACED`: when true, enables interlaced PNG compression. Default: false;

## WebP support detection

imgproxy can use the `Accept` HTTP header to detect if the browser supports WebP and use it as the default format. This feature is disabled by default and can be enabled by the following options:

* `IMGPROXY_ENABLE_WEBP_DETECTION`: enables WebP support detection. When the file extension is omitted in the imgproxy URL and browser supports WebP, imgproxy will use it as the resulting format;
* `IMGPROXY_ENFORCE_WEBP`: enables WebP support detection and enforces WebP usage. If the browser supports WebP, it will be used as resulting format even if another extension is specified in the imgproxy URL.

When WebP support detection is enabled, please take care to configure your CDN or caching proxy to take the `Accept` HTTP header into account while caching.

**Warning**: Headers cannot be signed. This means that an attacker can bypass your CDN cache by changing the `Accept` HTTP header. Have this in mind when configuring your production caching setup.

### Watermark

* `IMGPROXY_WATERMARK_DATA` - Base64-encoded image data. You can easily calculate it with `base64 tmp/watermark.png | tr -d '\n'`;
* `IMGPROXY_WATERMARK_PATH` - path to the locally stored image;
* `IMGPROXY_WATERMARK_URL` - watermark image URL;
* `IMGPROXY_WATERMARK_OPACITY` - watermark base opacity.

Read more about watermarks in the [Watermark](./watermark.md) guide.

### Presets

Read about imgproxy presets in the [Presets](./presets.md) guide.

There are two ways to define presets:

##### Using an environment variable

* `IMGPROXY_PRESETS`: set of preset definitions, comma-divided. Example: `default=resize_type:fill/enlarge:1,sharp=sharpen:0.7,blurry=blur:2`. Default: blank.

##### Using a command line argument

```bash
$ imgproxy -presets /path/to/file/with/presets
```

The file should contain preset definitions, one per line. Lines starting with `#` are treated as comments. Example:

```
default=resize_type:fill/enlarge:1

# Sharpen the image to make it look better
sharp=sharpen:0.7

# Blur the image to hide details
blurry=blur:2
```

### Serving local files

imgproxy can serve your local images, but this feature is disabled by default. To enable it, specify your local filesystem root:

* `IMGPROXY_LOCAL_FILESYSTEM_ROOT`: the root of the local filesystem. Keep empty to disable serving of local files.

Check out the [Serving local files](./serving_local_files.md) guide to learn more.

### Serving files from Amazon S3

imgproxy can process files from Amazon S3 buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_USE_S3` to `true`:

* `IMGPROXY_USE_S3`: when `true`, enables image fetching from Amazon S3 buckets. Default: false.

Check out the [Serving files from S3](./serving_files_from_s3.md) guide to learn more.

### Serving files from Google Cloud Storage

imgproxy can process files from Google Cloud Storage buckets, but this feature is disabled by default. To enable it, set `IMGPROXY_GCS_KEY` to the content of Google Cloud JSON key:

* `IMGPROXY_GCS_KEY`: Google Cloud JSON key. When set, enables image fetching from Google Cloud Storage buckets. Default: blank.

Check out the [Serving files from Google Cloud Storage](./serving_files_from_google_cloud_storage.md) guide to learn more.

### New Relic metrics

imgproxy can send its metrics to New Relic. Specify your New Relic license key to activate this feature:

* `IMGPROXY_NEW_RELIC_KEY` - New Relic license key;
* `IMGPROXY_NEW_RELIC_APP_NAME` - application name. If not specified, `imgproxy` will be used as the application name.

Check out the [New Relic](./new_relic.md) guide to learn more.

### Prometheus metrics

imgproxy can collect its metrics for Prometheus. Specify binding for Prometheus metrics server to activate this feature:

* `IMGPROXY_PROMETHEUS_BIND`: prometheus metrics server binding. Can't be the same as `IMGPROXY_BIND`. Default: blank.

Check out the [Prometheus](./prometheus.md) guide to learn more.

### Miscellaneous

* `IMGPROXY_BASE_URL`: base URL prefix that will be added to every requested image URL. For example, if the base URL is `http://example.com/images` and `/path/to/image.png` is requested, imgproxy will download the source image from `http://example.com/images/path/to/image.png`. Default: blank.
