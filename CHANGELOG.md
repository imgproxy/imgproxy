# Changelog

## [Unreleased]
### Added
- TIFF and BMP support.
- `IMGPROXY_REPORT_DOWNLOADING_ERRORS` config. Setting it to `false` disables reporting of downloading errors.
- SVG passthrough. When source image and requested format are SVG, image will be returned without changes.
- `IMGPROXY_USE_GCS` config. When it set to true and `IMGPROXY_GCS_KEY` is not set, imgproxy tries to use Application Default Credentials to get access to GCS bucket.

### Changed
- Reimplemented and more errors-tolerant image size parsing;
- Log only modified processing options;

### Fixed
- Fixed sharpening+watermarking;
- Fixed path parsing when no options is provided and image URL is Base64 encoded.

### Deprecated

- Using `IMGPROXY_GCS_KEY` without `IMGPROXY_USE_GCS` set to `true` is deprecated.

## [2.5.0] - 2019-09-19
### Added
- `structured` and `json` log formats. Can be set with `IMGPROXY_LOG_FORMAT`.

### Changed
- New default log format.
- Better watermarking: image transparency doesn't affect watermarks, faster watermark scaling.

## [2.4.1] - 2019-08-29
### Changed
- More verbose URL parsing errors.

## [2.4.0] - 2019-08-20
### Added
- `SO_REUSEPORT` socker option support. Can be enabled with `IMGPROXY_SO_REUSEPORT`.
- [filename](./docs/generating_the_url_advanced.md#filename) option.

### Changed
- Better handling if non-sRGB images.
- `dpr` option always changes the resulting size even if it leads to enlarge and `enlarge` is falsey.
- Log to STDOUT.
- Only unexpected errors are reported to Bugsnag/Honeybadger/Sentry.
- Better Sentry support.

### Deprecated
- GZip compression support is deprecated.

## [2.3.0] - 2019-06-25
### Added
- `libvips` v8.8 support: better processing of animated GIFs, built-in CMYK profile, better WebP scale-on-load, etc;
- Animated WebP support. `IMGPROXY_MAX_GIF_FRAMES` is deprecated, use `IMGPROXY_MAX_ANIMATION_FRAMES`;
- [HEIC support](./docs/image_formats_support.md#heic-support);
- [crop](./docs/generating_the_url_advanced.md#crop) processing option. `resizing_type:crop` is deprecated;
- Offsets for [gravity](./docs/generating_the_url_advanced.md#gravity);
- Resizing type `auto`. If both source and resulting dimensions have the same orientation (portrait or landscape), imgproxy will use `fill`. Otherwise, it will use `fit`;
- Development errors mode. When `IMGPROXY_DEVELOPMENT_ERRORS_MODE` is true, imgproxy will respond with detailed error messages. Not recommended for production because some errors may contain stack trace;
- `IMGPROXY_KEEP_ALIVE_TIMEOUT` config.

### Changed
- Allow URL query for `/health`;
- Better stack trace for image processing errors;

## [2.2.13] - 2019-05-07
### Added
- Send `X-Request-ID` header in response.

### Changed
- Better shrink-on-load.
- Don't import common sRGB IEC61966-2.1 ICC profile unless linear colorspace is used.
- Don't fail on recursive preset usage, just ignore already used preset and log warning.

## [2.2.12] - 2019-04-11
### Changed
- Don't fail processing when embedded ICC profile is not compatible with the image.

## [2.2.11] - 2019-04-08
### Changed
- Optimized ICC import when linear colorspace usage is disabled.

## [2.2.10] - 2019-04-05
### Added
- PNG quantization. Can be enabled with `IMGPROXY_PNG_QUANTIZE`. Palette size can be specified with `IMGPROXY_PNG_QUANTIZATION_COLORS`.

### Changed
- Resizing images in linear colorspace is disabled by default. Can be enabled with `IMGPROXY_USE_LINEAR_COLORSPACE`;

## [2.2.9] - 2019-04-02
### Fixed
Fixed processing of images with embedded profiles that was broken in v2.2.8.

## [2.2.8] - 2019-04-01
### Added
- Resizing in linear colorspace;
- `IMGPROXY_DISABLE_SHRINK_ON_LOAD` config to disable shring-on-load of JPEG and WebP.

### Fixed
- Remove orc from Docker image (causes segfaults in some cases).

## [2.2.7] - 2019-03-22
### Changed
- Memory usage optimizations.

### Fixed
- Fix color management.

## [2.2.6] - 2019-02-27
### Fixed
- Fix signature check when source URL is escaped.

## [2.2.5] - 2019-02-21
### Added
- [extend](./docs/generating_the_url_advanced.md#extend) processing option.
- `vips_memory_bytes`, `vips_max_memory_bytes` and `vips_allocs` metrics for Prometheus.

### Fixed
- Fix SVG detection.

## [2.2.4] - 2019-02-13
### Changed
- Minor improvements.

## [2.2.3] - 2019-02-04
### Changed
- Simple filesystem transport withh less memory usage.

### Fixed
- Fix critical bug with cached C strings;

## [2.2.2] - 2019-02-01

- Memory usage optimizations.

## [2.2.1] - 2019-01-21
### Added
- Source file size limit.

### Changed
- More memory usage optimizations.

## [2.2.0] - 2019-01-19
### Changed
- Optimized memory usage. [Memory usage tweaks](./docs/memory_usage_tweaks.md).
- `Vary` header is set when WebP detection, client hints or GZip compression are enabled.
- Health check doesn't require `Authorization` header anymore.

## [2.1.5] - 2019-01-14
### Added
- [Sentry support](./docs/configuration.md#error-reporting) (thanks to [@koenpunt](https://github.com/koenpunt)).
- [Syslog support](./docs/configuration.md#syslog).

### Fixed
- Fix detection of some kind of WebP images;

## [2.1.4] - 2019-01-10
### Added
- SVG sources support.

### Changed
- Memory usage optimizations.
- Proper filename in the `Content-Disposition` header.

### Fixed
- Fix support for not animated GIFs.

## [2.1.3] - 2018-12-10
### Added
- [Minio support](./docs/serving_files_from_s3.md#minio)

## [2.1.2] - 2018-12-02
### Added
- ICO support

## [2.1.1] - 2018-11-29
### Changed
- When libvips failed to save PNG, imgproxy will try to save is without embedded ICC profile.

### Fixed
- Fixed EXIF orientation fetching.

## [2.1.0] - 2018-11-16
### Added
- [Plain source URLs](./docs/generating_the_url_advanced.md#plain) support.
- [Serving images from Google Cloud Storage](./docs/serving_files_from_google_cloud_storage.md).
- [Full support of GIFs](./docs/image_formats_support.md#gif-support) including animated ones.
- [Watermarks](./docs/watermark.md).
- [New Relic](./docs/new_relic.md) metrics.
- [Prometheus](./docs/prometheus.md) metrics.
- [DPR](./docs/generating_the_url_advanced.md#dpr) option (thanks to [selul](https://github.com/selul)).
- [Cache buster](./docs/generating_the_url_advanced.md#cache-buster) option.
- [Quality](./docs/generating_the_url_advanced.md#quality) option.
- Support for custom [Amazon S3](./docs/serving_files_from_s3.md) endpoints.
- Support for [Amazon S3](./docs/serving_files_from_s3.md) versioning.
- [Client hints](./docs/configuration.md#client-hints-support) support (thanks to [selul](https://github.com/selul)).
- Truncated signature support (thanks to [printercu](https://github.com/printercu)).

### Changed
- imgproxy uses source image format by default for the resulting image.
- Send `User-Agent` header when downloading a source image.
- Proper filename in `Content-Disposition` header in the response.

### Deprecated
- `IMGPROXY_MAX_SRC_DIMENSION` is **deprecated**, use `IMGPROXY_MAX_SRC_RESOLUTION` instead.

## [2.0.3] - 2018-11-02
### Fixed
- Fix URL validation when IMGPROXY_BASE_URL is used.

## [2.0.2] - 2018-10-25
### Fixed
- Fix smart crop + blur/sharpen SIGSEGV on Alpine.

## [2.0.1] - 2018-10-18
### Fixed
- Minor fixes.

## [2.0.0] - 2018-10-08
All-You-Ever-Wanted release! :tada:
### Added
- [New advanced URL format](./docs/generating_the_url_advanced.md). Unleash the full power of imgproxy v2.0.
- [Presets](./docs/presets.md). Shorten your urls by reusing processing options.
- [Serving images from Amazon S3](./docs/serving_files_from_s3.md). Thanks to [@crohr](https://github.com/crohr), now we have a way to serve files from private S3 buckets.
- [Autoconverting to WebP when supported by browser](./docs/configuration.md#webp-support-detection) (disabled by default). Use WebP as resulting format when browser supports it.
- [Gaussian blur](./docs/generating_the_url_advanced.md#blur) and [sharpen](./docs/generating_the_url_advanced.md#sharpen) filters. Make your images look better than before.
- [Focus point gravity](./docs/generating_the_url_advanced.md#gravity). Tell imgproxy what point will be the center of the image.
- [Background color](./docs/generating_the_url_advanced.md#background). Control the color of background when converting PNG with alpha-channel to JPEG.

### Changed
- Key and salt are not required anymore. When key or salt is not specified, signature checking is disabled.
- Imgproxy calcs resulting width/height automaticly when one specified as zero.
- Memory usage is optimized.

## [1.1.8] - 2018-10-01
### Fixed
- Disable libvips cache to prevent SIGSEGV on Alpine.

## [1.1.7] - 2018-09-06
### Changed
- Improved ETag generation.

## [1.1.6] - 2018-07-26
### Added
- Progressive JPEG and interlaced PNG support.

## [1.1.5.1] - 2018-05-25
### Fixed
- Fix autorotation when image is not resized.

## [1.1.5] - 2018-04-27
### Added
- CORS headers.
- `IMGPROXY_BASE_URL` config.
- `Content-Length` header.

## [1.1.4] - 2018-03-19
### Added
- Request ID in the logs.

### Changed
- Idle time does not causes timeout.
- Increased default maximum number of simultaneous active connections.
