# Image formats support

At the moment, imgproxy supports only the most popular Web image formats:

* PNG;
* JPEG;
* WebP;
* GIF.

## GIF support

imgproxy supports GIF output only when using libvips 8.7.0+ compiled with ImageMagick support. Official imgproxy Docker image supports GIF out of the box.

Since processing of animated GIFs is pretty heavy, only one frame is processed by default. You can increase the maximum of GIF frames to process with the following variable:

* `IMGPROXY_MAX_GIF_FRAMES`: the maximum of animated GIF frames to being processed. Default: `1`.

**Note:** imgproxy summarizes all GIF frames resolutions while checking source image resolution.
