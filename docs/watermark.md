# Watermark

imgproxy supports watermarking processed images with another image.

## Specifying watermark image

There are three ways to specify a watermark image using environment variables:

* `IMGPROXY_WATERMARK_PATH`: path to the locally stored image.
* `IMGPROXY_WATERMARK_URL`: watermark image URL.
* `IMGPROXY_WATERMARK_DATA`: Base64-encoded image data. You can easily calculate it with the following snippet:
  ```bash
  base64 tmp/watermark.webp | tr -d '\n'`.
  ```

You can also specify the base opacity of watermark with `IMGPROXY_WATERMARK_OPACITY`.

**📝Note:** If you're going to use `scale` argument of `watermark`, it's highly recommended to use SVG, WebP or JPEG watermarks since these formats support scale-on-load.

## Watermarking an image

Watermarks are only available with [advanced URL format](generating_the_url_advanced.md). Use `watermark` processing option to put the watermark on the processed image:

```
watermark:%opacity:%position:%x_offset:%y_offset:%scale
wm:%opacity:%position:%x_offset:%y_offset:%scale
```

Where arguments are:

* `opacity` - watermark opacity modifier. Final opacity is calculated like `base_opacity * opacity`.
* `position` - (optional) specifies the position of the watermark. Available values:
  * `ce`: (default) center;
  * `no`: north (top edge);
  * `so`: south (bottom edge);
  * `ea`: east (right edge);
  * `we`: west (left edge);
  * `noea`: north-east (top-right corner);
  * `nowe`: north-west (top-left corner);
  * `soea`: south-east (bottom-right corner);
  * `sowe`: south-west (bottom-left corner);
  * `re`: replicate watermark to fill the whole image;
* `x_offset`, `y_offset` - (optional) specify watermark offset by X and Y axes. Not applicable to `re` position;
* `scale` - (optional) floating point number that defines watermark size relative to the resulting image size. When set to `0` or omitted, watermark size won't be changed.

## Custom watermarks <img class="pro-badge" src="assets/pro.svg" alt="pro" />

You can use a custom watermark specifying its URL with `watermark_url` processing option:

```
watermark_url:%url
wmu:%url
```

Where `url` is Base64-encoded URL of the custom watermark.

By default imgproxy caches 256 custom watermarks with adaptive replacement cache (ARC). You can change the cache size with `IMGPROXY_WATERMARKS_CACHE_SIZE` environment variable. When `IMGPROXY_WATERMARKS_CACHE_SIZE` is set to `0`, the cache is disabled.

