# Generating the URL

The URL should contain the signature, processing options, and source URL, like this:

```
/%signature/%processing_options/plain/%source_url@%extension
/%signature/%processing_options/%encoded_source_url.%extension
```

Check out the [example](#example) at the end of this guide.

## Signature

A signature protects your URL from being altered by an attacker. It is highly recommended to sign imgproxy URLs when imgproxy is being used in production.

Once you set up your [URL signature](configuration.md#url-signature), check out the [Signing the URL](signing_the_url.md) guide to find out how to sign your URLs. Otherwise, since the signature still needs to be present, feel free to use any string here.

## Processing options

Processing options should be specified as URL parts divided by slashes (`/`). A processing option has the following format:

```
%option_name:%argument1:%argument2:...:argumentN
```

The list of processing options does not define imgproxy's processing pipeline. Instead, imgproxy already comes with a specific, built-in image processing pipeline for maximum performance. Read more about this in the [About processing pipeline](about_processing_pipeline.md) guide.

imgproxy supports the following processing options:

### Resize

```
resize:%resizing_type:%width:%height:%enlarge:%extend
rs:%resizing_type:%width:%height:%enlarge:%extend
```

This is a meta-option that defines the [resizing type](#resizing-type), [width](#width), [height](#height), [enlarge](#enlarge), and [extend](#extend). All arguments are optional and can be omitted to use their default values.

### Size

```
size:%width:%height:%enlarge:%extend
s:%width:%height:%enlarge:%extend
```

This is a meta-option that defines the [width](#width), [height](#height), [enlarge](#enlarge), and [extend](#extend). All arguments are optional and can be omitted to use their default values.

### Resizing type

```
resizing_type:%resizing_type
rt:%resizing_type
```

Defines how imgproxy will resize the source image. Supported resizing types are:

* `fit`: resizes the image while keeping aspect ratio to fit a given size.
* `fill`: resizes the image while keeping aspect ratio to fill a given size and crops projecting parts.
* `fill-down`: the same as `fill`, but if the resized image is smaller than the requested size, imgproxy will crop the result to keep the requested aspect ratio.
* `force`: resizes the image without keeping the aspect ratio.
* `auto`: if both source and resulting dimensions have the same orientation (portrait or landscape), imgproxy will use `fill`. Otherwise, it will use `fit`.

Default: `fit`

### Resizing algorithm![pro](./assets/pro.svg) :id=resizing-algorithm

```
resizing_algorithm:%algorithm
ra:%algorithm
```

Defines the algorithm that imgproxy will use for resizing. Supported algorithms are `nearest`, `linear`, `cubic`, `lanczos2`, and `lanczos3`.

Default: `lanczos3`

### Width

```
width:%width
w:%width
```

Defines the width of the resulting image. When set to `0`, imgproxy will calculate width using the defined height and source aspect ratio. When set to `0` and resizing type is `force`, imgproxy will keep the original width.

Default: `0`

### Height

```
height:%height
h:%height
```

Defines the height of the resulting image. When set to `0`, imgproxy will calculate resulting height using the defined width and source aspect ratio. When set to `0` and resizing type is `force`, imgproxy will keep the original height.

Default: `0`

### Min width

```
min-width:%width
mw:%width
```

Defines the minimum width of the resulting image.

**⚠️Warning:** When both `width` and `min-width` are set, the final image will be cropped according to `width`, so use this combination with care.

Default: `0`

### Min height

```
min-height:%height
mh:%height
```

Defines the minimum height of the resulting image.

**⚠️Warning:** When both `height` and `min-height` are set, the final image will be cropped according to `height`, so use this combination with care.

Default: `0`

### Zoom

```
zoom:%zoom_x_y
z:%zoom_x_y

zoom:%zoom_x:%zoom_y
z:%zoom_x:%zoom_y
```

When set, imgproxy will multiply the image dimensions according to these factors. The values must be greater than 0.

Can be combined with `width` and `height` options. In this case, imgproxy calculates scale factors for the provided size and then multiplies it with the provided zoom factors.

**📝Note:** Unlike [dpr](#dpr), `zoom` doesn't set the `Content-DPR` header in the response.

Default: `1`

### Dpr

```
dpr:%dpr
```

When set, imgproxy will multiply the image dimensions according to this factor for HiDPI (Retina) devices. The value must be greater than 0.

**📝Note:** `dpr` also sets the `Content-DPR` header in the response so the browser can correctly render the image.

Default: `1`

### Enlarge

```
enlarge:%enlarge
el:%enlarge
```

When set to `1`, `t` or `true`, imgproxy will enlarge the image if it is smaller than the given size.

Default: `false`

### Extend

```
extend:%extend:%gravity
ex:%extend:%gravity
```

* When `extend` is set to `1`, `t` or `true`, imgproxy will extend the image if it is smaller than the given size.
* `gravity` _(optional)_ accepts the same values as the [gravity](#gravity) option, except `sm`. When `gravity` is not set, imgproxy will use `ce` gravity without offsets.

Default: `false:ce:0:0`

### Extend aspect ratio

```
extend_aspect_ratio:%extend:%gravity
extend_ar:%extend:%gravity
exar:%extend:%gravity
```

* When `extend` is set to `1`, `t` or `true`, imgproxy will extend the image to the requested aspect ratio.
* `gravity` _(optional)_ accepts the same values as the [gravity](#gravity) option, except `sm`. When `gravity` is not set, imgproxy will use `ce` gravity without offsets.

Default: `false:ce:0:0`

### Gravity

```
gravity:%type:%x_offset:%y_offset
g:%type:%x_offset:%y_offset
```

When imgproxy needs to cut some parts of the image, it is guided by the gravity option.

* `type` - specifies the gravity type. Available values:
  * `no`: north (top edge)
  * `so`: south (bottom edge)
  * `ea`: east (right edge)
  * `we`: west (left edge)
  * `noea`: north-east (top-right corner)
  * `nowe`: north-west (top-left corner)
  * `soea`: south-east (bottom-right corner)
  * `sowe`: south-west (bottom-left corner)
  * `ce`: center
* `x_offset`, `y_offset` - (optional) specifies the gravity offset along the X and Y axes.

Default: `ce:0:0`

**Special gravities**:

* `gravity:sm`: smart gravity. `libvips` detects the most "interesting" section of the image and considers it as the center of the resulting image. Offsets are not applicable here.
* `gravity:obj:%class_name1:%class_name2:...:%class_nameN`: ![pro](./assets/pro.svg) object-oriented gravity. imgproxy [detects objects](object_detection.md) of provided classes on the image and calculates the resulting image center using their positions. If class names are omited, imgproxy will use all the detected objects.
* `gravity:fp:%x:%y`: the gravity focus point . `x` and `y` are floating point numbers between 0 and 1 that define the coordinates of the center of the resulting image. Treat 0 and 1 as right/left for `x` and top/bottom for `y`.

### Crop

```
crop:%width:%height:%gravity
c:%width:%height:%gravity
```

Defines an area of the image to be processed (crop before resize).

* `width` and `height` define the size of the area:
  * When `width` or `height` is greater than or equal to `1`, imgproxy treats it as an absolute value.
  * When `width` or `height` is less than `1`, imgproxy treats it as a relative value.
  * When `width` or `height` is set to `0`, imgproxy will use the full width/height of the source image.
* `gravity` _(optional)_ accepts the same values as the [gravity](#gravity) option. When `gravity` is not set, imgproxy will use the value of the [gravity](#gravity) option.

### Trim

```
trim:%threshold:%color:%equal_hor:%equal_ver
t:%threshold:%color:%equal_hor:%equal_ver
```

Removes surrounding background.

* `threshold` - color similarity tolerance.
* `color` - _(optional)_ a hex-coded value of the color that needs to be cut off.
* `equal_hor` - _(optional)_ set to `1`, `t` or `true`, imgproxy will cut only equal parts from left and right sides. That means that if 10px of background can be cut off from the left and 5px from the right, then 5px will be cut off from both sides. For example, this can be useful if objects on your images are centered but have non-symmetrical shadow.
* `equal_ver` - _(optional)_ acts like `equal_hor` but for top/bottom sides.

**⚠️Warning:** Trimming requires an image to be fully loaded into memory. This disables scale-on-load and significantly increases memory usage and processing time. Use it carefully with large images.

**📝Note:** If you know background color of your images then setting it explicitly via `color` will also save some resources because imgproxy won't need to automatically detect it.

**📝Note:** Use a `color` value of `FF00FF` for trimming transparent backgrounds as imgproxy uses magenta as a transparency key.

**📝Note:** The trimming of animated images is not supported.

### Padding

```
padding:%top:%right:%bottom:%left
pd:%top:%right:%bottom:%left
```

Defines padding size using CSS-style syntax. All arguments are optional but at least one dimension must be set. Padded space is filled according to the [background](#background) option.

* `top` - top padding (and for all other sides if they haven't been explicitly st)
* `right` - right padding (and left if it hasn't been explicitly set)
* `bottom` - bottom padding
* `left` - left padding

**📝Note:** Padding is applied after all image transformations (except watermarking) and enlarges the generated image. This means that if your resize dimensions were 100x200px and you applied the `padding:10` option, then you will end up with an image with dimensions of 120x220px.

**📝Note:** Padding follows the [dpr](#dpr) option so it will also be scaled if you've set it.

### Auto rotate

```
auto_rotate:%auto_rotate
ar:%auto_rotate
```

When set to `1`, `t` or `true`, imgproxy will automatically rotate images based on the EXIF Orientation parameter (if available in the image meta data). The orientation tag will be removed from the image in all cases. Normally this is controlled by the [IMGPROXY_AUTO_ROTATE](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

### Rotate

```
rotate:%angle
rot:%angle
```

Rotates the image on the specified angle. The orientation from the image metadata is applied before the rotation unless autorotation is disabled.

**📝Note:** Only 0, 90, 180, 270, etc., degree angles are supported.

Default: 0

### Background

```
background:%R:%G:%B
bg:%R:%G:%B

background:%hex_color
bg:%hex_color
```

When set, imgproxy will fill the resulting image background with the specified color. `R`, `G`, and `B` are the red, green and blue channel values of the background color (0-255). `hex_color` is a hex-coded value of the color. Useful when you convert an image with alpha-channel to JPEG.

With no arguments provided, disables any background manipulations.

Default: disabled

### Background alpha![pro](./assets/pro.svg) :id=background-alpha

```
background_alpha:%alpha
bga:%alpha
```

Adds an alpha channel to `background`. The value of `alpha` is a positive floating point number between `0` and `1`.

Default: 1

### Adjust![pro](./assets/pro.svg) :id=adjust

```
adjust:%brightness:%contrast:%saturation
a:%brightness:%contrast:%saturation
```

This is a meta-option that defines the [brightness](#brightness), [contrast](#contrast), and [saturation](#saturation). All arguments are optional and can be omitted to use their default values.

### Brightness![pro](./assets/pro.svg) :id=brightness

```
brightness:%brightness
br:%brightness
```

When set, imgproxy will adjust brightness of the resulting image. `brightness` is an integer number ranging from `-255` to `255`.

Default: 0

### Contrast![pro](./assets/pro.svg) :id=contrast

```
contrast:%contrast
co:%contrast
```

When set, imgproxy will adjust the contrast of the resulting image. `contrast` is a positive floating point number, where a value of `1` leaves the contrast unchanged.

Default: 1

### Saturation![pro](./assets/pro.svg) :id=saturation

```
saturation:%saturation
sa:%saturation
```

When set, imgproxy will adjust saturation of the resulting image. `saturation` is a positive floating-point number, where a value of `1` leaves the saturation unchanged.

Default: 1

### Blur

```
blur:%sigma
bl:%sigma
```

When set, imgproxy will apply a gaussian blur filter to the resulting image. The value of `sigma` defines the size of the mask imgproxy will use.

Default: disabled

### Sharpen

```
sharpen:%sigma
sh:%sigma
```

When set, imgproxy will apply the sharpen filter to the resulting image. The value of `sigma` defines the size of the mask imgproxy will use.

As an approximate guideline, use 0.5 sigma for 4 pixels/mm (display resolution), 1.0 for 12 pixels/mm and 1.5 for 16 pixels/mm (300 dpi == 12 pixels/mm).

Default: disabled

### Pixelate

```
pixelate:%size
pix:%size
```

When set, imgproxy will apply the pixelate filter to the resulting image. The value of `size` defines individual pixel size.

Default: disabled

### Unsharpening![pro](./assets/pro.svg) :id=unsharpening

```
unsharpening:%mode:%weight:%dividor
ush:%mode:%weight:%dividor
```

Allows redefining unsharpening options. All arguments have the same meaning as [Unsharpening](configuration.md#unsharpening) configs. All arguments are optional and can be omitted.

### Blur detections![pro](./assets/pro.svg) :id=blur-detections

```
blur_detections:%sigma:%class_name1:%class_name2:...:%class_nameN
bd:%sigma:%class_name1:%class_name2:...:%class_nameN
```

imgproxy [detects objects](object_detection.md) of the provided classes and blurs them. If class names are omitted, imgproxy blurs all the detected objects.

The value of `sigma` defines the size of the mask imgproxy will use.

### Draw detections![pro](./assets/pro.svg) :id=draw-detections

```
draw_detections:%draw:%class_name1:%class_name2:...:%class_nameN
dd:%draw:%class_name1:%class_name2:...:%class_nameN
```

When `draw` is set to `1`, `t` or `true`, imgproxy [detects objects](object_detection.md) of the provided classes and draws their bounding boxes. If class names are omitted, imgproxy draws the bounding boxes of all the detected objects.

### Gradient![pro](./assets/pro.svg) :id=gradient

```
gradient:%opacity:%color:%direction:%start%stop
gr:%opacity:%color:%direction:%start%stop
```

Places a gradient on the processed image. The placed gradient transitions from transparency to the specified color.

* `opacity`: specifies gradient opacity. When set to `0`, gradient is not applied.
* `color`:  _(optional)_ a hex-coded value of the gradient color. Default: `000` (black).
* `direction`: _(optional)_ specifies the direction of the gradient. Available values:
  * `down`: _(default)_ the top side of the gradient is transparrent, the bottom side is opaque
  * `up`: the bottom side of the gradient is transparrent, the top side is opaque
  * `right`: the left side of the gradient is transparrent, the right side is opaque
  * `left`: the right side of the gradient is transparrent, the left side is opaque
* `start`, `stop`: floating point numbers that define relative positions of where the gradient starts and where it ends. Default values are `0.0` and `1.0` respectively.

### Watermark

```
watermark:%opacity:%position:%x_offset:%y_offset:%scale
wm:%opacity:%position:%x_offset:%y_offset:%scale
```

Places a watermark on the processed image.

* `opacity`: watermark opacity modifier. Final opacity is calculated like `base_opacity * opacity`.
* `position`: (optional) specifies the position of the watermark. Available values:
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
* `scale`: (optional) a floating-point number that defines the watermark size relative to the resultant image size. When set to `0` or when omitted, the watermark size won't be changed.

Default: disabled

### Watermark URL![pro](./assets/pro.svg) :id=watermark-url

```
watermark_url:%url
wmu:%url
```

When set, imgproxy will use the image from the specified URL as a watermark. `url` is the URL-safe Base64-encoded URL of the custom watermark.

Default: blank

### Watermark text![pro](./assets/pro.svg) :id=watermark-text

```
watermark_text:%text
wmt:%text
```

When set, imgproxy will generate an image from the provided text and use it as a watermark. `text` is the URL-safe Base64-encoded text of the custom watermark.

By default, the text color is black and the font is `sans 16`. You can use [Pango markup](https://docs.gtk.org/Pango/pango_markup.html) in the `text` value to change the style.

If you want to use a custom font, you need to put it in `/usr/share/fonts` inside a container.

Default: blank

### Watermark size![pro](./assets/pro.svg) :id=watermark-size

```
watermark_size:%width:%height
wms:%width:%height
```

Defines the desired width and height of the watermark. imgproxy always uses `fit` resizing type when resizing watermarks and enlarges them when needed.

When `%width` is set to `0`, imgproxy will calculate the width using the defined height and watermark's aspect ratio.

When `%height` is set to `0`, imgproxy will calculate the height using the defined width and watermark's aspect ratio.

**📝Note:** This processing option takes effect only when the `scale` argument of the `watermark` option is set to zero.

Default: `0:0`

### Watermark shadow![pro](./assets/pro.svg) :id=watermark-shadow

```
watermark_shadow:%sigma
wmsh:%sigma
```
When set, imgproxy will add a shadow to the watermark. The value of `sigma` defines the size of the mask imgproxy will use to blur the shadow.

Default: disabled

### Style![pro](./assets/pro.svg) :id=style

```
style:%style
st:%style
```

When set, imgproxy will prepend a `<style>` node with the provided content to the `<svg>` node of a source SVG image. `%style` is URL-safe Base64-encoded CSS-styles.

Default: blank

### Strip metadata

```
strip_metadata:%strip_metadata
sm:%strip_metadata
```

When set to `1`, `t` or `true`, imgproxy will strip the metadata (EXIF, IPTC, etc.) on JPEG and WebP output images. This is normally controlled by the [IMGPROXY_STRIP_METADATA](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

### Keep copyright

```
keep_copyright:%keep_copyright
kcr:%keep_copyright
```

When set to `1`, `t` or `true`, imgproxy will not remove copyright info while stripping metadata. This is normally controlled by the [IMGPROXY_KEEP_COPYRIGHT](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

### Strip color profile

```
strip_color_profile:%strip_color_profile
scp:%strip_color_profile
```

When set to `1`, `t` or `true`, imgproxy will transform the embedded color profile (ICC) to sRGB and remove it from the image. Otherwise, imgproxy will try to keep it as is. This is normally controlled by the [IMGPROXY_STRIP_COLOR_PROFILE](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

### Enforce thumbnail

```
enforce_thumbnail:%enforce_thumbnail
eth:%enforce_thumbnail
```

When set to `1`, `t` or `true` and the source image has an embedded thumbnail, imgproxy will always use the embedded thumbnail instead of the main image. Currently, only thumbnails embedded in `heic` and `avif` are supported. This is normally controlled by the [IMGPROXY_ENFORCE_THUMBNAIL](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

### Quality

```
quality:%quality
q:%quality
```

Redefines quality of the resulting image, as a percentage. When set to `0`, quality is assumed based on `IMGPROXY_QUALITY` and [format_quality](#format-quality).

Default: 0.

### Format quality

```
format_quality:%format1:%quality1:%format2:%quality2:...:%formatN:%qualityN
fq:%format1:%quality1:%format2:%quality2:...:%formatN:%qualityN
```

Adds or redefines `IMGPROXY_FORMAT_QUALITY` values.

### Autoquality![pro](./assets/pro.svg) :id=autoquality

```
autoquality:%method:%target:%min_quality:%max_quality:%allowed_error
aq:%method:%target:%min_quality:%max_quality:%allowed_error
```

Redefines autoquality settings. All arguments have the same meaning as [Autoquality](configuration.md#autoquality) configs. All arguments are optional and can be omitted.

**⚠️Warning:** Autoquality requires the image to be saved several times. Use it only when you prefer the resulting size and quality over the speed.

### Max bytes

```
max_bytes:%bytes
mb:%bytes
```

When set, imgproxy automatically degrades the quality of the image until the image size is under the specified amount of bytes.

**📝Note:** Applicable only to `jpg`, `webp`, `heic`, and `tiff`.

**⚠️Warning:** When `max_bytes` is set, imgproxy saves image multiple times to achieve the specified image size.

Default: 0

### JPEG options![pro](./assets/pro.svg) :id=jpeg-options

```
jpeg_options:%progressive:%no_subsample:%trellis_quant:%overshoot_deringing:%optimize_scans:%quant_table
jpgo:%progressive:%no_subsample:%trellis_quant:%overshoot_deringing:%optimize_scans:%quant_table
```

Allows redefining JPEG saving options. All arguments have the same meaning as the [Advanced JPEG compression](configuration.md#advanced-jpeg-compression) configs. All arguments are optional and can be omitted.

### PNG options![pro](./assets/pro.svg) :id=png-options

```
png_options:%interlaced:%quantize:%quantization_colors
pngo:%interlaced:%quantize:%quantization_colors
```

Allows redefining PNG saving options. All arguments have the same meaning as with the [Advanced PNG compression](configuration.md#advanced-png-compression) configs. All arguments are optional and can be omitted.

<!-- ### GIF options![pro](./assets/pro.svg) :id=gif-options

```
gif_options:%optimize_frames:%optimize_transparency
gifo:%optimize_frames:%optimize_transparency
```

Allows redefining GIF saving options. All arguments have the same meaning as with the [Advanced GIF compression](configuration.md#advanced-gif-compression) configs. All arguments are optional and can be omitted. -->

### WebP options![pro](./assets/pro.svg) :id=webp-options

```
webp_options:%compression
webpo:%compression
```

Allows redefining WebP saving options. All arguments have the same meaning as with the [Advanced WebP compression](configuration.md#advanced-webp-compression) configs. All arguments are optional and can be omitted.

### Format

```
format:%extension
f:%extension
ext:%extension
```

Specifies the resulting image format. Alias for the [extension](#extension) part of the URL.

Default: `jpg`

### Page![pro](./assets/pro.svg) :id=page

```
page:%page
pg:%page
```

When a source image supports pagination (PDF, TIFF) or animation (GIF, WebP), this option allows specifying the page to use it on. Page numeration starts from zero.

Default: 0

### Disable animation![pro](./assets/pro.svg) :id=disable-animation

```
disable_animation:%disable
da:%disable
```

When set to `1`, `t` or `true`, imgproxy will use a single frame of animated images. Use the [page](#page) option to specify which frame imgproxy should use.

Default: `false`

### Video thumbnail second![pro](./assets/pro.svg) :id=video-thumbnail-second

```
video_thumbnail_second:%second
vts:%second
```

Allows redefining `IMGPROXY_VIDEO_THUMBNAIL_SECOND` config.

### Fallback image URL![pro](./assets/pro.svg) :id=fallback-image-url

You can use a custom fallback image by specifying its URL with the `fallback_image_url` processing option:

```
fallback_image_url:%url
fiu:%url
```

The value of `url` is the URL-safe Base64-encoded URL of the custom fallback image.

Default: blank

### Skip processing

```
skip_processing:%extension1:%extension2:...:%extensionN
skp:%extension1:%extension2:...:%extensionN
```

When set, imgproxy will skip the processing of the listed formats. Also available as the [IMGPROXY_SKIP_PROCESSING_FORMATS](configuration.md#skip-processing) configuration.

**📝Note:** Processing can only be skipped when the requested format is the same as the source format.

**📝Note:** Video thumbnail processing can't be skipped.

Default: empty

### Raw

```
raw:%raw
```

When set to `1`, `t` or `true`, imgproxy will respond with a raw unprocessed, and unchecked source image. There are some differences between `raw` and `skip_processing` options:

* While the `skip_processing` option has some conditions to skip the processing, the `raw` option allows to skip processing no matter what
* With the `raw` option set, imgproxy doesn't check the source image's type, resolution, and file size. Basically, the `raw` option allows streaming of any file type
* With the `raw` option set, imgproxy won't download the whole image to the memory. Instead, it will stream the source image directly to the response lowering memory usage
* The requests with the `raw` option set are not limited by the `IMGPROXY_CONCURRENCY` config

Default: `false`

### Cache buster

```
cachebuster:%string
cb:%string
```

Cache buster doesn't affect image processing but its changing allows for bypassing the CDN, proxy server and browser cache. Useful when you have changed some things that are not reflected in the URL, like image quality settings, presets, or watermark data.

It's highly recommended to prefer the `cachebuster` option over a URL query string because that option can be properly signed.

Default: empty

### Expires

```
expires:%timestamp
exp:%timestamp
```

When set, imgproxy will check the provided unix timestamp and return 404 when expired.

Default: empty

### Filename

```
filename:%string
fn:%string
```

Defines a filename for the `Content-Disposition` header. When not specified, imgproxy will get the filename from the source url.

Default: empty

### Return attachment

```
return_attachment:%return_attachment
att:%return_attachment
```

When set to `1`, `t` or `true`, imgproxy will return `attachment` in the `Content-Disposition` header, and the browser will open a 'Save as' dialog. This is normally controlled by the [IMGPROXY_RETURN_ATTACHMENT](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

### Preset

```
preset:%preset_name1:%preset_name2:...:%preset_nameN
pr:%preset_name1:%preset_name2:...:%preset_nameN
```

Defines a list of presets to be used by imgproxy. Feel free to use as many presets in a single URL as you need.

Read more about presets in the [Presets](presets.md) guide.

Default: empty

## Source URL
### Plain

The source URL can be provided as is, prepended by the `/plain/` segment:

```
/plain/http://example.com/images/curiosity.jpg
```

**📝Note:** If the source URL contains a query string or `@`, you'll need to escape it.

When using a plain source URL, you can specify the [extension](#extension) after `@`:

```
/plain/http://example.com/images/curiosity.jpg@png
```

### Base64 encoded

The source URL can be encoded with URL-safe Base64. The encoded URL can be split with `/` as desired:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn
```

When using an encoded source URL, you can specify the [extension](#extension) after `.`:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

### Encrypted with AES-CBC

The source URL can be encrypted with the AES-CBC algorithm, prepended by the `/enc/` segment. The encrypted URL can be split with `/` as desired:

```
/enc/jwV3wUD9r4VBIzgv/ang3Hbh0vPpcm5cc/VO5rHxzonpvZjppG/2VhDnX2aariBYegH/jlhw_-dqjXDMm4af/ZDU6y5sBog
```

When using an encrypted source URL, you can specify the [extension](#extension) after `.`:

```
/enc/jwV3wUD9r4VBIzgv/ang3Hbh0vPpcm5cc/VO5rHxzonpvZjppG/2VhDnX2aariBYegH/jlhw_-dqjXDMm4af/ZDU6y5sBog.png
```

## Extension

Extension specifies the format of the resulting image. Read more about image formats support [here](image_formats_support.md).

The extension can be omitted. In this case, imgproxy will use the source image format as resulting one. If the source image format is not supported as the resulting image, imgproxy will use `jpg`. You also can [enable WebP support detection](configuration.md#avifwebp-support-detection) to use it as the default resulting format when possible.

### Best format![pro](./assets/pro.svg)

You can use the `best` value for the [format](generating_the_url#format) option or the [extension](generating_the_url#extension) to make imgproxy pick the best format for the resultant image. Check out the [Best format](best_format) guide to learn more.

## Example

A signed imgproxy URL that uses the `sharp` preset, resizes `http://example.com/images/curiosity.jpg` to fill a `300x400` area using smart gravity without enlarging, and then converts the image to `png`:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/preset:sharp/resize:fill:300:400:0/gravity:sm/plain/http://example.com/images/curiosity.jpg@png
```

The same URL with shortcuts will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/pr:sharp/rs:fill:300:400:0/g:sm/plain/http://example.com/images/curiosity.jpg@png
```

The same URL with a Base64-encoded source URL will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/pr:sharp/rs:fill:300:400:0/g:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```
