# Image formats support

At the moment, imgproxy supports only the most popular image formats:

* PNG;
* JPEG;
* WebP;
* GIF;
* ICO;
* SVG;
* MP4 _(result only)_ <img class="pro-badge" src="assets/pro.svg" alt="pro" />;
* HEIC;
* BMP;
* TIFF.

## GIF support

imgproxy supports GIF output only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports GIF out of the box.

## ICO support

imgproxy supports ICO only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports ICO out of the box.

## SVG support

imgproxy supports SVG sources without limitations, but SVG results are not supported when the source image is not SVG.

When the source image is SVG and the SVG result is requested, imgproxy returns source image without modifications.

## HEIC support

imgproxy supports HEIC only when using libvips 8.8.0+. Official imgproxy Docker image supports HEIC out of the box.

By default, imgproxy saves HEIC images as JPEG. You need to explicitly specify the `format` option to get HEIC output.

## BMP support

imgproxy supports BMP only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports ICO out of the box.

By default, imgproxy saves BMP images as JPEG. You need to explicitly specify the `format` option to get BMP output.

## Animated images support

Since processing of animated images is pretty heavy, only one frame is processed by default. You can increase the maximum of animation frames to process with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum of animated image frames to being processed. Default: `1`.

**Note:** imgproxy summarizes all frames resolutions while checking source image resolution.

## Converting animated images to MP4 <img class="pro-badge" src="assets/pro.svg" alt="pro" />

Animated images results can be converted to MP4 by specifying `mp4` extension.

Since MP4 requires usage of a `<video>` tag instead of `<img>`, automatic conversion to MP4 is not provided.
