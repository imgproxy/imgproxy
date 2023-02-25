# Changelog

## [Unreleased]
## Add
- Add [extend_aspect_ratio](https://docs.imgproxy.net/latest/generating_the_url?id=extend-aspect-ratio) processing option.
- Add the `IMGPROXY_ALLOW_SECURITY_OPTIONS` config + `max_src_resolution`, `max_src_file_size`, `max_animation_frames`, and `max_animation_frame_resolution` processing options.
- (pro) Add [advanced smart crop](https://docs.imgproxy.net/latest/configuration?id=smart-crop).

### Change
- Make the `expires` processing option set `Expires` and `Cache-Control` headers.
- Sanitize `use` tags in SVGs.

## [3.13.2] - 2023-02-15
### Change
- Remove color-related EXIF data when stripping ICC profile.
- (pro) Optimize saving to MP4.

### Fix
- (pro) Fix saving with autoquality in some cases.
- (pro) Fix saving large images to MP4.

## [3.13.1] - 2023-01-16
### Fix
- Fix applying watermarks with replication.

## [3.13.0] - 2023-01-11
### Change
- Add support for Managed Identity or Service Principal credentials to Azure Blob Storage integration.
- Optimize memory usage in some scenarios.
- Better SVG sanitization.
- (pro) Allow usage of floating-point numbers in the `IMGPROXY_VIDEO_THUMBNAIL_SECOND` config and the `video_thumbnail_second` processing option.

### Fix
- Fix craches in some cases when using OpenTelemetry in Amazon ECS.
- (pro) Fix saving of GIF with too small frame delay to MP4

## [3.12.0] - 2022-12-11
### Add
- Add `IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION` config.
- Add [Amazon CloudWatch](https://docs.imgproxy.net/latest/cloud_watch) support.
- (pro) Add [`best` resultig image format](https://docs.imgproxy.net/latest/best_format).
- (pro) Add `IMGPROXY_WEBP_COMPRESSION` config and [webp_options](https://docs.imgproxy.net/latest/generating_the_url?id=webp-options) processing option.

### Change
- Change `IMGPROXY_FORMAT_QUALITY` default value to `avif=65`.
- Change `IMGPROXY_AVIF_SPEED` default value to `8`.
- Change `IMGPROXY_PREFERRED_FORMATS` default value to `jpeg,png,gif`.
- Set `Cache-Control: no-cache` header to the health check responses.
- Allow replacing line breaks with `\n` in `IMGPROXY_OPEN_TELEMETRY_SERVER_CERT`, `IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT`, and`IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY`.

### Fix
- Fix 3GP video format detection.

## [3.11.0] - 2022-11-17
### Add
- Add `IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE` config.
- Add `IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR` config.
- (pro) Add XMP data to the `/info` response.

### Change
- Better XMP data stripping.
- Use parent-based OpenTelemetry sampler by default.
- Use fixed TraceIdRatioBased sampler with OpenTelemetry.

## [3.10.0] - 2022-11-04
### Add
- Add `IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT` config.
- (pro) Add [disable_animation](https://docs.imgproxy.net/latest/generating_the_url?id=disable-animation) processing option.
- (pro) Add [gradient](https://docs.imgproxy.net/latest/generating_the_url?id=gradient) processing option.

### Fix
- Fix false-positive SVG detections.
- Fix possible infinite loop during SVG sanitization.
- (pro) Fix saving of GIF with variable frame delay to MP4.
- (pro) Fix the size of video thumbnails if the video has a defined sample aspect ratio.

## [3.9.0] - 2022-10-19
### Add
- Add `IMGPROXY_SVG_FIX_UNSUPPORTED` config.

### Fix
- Fix HTTP response status when OpenTelemetry support is enabled.
- (docker) Fix saving of paletted PNGs with low bit-depth.

## [3.8.0] - 2022-10-06
### Add
- Add [raw](https://docs.imgproxy.net/latest/generating_the_url?id=raw) processing option.
- Add [OpenTelemetry](https://docs.imgproxy.net/latest/open_telemetry) support.
- (pro) Add encrypted source URL support.
- (pro) Add [watermark_shadow](https://docs.imgproxy.net/generating_the_url?id=watermark-shadow) processing option.

### Changed
- Try to fix some invalid source URL cases that happen because of URL normalization.

## [3.7.2] - 2022-08-22
### Changed
- (docker) Faster images quantization.
- (docker) Faster loading of GIF.

## [3.7.1] - 2022-08-01
### Fix
- Fix memory bloat in some cases.
- Fix `format_quality` usage in presets.

## [3.7.0] - 2022-07-27
### Add
- Add support of 16-bit BMP.
- Add `IMGPROXY_NEW_RELIC_LABELS` config.
- Add support of JPEG files with differential Huffman coding or arithmetic coding.
- Add `IMGPROXY_PREFERRED_FORMATS` config.
- Add `IMGPROXY_REQUESTS_QUEUE_SIZE` config.
- Add `requests_in_progress` and `images_in_progress` metrics.
- Add queue segment/span to request traces.
- Add sending additional metrics to Datadog and `IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS` config.
- Add sending additional metrics to New Relic.

### Change
- Change `IMGPROXY_MAX_CLIENTS` default value to 2048.
- Allow unlimited connections when `IMGPROXY_MAX_CLIENTS` is set to `0`.
- Change `IMGPROXY_TTL` default value to `31536000` (1 year).
- Better errors tracking with metrics services.
- (docker) Faster and better saving of GIF.
- (docker) Faster saving of AVIF.
- (docker) Faster loading and saving of PNG.

### Fix
- Fix trimming of CMYK images.
- Respond with 404 when the source image can not be found in OpenStack Object Storage.
- Respond with 404 when file wasn't found in the GCS storage.

## [3.6.0] - 2022-06-13
### Add
- Add `IMGPROXY_RETURN_ATTACHMENT` config and [return_attachment](https://docs.imgproxy.net/generating_the_url?return-attachment) processing option.
- Add SVG sanitization and `IMGPROXY_SANITIZE_SVG` config.

### Change
- Better animation detection.

### Fix
- Respond with 404 when file wasn't found in the local storage.

## [3.5.1] - 2022-05-20
### Change
- Fallback from AVIF to JPEG/PNG if one of the result dimensions is smaller than 16px.

### Fix
- (pro) Fix some PDF pages background.
- (docker) Fix loading some HEIF images.

## [3.5.0] - 2022-04-25
### Add
- Add support of RLE-encoded BMP.
- Add `IMGPROXY_ENFORCE_THUMBNAIL` config and [enforce_thumbnail](https://docs.imgproxy.net/generating_the_url?id=enforce-thumbnail) processing option.
- Add `X-Result-Width` and `X-Result-Height` to debug headers.
- Add `IMGPROXY_KEEP_COPYRIGHT` config and [keep_copyright](https://docs.imgproxy.net/generating_the_url?id=keep-copyright) processing option.

### Change
- Use thumbnail embedded to HEIC/AVIF if its size is larger than or equal to the requested.

## [3.4.0] - 2022-04-07
### Add
- Add `IMGPROXY_FALLBACK_IMAGE_TTL` config.
- (pro) Add [watermark_size](https://docs.imgproxy.net/generating_the_url?id=watermark-size) processing option.
- Add OpenStack Object Storage ("Swift") support.
- Add `IMGPROXY_GCS_ENDPOINT` config.

### Change
- (pro) Don't check `Content-Length` header of videos.

### Fix
- (pro) Fix custom watermarks on animated images.

## [3.3.3] - 2022-03-21
### Fix
- Fix `s3` scheme status codes.
- (pro) Fix saving animations to MP4.

## [3.3.2] - 2022-03-17
### Fix
- Write logs to STDOUT instead of STDERR.
- (pro) Fix crashes when some options are used in presets.

## [3.3.1] - 2022-03-14
### Fix
- Fix transparrency in loaded ICO.
- (pro) Fix video thumbnails orientation.

## [3.3.0] - 2022-02-21
### Added
- Add the `IMGPROXY_MAX_REDIRECTS` config.
- (pro) Add the `IMGPROXY_SERVER_NAME` config.
- (pro) Add the `IMGPROXY_HEALTH_CHECK_MESSAGE` config.
- Add the `IMGPROXY_HEALTH_CHECK_PATH` config.

## [3.2.2] - 2022-02-08
### Fix
- Fix the `IMGPROXY_AVIF_SPEED` config.

## [3.2.1] - 2022-01-19
### Fix
- Fix support of BMP with unusual data offsets.

## [3.2.0] - 2022-01-18
### Added
- (pro) Add `video_meta` to the `/info` response.
- Add [zoom](https://docs.imgproxy.net/generating_the_url?id=zoom) processing option.
- Add 1/2/4-bit BMP support.

### Change
- Optimized `crop`.

### Fix
- Fix Datadog support.
- Fix `force` resizing + rotation.
- (pro) Fix `obj` gravity.

## [3.1.3] - 2021-12-17
### Fix
- Fix ETag checking when S3 is used.

## [3.1.2] - 2021-12-15
### Fix
- (pro) Fix object detection.

## [3.1.1] - 2021-12-10
### Fix
- Fix crashes in some scenarios.

## [3.1.0] - 2021-12-08
### Added
- Add `IMGPROXY_ETAG_BUSTER` config.
- (pro) [watermark_text](https://docs.imgproxy.net/generating_the_url?id=watermark-text) processing option.

### Change
- Improved ICC profiles handling.
- Proper error message when the deprecated basic URL format is used.
- Watermark offsets can be applied to replicated watermarks.

### Fix
- (pro) Fix parsing metadata of extended sequential JPEGs.

## [3.0.0] - 2021-11-23
### Added
- (pro) [Autoquality](https://docs.imgproxy.net/autoquality).
- (pro) [Object detection](https://docs.imgproxy.net/object_detection): `obj` [gravity](https://docs.imgproxy.net/generating_the_url?id=gravity) type, [blur_detections](https://docs.imgproxy.net/generating_the_url?id=blur-detections) processing option, [draw_detections](https://docs.imgproxy.net/generating_the_url?id=draw-detections) processing option.
- (pro) [Chained pipelines](https://docs.imgproxy.net/chained_pipelines)
- `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` config.
- (pro) [fallback_image_url](https://docs.imgproxy.net/generating_the_url?id=fallback-image-url) processing option.
- [expires](https://docs.imgproxy.net/generating_the_url?id=expires) processing option.
- [skip processing](https://docs.imgproxy.net/generating_the_url?id=skip-processing) processing option.
- [Datadog](./docs/datadog.md) metrics.
- `force` and `fill-down` resizing types.
- [min-width](https://docs.imgproxy.net/generating_the_url?id=min-width) and [min-height](https://docs.imgproxy.net/generating_the_url?id=min-height) processing options.
- [format_quality](https://docs.imgproxy.net/generating_the_url?id=format-quality) processing option.
- Add `X-Origin-Width` and `X-Origin-Height` to debug headers.
- Add `IMGPROXY_COOKIE_PASSTHROUGH` and `IMGPROXY_COOKIE_BASE_URL` configs.
- Add `client_ip` to requests and responses logs.

### Change
- ETag generator & checker uses source image ETag when possible.
- `304 Not Modified` responses includes `Cache-Control`, `Expires`, and `Vary` headers.
- `dpr` processing option doesn't enlarge image unless `enlarge` is true.
- imgproxy responds with `500` HTTP code when the source image downloading error seems temporary (timeout, server error, etc).
- When `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` is zero, imgproxy responds with the usual HTTP code.
- BMP support doesn't require ImageMagick.
- Save GIFs without ImageMagick (vips 8.12+ required).

### Fix
- Fix Client Hints behavior. `Width` is physical size, so we should divide it by `DPR` value.
- Fix scale-on-load in some rare cases.
- Fix the default Sentry release name.
- Fix the `health` command when the path prefix is set.
- Escape double quotes in content disposition.

### Removed
- Removed basic URL format, use [advanced one](./docs/generating_the_url.md) instead.
- Removed `IMGPROXY_MAX_SRC_DIMENSION` config, use `IMGPROXY_MAX_SRC_RESOLUTION` instead.
- Removed `IMGPROXY_GZIP_COMPRESSION` config.
- Removed `IMGPROXY_MAX_GIF_FRAMES` config, use `IMGPROXY_MAX_ANIMATION_FRAMES` instead.
- Removed `crop` resizing type, use [crop](./docs/generating_the_url.md#crop) processing option instead.
- Dropped old libvips (<8.10) support.
- (pro) Removed advanced GIF optimizations. All optimizations are applied by default ib both OSS and Pro versions.

## [3.0.0.beta2] - 2021-11-15
### Added
- Add `X-Origin-Width` and `X-Origin-Height` to debug headers.
- Add `IMGPROXY_COOKIE_PASSTHROUGH` and `IMGPROXY_COOKIE_BASE_URL` configs.

### Change
- `dpr` processing option doesn't enlarge image unless `enlarge` is true.
- `304 Not Modified` responses includes `Cache-Control`, `Expires`, and `Vary` headers.
- imgproxy responds with `500` HTTP code when the source image downloading error seems temporary (timeout, server error, etc).
- When `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` is zero, imgproxy responds with the usual HTTP code.
- BMP support doesn't require ImageMagick.

### Fix
- Fix Client Hints behavior. `Width` is physical size, so we should divide it by `DPR` value.
- Fix scale-on-load in some rare cases.
- Fix `requests_total` counter in Prometheus.

## [3.0.0.beta1] - 2021-10-01
### Added
- (pro) [Autoquality](https://docs.imgproxy.net/autoquality).
- (pro) [Object detection](https://docs.imgproxy.net/object_detection): `obj` [gravity](https://docs.imgproxy.net/generating_the_url?id=gravity) type, [blur_detections](https://docs.imgproxy.net/generating_the_url?id=blur-detections) processing option, [draw_detections](https://docs.imgproxy.net/generating_the_url?id=draw-detections) processing option.
- (pro) [Chained pipelines](https://docs.imgproxy.net/chained_pipelines)
- `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` config.
- (pro) [fallback_image_url](https://docs.imgproxy.net/generating_the_url?id=fallback-image-url) processing option.
- [expires](https://docs.imgproxy.net/generating_the_url?id=expires) processing option.
- [skip processing](https://docs.imgproxy.net/generating_the_url?id=skip-processing) processing option.
- [Datadog](./docs/datadog.md) metrics.
- `force` and `fill-down` resizing types.
- [min-width](https://docs.imgproxy.net/generating_the_url?id=min-width) and [min-height](https://docs.imgproxy.net/generating_the_url?id=min-height) processing options.
- [format_quality](https://docs.imgproxy.net/generating_the_url?id=format-quality) processing option.

### Change
- ETag generator & checker uses source image ETag when possible.

### Removed
- Removed basic URL format, use [advanced one](./docs/generating_the_url.md) instead.
- Removed `IMGPROXY_MAX_SRC_DIMENSION` config, use `IMGPROXY_MAX_SRC_RESOLUTION` instead.
- Removed `IMGPROXY_GZIP_COMPRESSION` config.
- Removed `IMGPROXY_MAX_GIF_FRAMES` config, use `IMGPROXY_MAX_ANIMATION_FRAMES` instead.
- Removed `crop` resizing type, use [crop](./docs/generating_the_url.md#crop) processing option instead.
- Dropped old libvips (<8.8) support.

## [2.17.0] - 2021-09-07
### Added
- Wildcard support in `IMGPROXY_ALLOWED_SOURCES`.

### Change
- If the source URL contains the `IMGPROXY_BASE_URL` prefix, it won't be added.

### Fix
- (pro) Fix path prefix support in the `/info` handler.

### Deprecated
- The [basic URL format](https://docs.imgproxy.net/generating_the_url_basic) is deprecated and can be removed in future versions. Use [advanced URL format](https://docs.imgproxy.net/generating_the_url_advanced) instead.

## [2.16.7] - 2021-07-20
### Change
- Reset DPI while stripping meta.

## [2.16.6] - 2021-07-08
### Fix
- Fix performance regression in ICC profile handling.
- Fix crashes while processing CMYK images without ICC profile.

## [2.16.5] - 2021-06-28
### Change
- More clear downloading errors.

### Fix
- Fix ICC profile handling in some cases.
- Fix handling of negative height value for BMP.

## [2.16.4] - 2021-06-16
### Change
- Use magenta (ff00ff) as a transparency key in `trim`.

### Fix
- Fix crashes while processing some SVGs (dockerized version).
- (pro) Fix parsing HEIF/AVIF meta.

## [2.16.3] - 2021-04-05
### Fix
- Fix PNG quantization palette size.
- Fix parsing HEIF meta.
- Fix debig header.

## [2.16.2] - 2021-03-04
### Change
- Updated dependencies in Docker.

## [2.16.1] - 2021-03-02
### Fix
- Fix delays and loop numbers of animated images.
- Fix scale-on-load of huge SVGs.
- (pro) Fix loading of PDFs with transparent background.

## [2.16.0] - 2021-02-08
### Added
- AVIF support.
- Azure Blob Storage support.
- `IMGPROXY_STRIP_COLOR_PROFILE` config and [strip_color_profile](https://docs.imgproxy.net/generating_the_url?id=strip-color-profile) processing option.
- `IMGPROXY_FORMAT_QUALITY` config.
- `IMGPROXY_AUTO_ROTATE` config and [auto_rotate](https://docs.imgproxy.net/generating_the_url?id=auto-rotate) processing option.
- [rotate](https://docs.imgproxy.net/generating_the_url?id=rotate) processing option.
- `width` and `height` arguments of the [crop](https://docs.imgproxy.net/generating_the_url?id=crop) processing option can be less than `1` that is treated by imgproxy as a relative value (a.k.a. crop by percentage).
- (pro) Remove Adobe Illustrator garbage from SVGs.
- (pro) Add IPTC tags to the `/info` response.

### Changed
- Disable scale-on-load for animated images since it causes many problems. Currently, only animated WebP is affected.
- Improved ICC profiles handling.
- (pro) Improved and optimized video thumbnails generation.

### Fix
- Fix `dpr` option.
- Fix non-strict SVG detection.
- Fix non-UTF8 SVG detection.
- Fix checking of connections in queue.
- (pro) Fix video thumbnail orientation.
- (pro) Fix EXIF fields titles in the `/info` response.

## [2.15.0] - 2020-09-03
### Added
- Ability to skip processing of some formats. See [Skip processing](https://docs.imgproxy.net/configuration?id=skip-processing).
- (pro) PDF support.
- (pro) [video_thumbnail_second](https://docs.imgproxy.net/generating_the_url?id=video-thumbnail-second) processing option.
- (pro) [page](https://docs.imgproxy.net/generating_the_url?id=page) processing option.
- (pro) [background_alpha](https://docs.imgproxy.net/generating_the_url?id=background-alpha) processing option.
- (pro) `IMGPROXY_VIDEO_THUMBNAIL_PROBE_SIZE` and `IMGPROXY_VIDEO_THUMBNAIL_MAX_ANALYZE_DURATION` configs.

### Changed
- Speed up generation of video thumbnails with large timestamps.

### Fix
- Fix `padding` and `extend` behaior when converting from a fromat without alpha support to one with alpha support.
- Fix WebP dimension limit handling.
- (pro) Fix thumbnails generation of some videos.

## [2.14.1] - 2020-07-22
### Fix
- Fix ICO saving.

## [2.14.0] - 2020-07-17
### Added
- `IMGPROXY_PROMETHEUS_NAMESPACE` config.
- [strip_metadata](https://docs.imgproxy.net/generating_the_url?id=strip-metadata) processing option.
- (pro) Configurable unsharpening. See [Unsharpening](https://docs.imgproxy.net/configuration?id=unsharpening) configs and [unsharpening](https://docs.imgproxy.net/generating_the_url?id=unsharpening) processing option.

### Changed
- Better for libvips memory metrics for Prometheus.
- Docker image includes the latest versions of dependencies.
- Optimize processing of animated images.

### Fix
- Fix error when requested WebP dimension exceeds the WebP dimension limit.
- Fix path parsing in some rare cases.
- Fix HEIC/HEIF header parsing bug.

### Deprecated
- (pro) `IMGPROXY_APPLY_UNSHARPEN_MASKING` config is deprecated, use `IMGPROXY_UNSHARPENING_MODE` instead.

## [2.13.1] - 2020-05-06
### Fixed
- Fix and optimize processing of animated images.

## [2.13.0] - 2020-04-22
### Added
- Fallback images.
- [padding](https://docs.imgproxy.net/generating_the_url?id=padding) processing option.

### Changed
- Optimized memory usage. Especially when dealing with animated images.

### Fixed
- Fix crashes during animated images processing.

## [2.12.0] - 2020-04-07
### Addded
- `IMGPROXY_PATH_PREFIX` config.
- (pro) Video thumbnails.
- (pro) [Getting the image info](https://docs.imgproxy.net/getting_the_image_info).

### Changed
- Improved `trim` processing option.
- Quantizr updated to 0.2.0 in Docker image.

## [2.11.0] - 2020-03-12
### Changed
- Replaced imagequant with [Quantizr](https://github.com/DarthSim/quantizr) in docker image.
- Removed HEIC saving support.
- Removed JBIG compressin support in TIFF.

## [2.10.1] - 2020-02-27
### Changed
- `imgproxy -v` is replaced with `imgproxy version`.

### Fixed
- Fix loadind BMP stored in ICO.
- Fix ambiguous HEIC magic bytes (MP4 videos has been detected as HEIC).
- Fix build with libvips < 8.6.
- Fix build with Go 1.14.
- Fix go module naming. Use `github.com/imgproxy/imgproxy/v2` to build imgproxy from source.

## [2.10.0] - 2020-02-13
### Added
- `IMGPROXY_NETWORK` config. Allows to bind on Unix socket.
- `IMGPROXY_CACHE_CONTROL_PASSTHROUGH` config.
- `imgproxy health` command.
- (pro) `IMGPROXY_GIF_OPTIMIZE_FRAMES` & `IMGPROXY_GIF_OPTIMIZE_TRANSPARENCY` configs and `gif_options` processing option.
- (pro) `IMGPROXY_CUSTOM_REQUEST_HEADERS`, `IMGPROXY_CUSTOM_RESPONSE_HEADERS`, and `IMGPROXY_CUSTOM_HEADERS_SEPARATOR` configs.

### Changed

- Better SVG detection.

### Fixed
- Fix detection of SVG starting with a comment.

## [2.9.0] - 2020-01-30
### Added
- `trim` processing option.
- `IMGPROXY_STRIP_METADATA` config.

### Fixed
- Fixed focus point crop calculation.

## [2.8.2] - 2020-01-13
### Changed
- Optimized memory usage.

### Fixed
- Fixed `IMGPROXY_ALLOWED_SOURCES` config.

## [2.8.1] - 2019-12-27
### Fixed
- Fix watermark top offset calculation.

## [2.8.0] - 2019-12-25
### Added
- `IMGPROXY_LOG_LEVEL` config.
- `max_bytes` processing option.
- `IMGPROXY_ALLOWED_SOURCES` config.

### Changed
- Docker image base is changed to Debian 10 for better stability and performance.
- `extend` option now supports gravity.

## [2.7.0] - 2019-11-13
### Changed
- Boolean processing options such as `enlarge` and `extend` are properly parsed. `1`, `t`, `TRUE`, `true`, `True` are truthy, `0`, `f`, `F`, `FALSE`, `false`, `False` are falsy. All other values are treated as falsy and generate a warning message.

### Fixed
- Fix segfaults on watermarking in some cases

## [2.6.2] - 2019-11-11
### Fixed
- Fix `format` option in presets.

## [2.6.1] - 2019-10-28
### Fixed
- Fix loading of some GIFs by using the edge version of giflib.

## [2.6.0] - 2019-10-23
### Added
- TIFF and BMP support.
- `IMGPROXY_REPORT_DOWNLOADING_ERRORS` config. Setting it to `false` disables reporting of downloading errors.
- SVG passthrough. When source image and requested format are SVG, image will be returned without changes.
- `IMGPROXY_USE_GCS` config. When it set to true and `IMGPROXY_GCS_KEY` is not set, imgproxy tries to use Application Default Credentials to get access to GCS bucket.

### Changed
- Reimplemented and more errors-tolerant image size parsing.
- Log only modified processing options.

### Fixed
- Fixed sharpening+watermarking.
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
- [filename](./docs/generating_the_url.md#filename) option.

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
- [crop](./docs/generating_the_url.md#crop) processing option. `resizing_type:crop` is deprecated;
- Offsets for [gravity](./docs/generating_the_url.md#gravity);
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
- [extend](./docs/generating_the_url.md#extend) processing option.
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
- [Plain source URLs](./docs/generating_the_url.md#plain) support.
- [Serving images from Google Cloud Storage](./docs/serving_files_from_google_cloud_storage.md).
- [Full support of GIFs](./docs/image_formats_support.md#gif-support) including animated ones.
- [Watermarks](./docs/watermark.md).
- [New Relic](./docs/new_relic.md) metrics.
- [Prometheus](./docs/prometheus.md) metrics.
- [DPR](./docs/generating_the_url.md#dpr) option (thanks to [selul](https://github.com/selul)).
- [Cache buster](./docs/generating_the_url.md#cache-buster) option.
- [Quality](./docs/generating_the_url.md#quality) option.
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
- [New advanced URL format](./docs/generating_the_url.md). Unleash the full power of imgproxy v2.0.
- [Presets](./docs/presets.md). Shorten your urls by reusing processing options.
- [Serving images from Amazon S3](./docs/serving_files_from_s3.md). Thanks to [@crohr](https://github.com/crohr), now we have a way to serve files from private S3 buckets.
- [Autoconverting to WebP when supported by browser](./docs/configuration.md#avifwebp-support-detection) (disabled by default). Use WebP as resulting format when browser supports it.
- [Gaussian blur](./docs/generating_the_url.md#blur) and [sharpen](./docs/generating_the_url.md#sharpen) filters. Make your images look better than before.
- [Focus point gravity](./docs/generating_the_url.md#gravity). Tell imgproxy what point will be the center of the image.
- [Background color](./docs/generating_the_url.md#background). Control the color of background when converting PNG with alpha-channel to JPEG.

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
