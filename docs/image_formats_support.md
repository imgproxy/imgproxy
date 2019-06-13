# Image formats support

At the moment, imgproxy supports only the most popular Web image formats:

* PNG;
* JPEG;
* WebP;
* GIF;
* ICO;
* SVG _(source only)_;
* HEIC.

## GIF support

imgproxy supports GIF output only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports GIF out of the box.

## ICO support

imgproxy supports ICO output only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports ICO out of the box.

## Animated images support

Since processing of animated images is pretty heavy, only one frame is processed by default. You can increase the maximum of animation frames to process with the following variable:

* `IMGPROXY_MAX_ANIMATION_FRAMES`: the maximum of animated image frames to being processed. Default: `1`.

**Note:** imgproxy summarizes all frames resolutions while checking source image resolution.
