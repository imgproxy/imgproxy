# Best format![pro](./assets/pro.svg)

You can use the `best` value for the [format](generating_the_url.md#format) option or the [extension](generating_the_url.md#extension) to make imgproxy pick the best format for the resultant image.

imgproxy measures the complexity of the image to choose when it should use a lossless or near-lossless encoding. Then imgproxy tries to save the image in multiple formats to pick one with the smallest resulting size.

**📝 Note:** imgproxy uses only the formats listed as [preferred](configuration.md#preferred-formats) when choosing the best format. It may also use AVIF or WebP if [AVIF/WebP support detection](configuration.md#avifwebp-support-detection) is enabled.

**📝 Note:** imgproxy will use AVIF or WebP _only_ if [AVIF/WebP support detection](configuration.md#avifwebp-support-detection) is enabled.

**📝 Note:** imgproxy may change your quality and autoquality settings if the `best` format is used.

## Configuration

* `IMGPROXY_BEST_FORMAT_COMPLEXITY_THRESHOLD `: the image complexity threshold. imgproxy will use a lossless or near-lossless encoding for images with low complexity. Default: `5.5`
* `IMGPROXY_BEST_FORMAT_MAX_RESOLUTION`: when greater than `0` and the image's resolution (in megapixels) is larger than the provided value, imgproxy won't try all the applicable formats and will just pick one that seems the best for the image
* `IMGPROXY_BEST_FORMAT_BY_DEFAULT`: when `true` and the resulting image format is not specified explicitly, imgproxy will use the `best` format instead of the source image format
* `IMGPROXY_BEST_FORMAT_ALLOW_SKIPS`: when `true` and the `best` format is used, imgproxy will skip processing of SVG and formats [listed to skip processing](configuration.md#skip-processing)
