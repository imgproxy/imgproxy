# Image formats support

At the moment, imgproxy supports only the most popular image formats:

* PNG;
* JPEG;
* WebP;
* GIF;
* ICO;
* SVG;
* MP4 <img class="pro-badge" src="assets/pro.svg" alt="pro" />;
* HEIC _(source only)_;
* BMP;
* TIFF.

## GIF support

imgproxy supports GIF output only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports GIF out of the box.

## ICO support

imgproxy supports ICO only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports ICO out of the box.

## SVG support

imgproxy supports SVG sources without limitations, but SVG results are not supported when the source image is not SVG.

When the source image is SVG and the SVG result is requested, imgproxy returns source image without modifications.

imgproxy reads some amount of bytes to check if the source image is SVG. By default it reads maximum of 32KB, but you can change this:

* `IMGPROXY_MAX_SVG_CHECK_BYTES`: the maximum number of bytes imgproxy will read to recognize SVG. If imgproxy can't recognize your SVG, try to increase this number. Default: `32768` (32KB)

## HEIC support

imgproxy supports HEIC only when using libvips 8.8.0+. Official imgproxy Docker image supports HEIC out of the box.

## BMP support

imgproxy supports BMP only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports ICO out of the box.

By default, imgproxy saves BMP images as JPEG. You need to explicitly specify the `format` option to get BMP output.

## Animated images support

Since processing of animated images is pretty heavy, only one frame is processed by default. You can increase the maximum of animation frames to process with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum of animated image frames to being processed. Default: `1`.

**📝Note:** imgproxy summarizes all frames resolutions while checking source image resolution.

## Converting animated images to MP4 <img class="pro-badge" src="assets/pro.svg" alt="pro" />

Animated images results can be converted to MP4 by specifying `mp4` extension.

Since MP4 requires usage of a `<video>` tag instead of `<img>`, automatic conversion to MP4 is not provided.

## Video thumbnails <img class="pro-badge" src="assets/pro.svg" alt="pro" />

If you provide a video as a source, imgproxy takes its specific frame to create a thumbnail. Doing this imgproxy downloads only the amount of data required to reach the needed frame.

Since this still requires more data to be downloaded, video thumbnails generation is disabled by default and should be enabled with `IMGPROXY_ENABLE_VIDEO_THUMBNAILS` config option.

* `IMGPROXY_ENABLE_VIDEO_THUMBNAILS`: when true, enables video thumbnails generation. Default: false;
* `IMGPROXY_VIDEO_THUMBNAIL_SECOND`: the timestamp of the frame in seconds that will be used for a thumbnail. Default: 1.
