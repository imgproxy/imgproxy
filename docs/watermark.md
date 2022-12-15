# Watermark

imgproxy supports the watermarking of processed images using another image.

## Specifying watermark image

There are three ways to specify a watermark image using environment variables:

* `IMGPROXY_WATERMARK_PATH`: the path to the locally stored image
* `IMGPROXY_WATERMARK_URL`: the watermark image URL
* `IMGPROXY_WATERMARK_DATA`: Base64-encoded image data. You can easily calculate it with the following snippet:
  ```bash
  base64 tmp/watermark.webp | tr -d '\n'`.
  ```

You can also specify the base opacity of a watermark using `IMGPROXY_WATERMARK_OPACITY`.

**📝Note:** If you're going to use the `scale` argument of `watermark`, it's highly recommended to use SVG, WebP or JPEG watermarks since these formats support scale-on-load.

## Watermarking an image

Use the `watermark` processing option to put a watermark on a processed image:

```
watermark:%opacity:%position:%x_offset:%y_offset:%scale
wm:%opacity:%position:%x_offset:%y_offset:%scale
```

The available arguments are:

* `opacity` - watermark opacity modifier. The final opacity is calculated as `base_opacity * opacity`.
* `position` - (optional) specifies the position of the watermark. Available values:
  * `ce`: (default) center
  * `no`: north (top edge)
  * `so`: south (bottom edge)
  * `ea`: east (right edge)
  * `we`: west (left edge)
  * `noea`: north-east (top-right corner)
  * `nowe`: north-west (top-left corner)
  * `soea`: south-east (bottom-right corner)
  * `sowe`: south-west (bottom-left corner)
  * `re`: repeat and tile the watermark to fill the entire image
* `x_offset`, `y_offset` - (optional) specify watermark offset by X and Y axes. When using `re` position, these values define the spacing between the tiles.
* `scale` - (optional) a floating point number that defines the watermark size relative to the resulting image size. When set to `0` or omitted, the watermark size won't be changed.

## Custom watermarks![pro](./assets/pro.svg) :id=custom-watermarks

You can use a custom watermark by specifying its URL with the `watermark_url` processing option:

```
watermark_url:%url
wmu:%url
```

The value of `url` should be the Base64-encoded URL of the custom watermark.

By default, imgproxy caches 256 custom watermarks with an adaptive replacement cache (ARC). You can change the cache size using the `IMGPROXY_WATERMARKS_CACHE_SIZE` environment variable. When `IMGPROXY_WATERMARKS_CACHE_SIZE` is set to `0`, the cache is disabled.

