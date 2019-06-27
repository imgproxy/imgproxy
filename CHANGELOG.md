# Changelog

## master

- Better handling if non-sRGB images;
- `SO_REUSEPORT` socker option support. Can be enabled with `IMGPROXY_SO_REUSEPORT`;
- `dpr` option always changes the resulting size even if it leads to enlarge and `enlarge` is falsey;

## v2.3.0

- `libvips` v8.8 support: better processing of animated GIFs, built-in CMYK profile, better WebP scale-on-load, etc;
- Animated WebP support. `IMGPROXY_MAX_GIF_FRAMES` is deprecated, use `IMGPROXY_MAX_ANIMATION_FRAMES`;
- [HEIC support](./docs/image_formats_support.md#heic-support);
- [crop](./docs/generating_the_url_advanced.md#crop) processing option. `resizing_type:crop` is deprecated;
- Offsets for [gravity](./docs/generating_the_url_advanced.md#gravity);
- Resizing type `auto`. If both source and resulting dimensions have the same orientation (portrait or landscape), imgproxy will use `fill`. Otherwise, it will use `fit`;
- Development errors mode. When `IMGPROXY_DEVELOPMENT_ERRORS_MODE` is true, imgproxy will respond with detailed error messages. Not recommended for production because some errors may contain stack trace;
- Better stack trace for image processing errors;
- Allowed URL query for `/health`;
- `IMGPROXY_KEEP_ALIVE_TIMEOUT` config.

## v2.2.13

- Better shrink-on-load;
- Don't import common sRGB IEC61966-2.1 ICC profile unless linear colorspace is used;
- Send `X-Reqiest-ID` header;
- Don't fail on recursive preset usage, just ignore already used preset and log warning.

## v2.2.12

- Don't fail processing when embedded ICC profile is not compatible with the image.

## v2.2.11

- Optimized ICC import when linear colorspace usage is disabled.

## v2.2.10

- Resizing images in linear colorspace is disabled by default. Can be enabled with `IMGPROXY_USE_LINEAR_COLORSPACE`;
- Add PNG quantization. Can be enabled with `IMGPROXY_PNG_QUANTIZE`. Palette size can be specified with `IMGPROXY_PNG_QUANTIZATION_COLORS`.

## v2.2.9

Fixed processing of images with embedded profiles that was broken in v2.2.8.

## v2.2.8

- Resize images in linear colorspace;
- Add `IMGPROXY_DISABLE_SHRINK_ON_LOAD` config to disable shring-on-load of JPEG and WebP;
- Remove orc from Docker image (causes segfaults in some cases).

## v2.2.7

- Fixed color management;
- Memory usage optimizations.

## v2.2.6

- Fixed signature check when source URL is escaped.

## v2.2.5

- [extend](./docs/generating_the_url_advanced.md#extend) processing option;
- Fixed SVG detection;
- Add `vips_memory_bytes`, `vips_max_memory_bytes` and `vips_allocs` metrics to Prometheus.

## v2.2.4

- Minor improvements.

## v2.2.3

- Fixed critical bug with cached C strings;
- Simple filesystem transport withh less memory usage.

## v2.2.2

- Memory usage optimizations.

## v2.2.1

- Source file size limit;
- More memory usage optimizations.

## v2.2.0

- Optimized memory usage. [Memory usage tweaks](./docs/memory_usage_tweaks.md);
- `Vary` header is set when WebP detection, client hints or GZip compression are enabled;
- Health check doesn't require `Authorization` header anymore.

## v2.1.5

- [Sentry support](./docs/configuration.md#error-reporting) (thanks to [@koenpunt](https://github.com/koenpunt));
- Fixed detection of some kind of WebP images;
- [Syslog support](./docs/configuration.md#syslog).

## v2.1.4

- SVG sources support;
- Fixed support for not animated GIFs;
- Proper filename in the `Content-Disposition` header;
- Memory usage optimizations.

## v2.1.3

- [Minio support](./docs/serving_files_from_s3.md#minio)

## v2.1.2

- ICO support

## v2.1.1

- Fixed EXIF orientation fetching;
- When libvips failed to save PNG, imgproxy will try to save is without embedded ICC profile.

## v2.1.0

- [Plain source URLs](./docs/generating_the_url_advanced.md#plain) support;
- [Serving images from Google Cloud Storage](./docs/serving_files_from_google_cloud_storage.md);
- [Full support of GIFs](./docs/image_formats_support.md#gif-support) including animated ones;
- [Watermarks](./docs/watermark.md);
- [New Relic](./docs/new_relic.md) metrics;
- [Prometheus](./docs/prometheus.md) metrics;
- [DPR](./docs/generating_the_url_advanced.md#dpr) option (thanks to [selul](https://github.com/selul));
- [Cache buster](./docs/generating_the_url_advanced.md#cache-buster) option;
- [Quality](./docs/generating_the_url_advanced.md#quality) option;
- Support for custom [Amazon S3](./docs/serving_files_from_s3.md) endpoints;
- Support for [Amazon S3](./docs/serving_files_from_s3.md) versioning;
- [Client hints](./docs/configuration.md#client-hints-support) support (thanks to [selul](https://github.com/selul));
- Using source image format when one is not specified in the URL;
- Sending `User-Agent` header when downloading a source image;
- Setting proper filename in `Content-Disposition` header in the response;
- Truncated signature support (thanks to [printercu](https://github.com/printercu));
- imgproxy uses source image format by default for the resulting image;
- `IMGPROXY_MAX_SRC_DIMENSION` is **deprecated**, use `IMGPROXY_MAX_SRC_RESOLUTION` instead.

## v2.0.3

Fixed URL validation when IMGPROXY_BASE_URL is used

## v2.0.2

Fixed smart crop + blur/sharpen SIGSEGV on Alpine

## v2.0.1

Minor fixes

## v2.0.0

All-You-Ever-Wanted release! :tada:

- Key and salt are not required anymore. When key or salt is not specified, signature checking is disabled;
- [New advanced URL format](./docs/generating_the_url_advanced.md). Unleash the full power of imgproxy v2.0;
- [Presets](./docs/presets.md). Shorten your urls by reusing processing options;
- [Serving images from Amazon S3](./docs/serving_files_from_s3.md). Thanks to [@crohr](https://github.com/crohr), now we have a way to serve files from private S3 buckets;
- [Autoconverting to WebP when supported by browser](./docs/configuration.md#webp-support-detection) (disabled by default). Use WebP as resulting format when browser supports it;
- [Gaussian blur](./docs/generating_the_url_advanced.md#blur) and [sharpen](./docs/generating_the_url_advanced.md#sharpen) filters. Make your images look better than before;
- [Focus point gravity](./docs/generating_the_url_advanced.md#gravity). Tell imgproxy what point will be the center of the image;
- [Background color](./docs/generating_the_url_advanced.md#background). Control the color of background when converting PNG with alpha-channel to JPEG;
- Imgproxy calcs resulting width/height automaticly when one specified as zero;
- Memory usage is optimized.

## v1.1.8

- Disabled libvips cache to prevent SIGSEGV on Alpine

## v1.1.7

- Improved ETag generation

## v1.1.6

- Added progressive JPEG and interlaced PNG support

## v1.1.5.1

- Fixed autorotation when image is not resized

## v1.1.5

- Add CORS headers
- Add IMGPROXY_BASE_URL config
- Add Content-Length header

## v1.1.4

- Added request ID
- Idle time does not causes timeout
- Increased default maximum number of simultaneous active connections
