# Changelog

## [Unreleased]
### Added
- Add [IMGPROXY_FAIL_ON_DEPRECATION](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_FAIL_ON_DEPRECATION) config. When set to `true`, imgproxy will exit with a fatal error if a deprecated config option is used.
- Add [flip](https://docs.imgproxy.net/latest/usage/processing#flip) processing option.
- (pro) Add [IMGPROXY_AVIF_SUBSAMPLE](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_AVIF_SUBSAMPLE) config.
- (pro) Add [avif_options](https://docs.imgproxy.net/latest/usage/processing#avif-options) processing option.
- (pro) Return `orientation` field in the `/info` endpoint response when the [dimensions](https://docs.imgproxy.net/latest/usage/getting_info#dimensions) info option is enabled.

### Changed
- When image source responds with a 4xx status code, imgproxy now responds with the same status code instead of always responding with `404 Not Found`.
- When image source responds with a 5xx status code, imgproxy now responds with `502 Bad Gateway` instead of `500 Internal Server Error`.
- Remove `iframe` elements from SVGs during sanitization.

### Fixed
- Fix crop coordinates calculation when the image has an EXIF orientation different from `1` and the `rotate` processing option is used.
- Fix responding with 404 when a GCS bucket or object is missing.
- Fix handling `:` encoded as `%3A` in processing/info options.

## [3.30.1] - 2025-10-10
### Changed
- Format New Relic and OpenTelemetry metadata values that implement the `fmt.Stringer` interface as strings.

### Fixed
- (pro) Fix memory leak during video thumbnail generation.

## [3.30.0] - 2025-09-17
### Added
- Add [IMGPROXY_GRACEFUL_STOP_TIMEOUT](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_GRACEFUL_STOP_TIMEOUT) config.
- (pro) Add [color_profile](https://docs.imgproxy.net/latest/usage/processing#color-profile) processing option and [IMGPROXY_COLOR_PROFILES_DIR](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_COLOR_PROFILES_DIR) config.

### Changed
- Update the default graceful stop timeout to twice the [IMGPROXY_TIMEOUT](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_TIMEOUT) config value.
- (pro) Improve video decoding performance.
- (pro) Respond with `422 Unprocessable Entity` on error during video decoding.

### Fixed
- Fix the `Vary` header value when `IMGPROXY_AUTO_JXL` or `IMGPROXY_ENFORCE_JXL` configs are set to `true`.
- Fix connection break when the `raw` processing option is used and the response status code does not allow a response body (such as `304 Not Modified`).
- Fix the `If-Modified-Since` request header handling when the `raw` processing option is used.
- Fix `X-Origin-Height` and `X-Result-Height` debug header values for animated images.
- Fix keeping copyright info in EXIF.
- Fix preserving color profiles in TIFF images.
- Fix freezes during sanitization or minification of some broken SVGs.
- (pro) Fix generating thumbnails for VP9 videos with high bit depth.
- (pro) Fix `IMGPROXY_CUSTOM_RESPONSE_HEADERS` and `IMGPROXY_RESPONSE_HEADERS_PASSTHROUGH` configs behavior when the `raw` processing option is used.

## [3.29.1] - 2025-07-11
### Fixed
- Fix parsing and minifying some SVGs.

## [3.29.0] - 2025-07-08
### Added
- Add [IMGPROXY_MAX_RESULT_DIMENSION](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_MAX_RESULT_DIMENSION) config and [max_result_dimension](https://docs.imgproxy.net/latest/usage/processing#max-result-dimension) processing option.
- Add [IMGPROXY_ALLOWED_PROCESSING_OPTIONS](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_ALLOWED_PROCESSING_OPTIONS) config.
- (pro) Add [IMGPROXY_ALLOWED_INFO_OPTIONS](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_ALLOWED_INFO_OPTIONS) config.
- (pro) Add [IMGPROXY_MAX_CHAINED_PIPELINES](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_MAX_CHAINED_PIPELINES) config.
- Add `imgproxy.source_image_origin` attribute to New Relic, DataDog, and OpenTelemetry traces.
- Add `imgproxy.source_image_url` and `imgproxy.source_image_origin` attributes to `downloading_image` spans in New Relic, DataDog, and OpenTelemetry traces.
- Add `imgproxy.processing_options` attribute to `processing_image` spans in New Relic, DataDog, and OpenTelemetry traces.
- Add `Source Image Origin` attribute to error reports.
- Add `workers` and `workers_utilization` metrics to all metrics services.
- (pro) Add [crop_aspect_ratio](https://docs.imgproxy.net/latest/usage/processing#crop-aspect-ratio) processing option.
- (pro) Add [IMGPROXY_PDF_NO_BACKGROUND](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_PDF_NO_BACKGROUND) config.
- Add [IMGPROXY_WEBP_EFFORT](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_WEBP_EFFORT) config.
- Add [IMGPROXY_WEBP_PRESET](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_WEBP_PRESET) config.
- (pro) Add support for saving images as PDF.

### Changed
- Suppress "Response has no supported checksum" warnings from S3 SDK.
- The [IMGPROXY_USER_AGENT](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_USER_AGENT) config now supports the `%current_version` variable that is replaced with the current imgproxy version.
- (docker) Optimized image quantization.
- (pro) Improved BlurHash generation performance.

### Fixed
- Fix `X-Origin-Content-Length` header value when SVG is sanitized or minified.
- Mark JPEG XL format as supporting quality. Fixes autoquality for JPEG XL.
- Fix the `extend` processing option when only one dimension is set.
- (pro) Fix object detection when the `IMGPROXY_USE_LINEAR_COLORSPACE` config is set to `true`.
- (pro) Fix BlurHash generation when the `IMGPROXY_USE_LINEAR_COLORSPACE` config is set to `true`.
- (pro) Fix detection of PDF files with a header offset.

### Removed
- Remove the `IMGPROXY_SVG_FIX_UNSUPPORTED` config. The problem it was solving is now fixed in librsvg.

## [3.28.0] - 2025-03-31
### Added
- Add [IMGPROXY_BASE64_URL_INCLUDES_FILENAME](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_BASE64_URL_INCLUDES_FILENAME) config.
- Add [IMGPROXY_COOKIE_PASSTHROUGH_ALL](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_COOKIE_PASSTHROUGH_ALL) config.
- (pro) Add PNG EXIF and XMP data to the `/info` endpoint response.
- (pro) Add the `mime_type` field to the `/info` endpoint response.

### Changed
- Treat 206 (Partial Content) responses from a source server as 200 (OK) when they contain a full content range.
- Improved error reporting.
- (pro) Change saturation adjustment algorithm to be more CIE-correct.
- (pro) Don't check image complexity during best format selection when `IMGPROXY_BEST_FORMAT_COMPLEXITY_THRESHOLD` is set to `0`.

### Fixed
- Fix determinimg the default hostname for cookies passthrough.
- (pro) Fix setting the `Host` header with the `IMGPROXY_CUSTOM_REQUEST_HEADERS` config.
- (pro) Fix passing through the `Host` header with the `IMGPROXY_REQUEST_HEADERS_PASSTHROUGH` config.
- (pro) Fix passing through request headers with the `IMGPROXY_REQUEST_HEADERS_PASSTHROUGH` when the `raw` option is used.
- (pro) Fix `IMGPROXY_BEST_FORMAT_ALLOW_SKIPS` config behavior.
- (pro) Fix flattening behavior when chained pipelines are used and the resulting format doesn't support transparency.
- (pro) Fix advanced smart crop when the most feature points are located close to the right or the bottom edge.

### Removed
- Remove the `IMGPROXY_S3_MULTI_REGION` config. imgproxy now always work in multi-regional S3 mode.

## [3.27.2] - 2025-01-27
### Fixed
- Fix preventing requests to `0.0.0.0` when imgproxy is configured to deny loopback addresses.
- (pro) Fix timeouts in AWS Lambda when running in development mode.

## [3.27.1] - 2025-01-13
### Added
- Add [IMGPROXY_SOURCE_URL_QUERY_SEPARATOR](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_SOURCE_URL_QUERY_SEPARATOR) config.
- (pro) Add YOLOv11 object detection model support.

### Changed
- (pro) Improve image complexity calculation for best format selection.
- (pro) Use PNG quantization for very low-complexity images when the `best` format is used.

### Fixed
- Fix blur and sharpen performance for images with alpha channel.
- (pro) Fix object detecttion accuracy.

## [3.27.0] - 2024-12-18
### Add
- Add JPEG XL (JXL) support.
- Add PSD (Photoshop Document) and PSB (Photoshop Big) images support.
- Add [IMGPROXY_AUTO_JXL](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_AUTO_JXL), [IMGPROXY_ENFORCE_JXL](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_ENFORCE_JXL), and [IMGPROXY_JXL_EFFORT](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_JXL_EFFORT) configs.
- (pro) Add [IMGPROXY_AUTOQUALITY_JXL_NET](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_AUTOQUALITY_JXL_NET) config.
- (pro) Add [objects_position](https://docs.imgproxy.net/latest/usage/processing#objects-position) processing and info options.
- (pro) Add [IMGPROXY_OBJECT_DETECTION_SWAP_RB](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_OBJECT_DETECTION_SWAP_RB) config.
- (pro) Add [IMGPROXY_OBJECT_DETECTION_GRAVITY_MODE](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_OBJECT_DETECTION_GRAVITY_MODE) config.

### Changed
- Change `IMGPROXY_AVIF_SPEED` default value to `8`.
- Change `IMGPROXY_FORMAT_QUALITY` default value to `webp=79,avif=63,jxl=77`.
- Rename `IMGPROXY_ENABLE_WEBP_DETECTION` to `IMGPROXY_AUTO_WEBP`. The old name is deprecated but still supported.
- Rename `IMGPROXY_ENABLE_AVIF_DETECTION` to `IMGPROXY_AUTO_AVIF`. The old name is deprecated but still supported.
- (pro) Change `IMGPROXY_AUTOQUALITY_FORMAT_MIN` default value to `avif=60`.
- (pro) Change `IMGPROXY_AUTOQUALITY_FORMAT_MAX` default value to `avif=65`.
- (pro) Use the last page/frame of the source image when the `page` processing option value is greater than or equal to the number of pages/frames in the source image.

### Fixed
- Fix detecting of width and height of HEIF images that include `irot` boxes.
- Set `Error` status for errorred traces in OpenTelemetry.
- Fix URL parsing error when a non-http(s) URL contains a `%` symbol outside of the percent-encoded sequence.
- Fix importing ICC profiles for 16-bit images with an alpha channel.
- Fix handling ICC profiles with vips 8.15+.
- (pro) Fix opject detection accuracy when using YOLOv8 or YOLOv10 models.
- (pro) Fix usage of the `obj` and `objw` gravity types inside the `crop` processing option.
- (pro) Fix detecting of width and height when orientation is specified in EXIF but EXIF info is not requested.
- (pro) Fix watermark shadow clipping.

### Deprecated
- `IMGPROXY_ENABLE_WEBP_DETECTION` config is deprecated. Use `IMGPROXY_AUTO_WEBP` instead.
- `IMGPROXY_ENABLE_AVIF_DETECTION` config is deprecated. Use `IMGPROXY_AUTO_AVIF` instead.

## [3.26.1] - 2024-10-28
### Changed
- (pro) Improve `monochrome` and `duotone` processing options.

### Fixed
- Fix loading log configs from local files and secret managers.
- Fix detecting HEIF images with the `heix` brand.
- Fix downloading source images when the image source requires a cookie challenge.
- (pro) Fix playback of videos created from animations in Google Chrome.
- (pro) Fix detecting of width and height of HEIF images that have orientation specified in EXIF.

## [3.26.0] - 2024-09-16
### Added
- Add `imgproxy.source_image_url` and `imgproxy.processing_options` attributes to New Relic, DataDog, and OpenTelemetry traces.
- Add [IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE) config.
- (pro) Add [monochrome](https://docs.imgproxy.net/latest/usage/processing#monochrome) processing option.
- (pro) Add [duotone](https://docs.imgproxy.net/latest/usage/processing#duotone) processing option.
- (pro) Add `objw` [gravity](https://docs.imgproxy.net/latest/usage/processing#gravity) type.
- (pro) Add an object pseudo-class `all` that can be used with the `obj` and `objw` [gravity](https://docs.imgproxy.net/latest/usage/processing#gravity) types to match all detected objects.
- (pro) Add [IMGPROXY_SMART_CROP_ADVANCED_MODE](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_SMART_CROP_ADVANCED_MODE) config.
- (pro) Add [IMGPROXY_OBJECT_DETECTION_FALLBACK_TO_SMART_CROP](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_OBJECT_DETECTION_FALLBACK_TO_SMART_CROP) config.
- (docker) Add a script for [building Linux packages](https://docs.imgproxy.net/latest/installation#building-linux-packages).

### Changed
- Properly set the `net.host.name` and `http.url` tags in OpenTelemetry traces.
- (pro) Object detection: [class names file](https://docs.imgproxy.net/latest/object_detection#class-names-file) can contain object classes' weights.

### Fixed
- Fix handling `#` symbols in `local://`, `s3://`, `gcs://`, `abs://`, and `swift://` URLs.
- Fix `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` value check. Allow `0` value.
- (docker) Fix loading HEIC images made with iOS 18.

## [3.25.0] - 2024-07-08
### Added
- Add [IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID) config.
- Add [IMGPROXY_WRITE_RESPONSE_TIMEOUT](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_WRITE_RESPONSE_TIMEOUT) config.
- Add [IMGPROXY_REPORT_IO_ERRORS](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_REPORT_IO_ERRORS) config.
- Add [IMGPROXY_ARGUMENTS_SEPARATOR](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_ARGUMENTS_SEPARATOR) config.
- Add [IMGPROXY_PRESETS_SEPARATOR](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_PRESETS_SEPARATOR) config.
- (pro) Add support for object detection models in ONNX format.
- (pro) Add [colorize](https://docs.imgproxy.net/latest/usage/processing#colorize) processing option.
- (pro) Add [IMGPROXY_WATERMARK_PREPROCESS_URL](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_WATERMARK_PREPROCESS_URL) config.
- (pro) Add [IMGPROXY_FALLBACK_IMAGE_PREPROCESS_URL](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_FALLBACK_IMAGE_PREPROCESS_URL) config.

### Changed
- Automatically add `http://` scheme to the `IMGPROXY_S3_ENDPOINT` value if it has no scheme.
- Trim redundant slashes in the S3, GCS, ABS, and Swift object keys.
- Rename `IMGPROXY_WRITE_TIMEOUT` to `IMGPROXY_TIMEOUT`. The old name is deprecated but still supported.
- Rename `IMGPROXY_READ_TIMEOUT` to `IMGPROXY_READ_REQUEST_TIMEOUT`. The old name is deprecated but still supported.
- (pro) Allow specifying [gradient](https://docs.imgproxy.net/latest/usage/processing#gradient) direction as an angle in degrees.
- (pro) Speed up ML features.
- (pro) Update the face detection model.

### Fixed
- Fix HEIC/AVIF dimension limit handling.
- Fix SVG detection when the root element has a namespace.
- Fix treating percent-encoded symbols in `local://`, `s3://`, `gcs://`, `abs://`, and `swift://` URLs.
- (pro) Fix style injection to SVG.
- (pro) Fix video tiles generation when the video's SAR is not `1`.

### Deprecated
- `IMGPROXY_WRITE_TIMEOUT` config is deprecated. Use `IMGPROXY_TIMEOUT` instead.
- `IMGPROXY_READ_TIMEOUT` config is deprecated. Use `IMGPROXY_READ_REQUEST_TIMEOUT` instead.

## [3.24.1] - 2024-04-30
### Fixed
- Fix the default `IMGPROXY_WORKERS` value when cgroup limits are applied.

## [3.24.0] - 2024-04-29
### Added
- Add [IMGPROXY_ALWAYS_RASTERIZE_SVG](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_ALWAYS_RASTERIZE_SVG) config.
- Add [IMGPROXY_PNG_UNLIMITED](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_PNG_UNLIMITED) and [IMGPROXY_SVG_UNLIMITED](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_SVG_UNLIMITED) configs.
- (pro) Add `fill`, `foxus_x`, and `focus_y` arguments to the [video_thumbnail_tile](https://docs.imgproxy.net/latest/usage/processing#video-thumbnail-tile) processing option.
- (pro) Add the `ch` (chessboard order) position for watermarks.
- (pro) Add the [watermark_rotate](https://docs.imgproxy.net/latest/usage/processing#watermark-rotate) processing option.

### Changed
- Respond with 404 when the bucket/container name or object key is empty in an S3, Google Cloud Storage, Azure Blob Storage, or OpenStack Object Storage (Swift) URL.
- Ensure that the watermark is always centered when replicated.
- (pro) Improve unsharp masking.
- (docker) Update AWS Lambda adapter to 0.8.3.
- (docker) Increase EXIF size limit to 8MB.

### Fixed
- Fix parsing some TIFFs.
- Fix over-shrinking during scale-on-load.
- Fix watermarks overlapping animation frames in some cases.
- (pro) Fix false-positive video detections.

## [3.23.0] - 2024-03-11
### Added
- Add request ID, processing/info options, and source image URL to error reports.

### Changed
- Support configuring OpenTelemetry with standard [general](https://opentelemetry.io/docs/languages/sdk-configuration/general/) and [OTLP Exporter](https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/) environment variables.
- `IMGPROXY_MAX_SRC_RESOLUTION` default value is increased to 50.

### Fixed
- Fix loading environment variables from the AWS System Manager Parameter Store when there are more than 10 parameters.
- (pro) Fixed thumbnail generation for MKV/WebM files containing blocks invalidly marked as keyframes.

### Deprecated
- `IMGPROXY_OPEN_TELEMETRY_ENDPOINT`, `IMGPROXY_OPEN_TELEMETRY_PROTOCOL`, `IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE`, `IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME`, `IMGPROXY_OPEN_TELEMETRY_PROPAGATORS`, and `IMGPROXY_OPEN_TELEMETRY_CONNECTION_TIMEOUT` config options are deprecated. Use standard OpenTelemetry environment variables instead.

## [3.22.0] - 2024-02-22
### Added
- Add the [IMGPROXY_TRUSTED_SIGNATURES](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_TRUSTED_SIGNATURES) config.
- (pro) Add the [hashsum](https://docs.imgproxy.net/latest/usage/processing#hashsum) processing and info options.
- (pro) Add the [calc_hashsums](https://docs.imgproxy.net/latest/usage/getting_info#calc-hashsums) info option.
- (pro) Add the [IMGPROXY_VIDEO_THUMBNAIL_TILE_AUTO_KEYFRAMES](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_VIDEO_THUMBNAIL_TILE_AUTO_KEYFRAMES) config.
- (pro) Add the [IMGPROXY_WEBP_SMART_SUBSAMPLE](https://docs.imgproxy.net/latest/configuration/options#IMGPROXY_WEBP_SMART_SUBSAMPLE) config and the `smart_subsample` argument to the [webp_options](https://docs.imgproxy.net/latest/usage/processing#webp-options) processing option
- (docker) Add lambda adapter to the Docker image.

### Changed
- Allow relative values for `gravity` and `watermark` offsets.
- Revised downloading errors reporting.
- Allow `IMGPROXY_TTL` to be zero.
- Don't set `Expires` HTTP header as it is ignored if the `Cache-Control` header is set.
- Don't log health-check requests and responses.
- Enforce `IMGPROXY_WORKERS=1` when running in AWS Lambda.
- Reduce memory usage when scaling down animated images.
- (pro) If the `step` argument of the `video_thumbnail_tile` is negative, calculate `step` automatically.

### Fixed
- Fix loading animated images with a huge number of frames.
- Fix recursive presets detection.
- (pro) Fix `video_thumbnail_tile` option behavior when the video has a single keyframe.
- (pro) Fix the `trim` argument of the `video_thumbnail_tile` processing option.
- (pro) Fix `video_thumbnail_tile` behavior when the `step` argument value is less than frame duration.
- (pro) Fix VPx video stream duration detection.
- (pro) Fix thumbnal generation for VP9 videos.
- (pro) Fix thumbnal generation for videos with a large time base denumenator.

## [3.21.0] - 2023-11-23
### Added
- Add `status_codes_total` counter to Prometheus metrics.
- Add client-side decryption support for S3 integration.
- Add HEIC saving support.
- (pro) Add the `IMGPROXY_VIDEO_THUMBNAIL_KEYFRAMES` config and the [video_thumbnail_keyframes](https://docs.imgproxy.net/latest/usage/processing#video-thumbnail-keyframes) processing option.
- (pro) Add the [video_thumbnail_tile](https://docs.imgproxy.net/latest/usage/processing#video-thumbnail-tile) processing option.
- (pro) Add the `duration` field to the video streams information in the `/info` endpoint response.
- (pro) Add the [colorspace](https://docs.imgproxy.net/latest/usage/getting_info#colorspace), [bands](https://docs.imgproxy.net/latest/usage/getting_info#bands), [sample_format](https://docs.imgproxy.net/latest/usage/getting_info#sample-format), [pages_number](https://docs.imgproxy.net/latest/usage/getting_info#pages-number), and [alpha](https://docs.imgproxy.net/latest/usage/getting_info#alpha) info options.

### Changed
- (pro) Improve video detection.

### Fixed
- (pro) Fix detection of some videos.
- (pro) Fix headers and cookies passthrough when the source is a video.
- (pro) Fix wrong behavior of the `background_alpha` option when the `best` format is used.
- (docker) Fix saving EXIF strings containing invalid UTF-8 characters.
- (docker) Fix possible segfaults while processing HEIC/AVIF images.
- (docker) Fix rendering GIFs embedded in SVGs.

## [3.20.0] - 2023-10-09
### Added
- (pro) Add [info options](https://docs.imgproxy.net/latest/getting_the_image_info?id=info-options) support to the `/info` endpoint.
- (pro) Add video streams info to the `/info` endpoint response.
- (docker) Add support for TIFFs with 16-bit float samples.
- (docker) Add support for TIFFs with the old-style JPEG compression.

### Changed
- Limit vector image sizes to `IMGPROXY_MAX_SRC_RESOLUTION`.
- (pro) Respect image orientation when extracting image dimensions for the `/info` endpoint response.
- (pro) Respect `IMGPROXY_WORKERS` and `IMGPROXY_REQUESTS_QUEUE_SIZE` configs in the `/info` endpoint.
- (pro) Collect detailed metrics for the `/info` endpoint.
- (docker) Invalid UTF-8 strings in image metadata are fixed instead of being ignored.

### Fixed
- Fix parsing of HEIF files with large boxes.
- Fix wrong colors when the source image has a linear colorspace.
- Fix wrong colors or opacity when the source image is a TIFF with a float sample format.
- Fix crashes during processing of large animated WebPs.
- Fix `vips_allocs` OTel metric unit (was `By`, fixed to `1`).
- (pro) Fix generating thumbnails for WebM videos with transparency.
- (pro) Fix style injection into some SVGs.

## [3.19.0] - 2023-08-21
### Added
- Add `IMGPROXY_WORKERS` alias for the `IMGPROXY_CONCURRENCY` config.
- Add [multi-region mode](https://docs.imgproxy.net/latest/serving_files_from_s3?id=multi-region-mode) to S3 integration.
- Add the ability to [load environment variables](https://docs.imgproxy.net/latest/loading_environment_variables) from a file or a cloud secret.
- (pro) Add [pages](https://docs.imgproxy.net/latest/generating_the_url?id=pages) processing option.

### Changed
- Don't report `The image request is cancelled` errors.
- Create and destroy a tiny image during health check to check that vips is operational.
- (pro) Change the `/info` endpoint behavior to return only the first EXIF/XMP/IPTC block data of JPEG if the image contains multiple metadata blocks of the same type.

### Fixed
- Fix reporting image loading errors.
- Fix the `Cache-Control` and `Expires` headers behavior when both `IMGPROXY_CACHE_CONTROL_PASSTHROUGH` and `IMGPROXY_FALLBACK_IMAGE_TTL` configs are set.
- (pro) Fix the `IMGPROXY_FALLBACK_IMAGE_TTL` config behavior when the `fallback_image_url` processing option is used.

## [3.18.2] - 2023-07-13
### Fixed
- Fix saving to JPEG when using linear colorspace.
- Fix the `Cache-Control` and `Expires` headers passthrough when SVG is sanitized or fixed.
- (pro) Fix complexity calculation for still images.
- (docker) Fix crashes during some resizing cases.

## [3.18.1] - 2023-06-29
### Changed
- Change maximum and default values of `IMGPROXY_AVIF_SPEED` to `9`.
- (pro) Fix detection of some videos.
- (pro) Better calculation of the image complexity during choosing the best format.
- (docker) Fix freezes and crashes introduced in v3.18.0 by liborc.

## [3.18.0] - 2023-05-31
### Added
- Add `IMGPROXY_URL_REPLACEMENTS` config.
- (pro) Add `IMGPROXY_STRIP_METADATA_DPI` config.
- (pro) Add [dpi](https://docs.imgproxy.net/latest/generating_the_url?id=dpi) processing option.
- (pro) Add WebP EXIF and XMP to the `/info` response.
- (pro) Add Photoshop resolution data to the `/info` response.

### Changed
- Preserve GIF's bit-per-sample.
- Respond with 422 on error during image loading.

### Fixed
- (pro) Fix applying the `resizing_algorithm` processing option when resizing images with an alpha channel.

## [3.17.0] - 2023-05-10
### Added
- Add `process_resident_memory_bytes`, `process_virtual_memory_bytes`, `go_memstats_sys_bytes`, `go_memstats_heap_idle_bytes`, `go_memstats_heap_inuse_bytes`, `go_goroutines`, `go_threads`, `buffer_default_size_bytes`, `buffer_max_size_bytes`, and `buffer_size_bytes` metrics to OpenTelemetry.
- Add support for the `Last-Modified` response header and the `If-Modified-Since` request header (controlled by the `IMGPROXY_USE_LAST_MODIFIED` config).
- Add `IMGPROXY_S3_ASSUME_ROLE_ARN` config.
- Add `IMGPROXY_MALLOC` Docker-only config.

### Changed
- Optimized memory buffers pooling for better performance and memory reusage.
- Optimized watermarks application.

### Fixed
- Fix crushes when `watermark_text` has an invalid value.

## [3.16.1] - 2023-04-26
### Fixed
- Fix crashes in cases where the `max_bytes` processing option was used and image saving failed.
- Fix error when using the `extend` or `extend_aspect_ratio` processing option while setting zero width or height.
- Fix color loss when applying a watermark with a palette on an image without a palette.
- (pro) Fix crashes when using `IMGPROXY_SMART_CROP_FACE_DETECTION` with large `IMGPROXY_CONCURRENCY`.
- (pro) Fix watermark scaling when neither watermark scale nor watermark size is specified.

## [3.16.0] - 2023-04-18
### Added
- Add support for `Sec-CH-DPR` and `Sec-CH-Width` client hints.
- Add support for Base64-encoded `filename` processing option values.
- Add `IMGPROXY_REQUEST_HEADERS_PASSTHROUGH` and `IMGPROXY_RESPONSE_HEADERS_PASSTHROUGH` configs.

### Changed
- Improved object-oriented crop.

### Fixed
- Fix detection of dead HTTP/2 connections.
- Fix the way the `dpr` processing option affects offsets and paddings.

### Remove
- Remove suport for `Viewport-Width` client hint.
- Don't set `Content-DPR` header (deprecated in the specification).

## [3.15.0] - 2023-04-10
### Added
- Add the `IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES`, `IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES`, and `IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES` configs.

### Changed
- Connecting to loopback, link-local multicast, and link-local unicast IP addresses when requesting source images is prohibited by default.
- Tuned source image downloading flow.
- Disable extension checking if the `raw` processing option is used.

### Fixed
- (pro) Fix face detection during advanced smart crop in some cases.

## [3.14.0] - 2023-03-07
## Added
- Add [extend_aspect_ratio](https://docs.imgproxy.net/latest/generating_the_url?id=extend-aspect-ratio) processing option.
- Add the `IMGPROXY_ALLOW_SECURITY_OPTIONS` config + `max_src_resolution`, `max_src_file_size`, `max_animation_frames`, and `max_animation_frame_resolution` processing options.
- (pro) Add [advanced smart crop](https://docs.imgproxy.net/latest/configuration?id=smart-crop).

### Changed
- Make the `expires` processing option set `Expires` and `Cache-Control` headers.
- Sanitize `use` tags in SVGs.

### Fixed
- Fix handling some ICC profiles.

## [3.13.2] - 2023-02-15
### Changed
- Remove color-related EXIF data when stripping ICC profile.
- (pro) Optimize saving to MP4.

### Fixed
- (pro) Fix saving with autoquality in some cases.
- (pro) Fix saving large images to MP4.

## [3.13.1] - 2023-01-16
### Fixed
- Fix applying watermarks with replication.

## [3.13.0] - 2023-01-11
### Changed
- Add support for Managed Identity or Service Principal credentials to Azure Blob Storage integration.
- Optimize memory usage in some scenarios.
- Better SVG sanitization.
- (pro) Allow usage of floating-point numbers in the `IMGPROXY_VIDEO_THUMBNAIL_SECOND` config and the `video_thumbnail_second` processing option.

### Fixed
- Fix crashes in some cases when using OpenTelemetry in Amazon ECS.
- (pro) Fix saving of GIF with too small frame delay to MP4

## [3.12.0] - 2022-12-11
### Added
- Add `IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION` config.
- Add [Amazon CloudWatch](https://docs.imgproxy.net/latest/cloud_watch) support.
- (pro) Add [`best` resultig image format](https://docs.imgproxy.net/latest/best_format).
- (pro) Add `IMGPROXY_WEBP_COMPRESSION` config and [webp_options](https://docs.imgproxy.net/latest/generating_the_url?id=webp-options) processing option.

### Changed
- Change `IMGPROXY_FORMAT_QUALITY` default value to `avif=65`.
- Change `IMGPROXY_AVIF_SPEED` default value to `8`.
- Change `IMGPROXY_PREFERRED_FORMATS` default value to `jpeg,png,gif`.
- Set `Cache-Control: no-cache` header to the health check responses.
- Allow replacing line breaks with `\n` in `IMGPROXY_OPEN_TELEMETRY_SERVER_CERT`, `IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT`, and`IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY`.

### Fixed
- Fix 3GP video format detection.

## [3.11.0] - 2022-11-17
### Added
- Add `IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE` config.
- Add `IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR` config.
- (pro) Add XMP data to the `/info` response.

### Changed
- Better XMP data stripping.
- Use parent-based OpenTelemetry sampler by default.
- Use fixed TraceIdRatioBased sampler with OpenTelemetry.

## [3.10.0] - 2022-11-04
### Added
- Add `IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT` config.
- (pro) Add [disable_animation](https://docs.imgproxy.net/latest/generating_the_url?id=disable-animation) processing option.
- (pro) Add [gradient](https://docs.imgproxy.net/latest/generating_the_url?id=gradient) processing option.

### Fixed
- Fix false-positive SVG detections.
- Fix possible infinite loop during SVG sanitization.
- (pro) Fix saving of GIF with variable frame delay to MP4.
- (pro) Fix the size of video thumbnails if the video has a defined sample aspect ratio.

## [3.9.0] - 2022-10-19
### Added
- Add `IMGPROXY_SVG_FIX_UNSUPPORTED` config.

### Fixed
- Fix HTTP response status when OpenTelemetry support is enabled.
- (docker) Fix saving of paletted PNGs with low bit-depth.

## [3.8.0] - 2022-10-06
### Added
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
### Fixed
- Fix memory bloat in some cases.
- Fix `format_quality` usage in presets.

## [3.7.0] - 2022-07-27
### Added
- Add support of 16-bit BMP.
- Add `IMGPROXY_NEW_RELIC_LABELS` config.
- Add support of JPEG files with differential Huffman coding or arithmetic coding.
- Add `IMGPROXY_PREFERRED_FORMATS` config.
- Add `IMGPROXY_REQUESTS_QUEUE_SIZE` config.
- Add `requests_in_progress` and `images_in_progress` metrics.
- Add queue segment/span to request traces.
- Add sending additional metrics to Datadog and `IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS` config.
- Add sending additional metrics to New Relic.

### Changed
- Change `IMGPROXY_MAX_CLIENTS` default value to 2048.
- Allow unlimited connections when `IMGPROXY_MAX_CLIENTS` is set to `0`.
- Change `IMGPROXY_TTL` default value to `31536000` (1 year).
- Better errors tracking with metrics services.
- (docker) Faster and better saving of GIF.
- (docker) Faster saving of AVIF.
- (docker) Faster loading and saving of PNG.

### Fixed
- Fix trimming of CMYK images.
- Respond with 404 when the source image can not be found in OpenStack Object Storage.
- Respond with 404 when file wasn't found in the GCS storage.

## [3.6.0] - 2022-06-13
### Added
- Add `IMGPROXY_RETURN_ATTACHMENT` config and [return_attachment](https://docs.imgproxy.net/generating_the_url?return-attachment) processing option.
- Add SVG sanitization and `IMGPROXY_SANITIZE_SVG` config.

### Changed
- Better animation detection.

### Fixed
- Respond with 404 when file wasn't found in the local storage.

## [3.5.1] - 2022-05-20
### Changed
- Fallback from AVIF to JPEG/PNG if one of the result dimensions is smaller than 16px.

### Fixed
- (pro) Fix some PDF pages background.
- (docker) Fix loading some HEIF images.

## [3.5.0] - 2022-04-25
### Added
- Add support of RLE-encoded BMP.
- Add `IMGPROXY_ENFORCE_THUMBNAIL` config and [enforce_thumbnail](https://docs.imgproxy.net/generating_the_url?id=enforce-thumbnail) processing option.
- Add `X-Result-Width` and `X-Result-Height` to debug headers.
- Add `IMGPROXY_KEEP_COPYRIGHT` config and [keep_copyright](https://docs.imgproxy.net/generating_the_url?id=keep-copyright) processing option.

### Changed
- Use thumbnail embedded to HEIC/AVIF if its size is larger than or equal to the requested.

## [3.4.0] - 2022-04-07
### Added
- Add `IMGPROXY_FALLBACK_IMAGE_TTL` config.
- (pro) Add [watermark_size](https://docs.imgproxy.net/generating_the_url?id=watermark-size) processing option.
- Add OpenStack Object Storage ("Swift") support.
- Add `IMGPROXY_GCS_ENDPOINT` config.

### Changed
- (pro) Don't check `Content-Length` header of videos.

### Fixed
- (pro) Fix custom watermarks on animated images.

## [3.3.3] - 2022-03-21
### Fixed
- Fix `s3` scheme status codes.
- (pro) Fix saving animations to MP4.

## [3.3.2] - 2022-03-17
### Fixed
- Write logs to STDOUT instead of STDERR.
- (pro) Fix crashes when some options are used in presets.

## [3.3.1] - 2022-03-14
### Fixed
- Fix transparrency in loaded ICO.
- (pro) Fix video thumbnails orientation.

## [3.3.0] - 2022-02-21
### Added
- Add the `IMGPROXY_MAX_REDIRECTS` config.
- (pro) Add the `IMGPROXY_SERVER_NAME` config.
- (pro) Add the `IMGPROXY_HEALTH_CHECK_MESSAGE` config.
- Add the `IMGPROXY_HEALTH_CHECK_PATH` config.

## [3.2.2] - 2022-02-08
### Fixed
- Fix the `IMGPROXY_AVIF_SPEED` config.

## [3.2.1] - 2022-01-19
### Fixed
- Fix support of BMP with unusual data offsets.

## [3.2.0] - 2022-01-18
### Added
- (pro) Add `video_meta` to the `/info` response.
- Add [zoom](https://docs.imgproxy.net/generating_the_url?id=zoom) processing option.
- Add 1/2/4-bit BMP support.

### Changed
- Optimized `crop`.

### Fixed
- Fix Datadog support.
- Fix `force` resizing + rotation.
- (pro) Fix `obj` gravity.

## [3.1.3] - 2021-12-17
### Fixed
- Fix ETag checking when S3 is used.

## [3.1.2] - 2021-12-15
### Fixed
- (pro) Fix object detection.

## [3.1.1] - 2021-12-10
### Fixed
- Fix crashes in some scenarios.

## [3.1.0] - 2021-12-08
### Added
- Add `IMGPROXY_ETAG_BUSTER` config.
- (pro) [watermark_text](https://docs.imgproxy.net/generating_the_url?id=watermark-text) processing option.

### Changed
- Improved ICC profiles handling.
- Proper error message when the deprecated basic URL format is used.
- Watermark offsets can be applied to replicated watermarks.

### Fixed
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
- [Datadog](https://docs.imgproxy.net/datadog) metrics.
- `force` and `fill-down` resizing types.
- [min-width](https://docs.imgproxy.net/generating_the_url?id=min-width) and [min-height](https://docs.imgproxy.net/generating_the_url?id=min-height) processing options.
- [format_quality](https://docs.imgproxy.net/generating_the_url?id=format-quality) processing option.
- Add `X-Origin-Width` and `X-Origin-Height` to debug headers.
- Add `IMGPROXY_COOKIE_PASSTHROUGH` and `IMGPROXY_COOKIE_BASE_URL` configs.
- Add `client_ip` to requests and responses logs.

### Changed
- ETag generator & checker uses source image ETag when possible.
- `304 Not Modified` responses includes `Cache-Control`, `Expires`, and `Vary` headers.
- `dpr` processing option doesn't enlarge image unless `enlarge` is true.
- imgproxy responds with `500` HTTP code when the source image downloading error seems temporary (timeout, server error, etc).
- When `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` is zero, imgproxy responds with the usual HTTP code.
- BMP support doesn't require ImageMagick.
- Save GIFs without ImageMagick (vips 8.12+ required).

### Fixed
- Fix Client Hints behavior. `Width` is physical size, so we should divide it by `DPR` value.
- Fix scale-on-load in some rare cases.
- Fix the default Sentry release name.
- Fix the `health` command when the path prefix is set.
- Escape double quotes in content disposition.

### Removed
- Removed basic URL format, use [advanced one](https://docs.imgproxy.net/generating_the_url) instead.
- Removed `IMGPROXY_MAX_SRC_DIMENSION` config, use `IMGPROXY_MAX_SRC_RESOLUTION` instead.
- Removed `IMGPROXY_GZIP_COMPRESSION` config.
- Removed `IMGPROXY_MAX_GIF_FRAMES` config, use `IMGPROXY_MAX_ANIMATION_FRAMES` instead.
- Removed `crop` resizing type, use [crop](https://docs.imgproxy.net/generating_the_url#crop) processing option instead.
- Dropped old libvips (<8.10) support.
- (pro) Removed advanced GIF optimizations. All optimizations are applied by default ib both OSS and Pro versions.

## [3.0.0.beta2] - 2021-11-15
### Added
- Add `X-Origin-Width` and `X-Origin-Height` to debug headers.
- Add `IMGPROXY_COOKIE_PASSTHROUGH` and `IMGPROXY_COOKIE_BASE_URL` configs.

### Changed
- `dpr` processing option doesn't enlarge image unless `enlarge` is true.
- `304 Not Modified` responses includes `Cache-Control`, `Expires`, and `Vary` headers.
- imgproxy responds with `500` HTTP code when the source image downloading error seems temporary (timeout, server error, etc).
- When `IMGPROXY_FALLBACK_IMAGE_HTTP_CODE` is zero, imgproxy responds with the usual HTTP code.
- BMP support doesn't require ImageMagick.

### Fixed
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
- [Datadog](https://docs.imgproxy.net/datadog) metrics.
- `force` and `fill-down` resizing types.
- [min-width](https://docs.imgproxy.net/generating_the_url?id=min-width) and [min-height](https://docs.imgproxy.net/generating_the_url?id=min-height) processing options.
- [format_quality](https://docs.imgproxy.net/generating_the_url?id=format-quality) processing option.

### Changed
- ETag generator & checker uses source image ETag when possible.

### Removed
- Removed basic URL format, use [advanced one](https://docs.imgproxy.net/generating_the_url) instead.
- Removed `IMGPROXY_MAX_SRC_DIMENSION` config, use `IMGPROXY_MAX_SRC_RESOLUTION` instead.
- Removed `IMGPROXY_GZIP_COMPRESSION` config.
- Removed `IMGPROXY_MAX_GIF_FRAMES` config, use `IMGPROXY_MAX_ANIMATION_FRAMES` instead.
- Removed `crop` resizing type, use [crop](https://docs.imgproxy.net/generating_the_url#crop) processing option instead.
- Dropped old libvips (<8.8) support.

## [2.17.0] - 2021-09-07
### Added
- Wildcard support in `IMGPROXY_ALLOWED_SOURCES`.

### Changed
- If the source URL contains the `IMGPROXY_BASE_URL` prefix, it won't be added.

### Fixed
- (pro) Fix path prefix support in the `/info` handler.

### Deprecated
- The [basic URL format](https://docs.imgproxy.net/generating_the_url_basic) is deprecated and can be removed in future versions. Use [advanced URL format](https://docs.imgproxy.net/generating_the_url_advanced) instead.

## [2.16.7] - 2021-07-20
### Changed
- Reset DPI while stripping meta.

## [2.16.6] - 2021-07-08
### Fixed
- Fix performance regression in ICC profile handling.
- Fix crashes while processing CMYK images without ICC profile.

## [2.16.5] - 2021-06-28
### Changed
- More clear downloading errors.

### Fixed
- Fix ICC profile handling in some cases.
- Fix handling of negative height value for BMP.

## [2.16.4] - 2021-06-16
### Changed
- Use magenta (ff00ff) as a transparency key in `trim`.

### Fixed
- Fix crashes while processing some SVGs (dockerized version).
- (pro) Fix parsing HEIF/AVIF meta.

## [2.16.3] - 2021-04-05
### Fixed
- Fix PNG quantization palette size.
- Fix parsing HEIF meta.
- Fix debig header.

## [2.16.2] - 2021-03-04
### Changed
- Updated dependencies in Docker.

## [2.16.1] - 2021-03-02
### Fixed
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

### Fixed
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

### Fixed
- Fix `padding` and `extend` behaior when converting from a fromat without alpha support to one with alpha support.
- Fix WebP dimension limit handling.
- (pro) Fix thumbnails generation of some videos.

## [2.14.1] - 2020-07-22
### Fixed
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

### Fixed
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
- [filename](https://docs.imgproxy.net/generating_the_url#filename) option.

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
- [HEIC support](https://docs.imgproxy.net/image_formats_support#heic-support);
- [crop](https://docs.imgproxy.net/generating_the_url#crop) processing option. `resizing_type:crop` is deprecated;
- Offsets for [gravity](https://docs.imgproxy.net/generating_the_url#gravity);
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
- [extend](https://docs.imgproxy.net/generating_the_url#extend) processing option.
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
- Optimized memory usage. [Memory usage tweaks](https://docs.imgproxy.net/memory_usage_tweaks).
- `Vary` header is set when WebP detection, client hints or GZip compression are enabled.
- Health check doesn't require `Authorization` header anymore.

## [2.1.5] - 2019-01-14
### Added
- [Sentry support](https://docs.imgproxy.net/configuration#error-reporting) (thanks to [@koenpunt](https://github.com/koenpunt)).
- [Syslog support](https://docs.imgproxy.net/configuration#syslog).

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
- [Minio support](https://docs.imgproxy.net/serving_files_from_s3#minio)

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
- [Plain source URLs](https://docs.imgproxy.net/generating_the_url#plain) support.
- [Serving images from Google Cloud Storage](https://docs.imgproxy.net/serving_files_from_google_cloud_storage).
- [Full support of GIFs](https://docs.imgproxy.net/image_formats_support#gif-support) including animated ones.
- [Watermarks](https://docs.imgproxy.net/watermark).
- [New Relic](https://docs.imgproxy.net/new_relic) metrics.
- [Prometheus](https://docs.imgproxy.net/prometheus) metrics.
- [DPR](https://docs.imgproxy.net/generating_the_url#dpr) option (thanks to [selul](https://github.com/selul)).
- [Cache buster](https://docs.imgproxy.net/generating_the_url#cache-buster) option.
- [Quality](https://docs.imgproxy.net/generating_the_url#quality) option.
- Support for custom [Amazon S3](https://docs.imgproxy.net/serving_files_from_s3) endpoints.
- Support for [Amazon S3](https://docs.imgproxy.net/serving_files_from_s3) versioning.
- [Client hints](https://docs.imgproxy.net/configuration#client-hints-support) support (thanks to [selul](https://github.com/selul)).
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
- [New advanced URL format](https://docs.imgproxy.net/generating_the_url). Unleash the full power of imgproxy v2.0.
- [Presets](https://docs.imgproxy.net/presets). Shorten your urls by reusing processing options.
- [Serving images from Amazon S3](https://docs.imgproxy.net/serving_files_from_s3). Thanks to [@crohr](https://github.com/crohr), now we have a way to serve files from private S3 buckets.
- [Autoconverting to WebP when supported by browser](https://docs.imgproxy.net/configuration#avifwebp-support-detection) (disabled by default). Use WebP as resulting format when browser supports it.
- [Gaussian blur](https://docs.imgproxy.net/generating_the_url#blur) and [sharpen](https://docs.imgproxy.net/generating_the_url#sharpen) filters. Make your images look better than before.
- [Focus point gravity](https://docs.imgproxy.net/generating_the_url#gravity). Tell imgproxy what point will be the center of the image.
- [Background color](https://docs.imgproxy.net/generating_the_url#background). Control the color of background when converting PNG with alpha-channel to JPEG.

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
