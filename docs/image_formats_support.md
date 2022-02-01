# Image formats support

At the moment, imgproxy supports only the most popular image formats:

| Format | Extension | Source | Result |
| -------|-----------|--------|--------|
| PNG    | `png`     | Yes    | Yes    |
| JPEG   | `jpg`     | Yes    | Yes    |
| WebP   | `webp`    | Yes    | Yes    |
| AVIF   | `avif`    | Yes    | Yes    |
| GIF    | `gif`     | Yes    | Yes    |
| ICO    | `ico`     | Yes    | Yes    |
| SVG    | `svg`     | Yes    | [See notes](#svg-support) |
| HEIC   | `heic`    | Yes    | No     |
| BMP    | `bmp`     | Yes    | Yes    |
| TIFF   | `tiff`    | Yes    | Yes    |
| PDF<i class='badge badge-pro'></i> | `pdf` | Yes | No |
| MP4 (h264)<i class='badge badge-pro'></i> | `mp4` | [See notes](#video-thumbnails) | Yes |
| Other video formats<i class='badge badge-pro'></i> | | [See notes](#video-thumbnails) | No |

## SVG support

imgproxy supports SVG sources without limitations, but SVG results are not supported when the source image is not SVG.

When the source image is SVG and an SVG result is requested, imgproxy returns the source image without modifications.

imgproxy reads some amount of bytes to check if the source image is SVG. By default it reads a maximum of 32KB, but you can change this:

* `IMGPROXY_MAX_SVG_CHECK_BYTES`: the maximum number of bytes imgproxy will read to recognize SVG. If imgproxy can't recognize your SVG, try to increase this number. Default: `32768` (32KB)

## Animated images support

Since the processing of animated images is a pretty heavy process, only one frame is processed by default. You can increase the maximum of animation frames to process with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum of animated image frames to be processed. Default: `1`.

**📝Note:** imgproxy summarizes all frames resolutions while the checking source image resolution.

## Converting animated images to MP4<i class='badge badge-pro'></i> :id=converting-animated-images-to-mp4

Animated image results can be converted to MP4 by specifying the `mp4` extension.

Since MP4 requires use of a `<video>` tag instead of `<img>`, automatic conversion to MP4 is not provided.

## Video thumbnails<i class='badge badge-pro'></i> :id=video-thumbnails

If you provide a video as a source, imgproxy takes a specific frame to create a thumbnail. To do this, imgproxy downloads only the amount of data required to reach the needed frame.

Since this still requires more data to be downloaded, video thumbnail generation is disabled by default and should be enabled with `IMGPROXY_ENABLE_VIDEO_THUMBNAILS` config option.

* `IMGPROXY_ENABLE_VIDEO_THUMBNAILS`: when true, enables video thumbnail generation. Default: `false`
* `IMGPROXY_VIDEO_THUMBNAIL_SECOND`: the timestamp of the frame (in seconds) that will be used for the thumbnail. Default: 1.
