# Watermark

imgproxy supports watermarking processed images with another image.

## Specifying watermark image

There are three ways to specify a watermark image using environment variables:

* `IMGPROXY_WATERMARK_DATA` - Base64-encoded image data. You can easily calculate it with `base64 tmp/watermark.png | tr -d '\n'`.
* `IMGPROXY_WATERMARK_PATH` - path to the locally stored image.
* `IMGPROXY_WATERMARK_URL` - watermark image URL.

You can also specify the base opacity of watermark with `IMGPROXY_WATERMARK_OPACITY`.

## Watermarking an image

Watermarks are only available with [advanced URL format](generating_the_url_advanced.md). Use `watermark` processing option to put the watermark on the processed image:

```
watermark:%opacity:%position:%x_offset:%y_offset
wm:%opacity:%position:%x_offset:%y_offset
```

Where arguments are:

* `opacity` - watermark opacity modifier. Final opacity is calculated like `base_opacity * opacity`. It's highly recommended to set this argument as `1` and adjust opacity with `IMGPROXY_WATERMARK_OPACITY` since this would optimize performance and memory usage.
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
* `x_offset`, `y_offset` - (optional) specify watermark offset by X and Y axes. Not applicable to `re` position.
