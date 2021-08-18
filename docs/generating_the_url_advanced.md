# Generating the URL (Advanced)

This guide describes the advanced URL format that allows the use of all the imgproxy features. Read our [Generating the URL (Basic)](generating_the_url_basic.md) guide to learn about the _basic_ URL format that is compatible with imgproxy 1.x.

## Format definition

The advanced URL should contain the signature, processing options, and source URL, like this:

```
/%signature/%processing_options/plain/%source_url@%extension
/%signature/%processing_options/%encoded_source_url.%extension
```

Check out the [example](#example) at the end of this guide.

### Signature

Signature protects your URL from being altered by an attacker. It is highly recommended to sign imgproxy URLs in production.

Once you set up your [URL signature](configuration.md#url-signature), check out the [Signing the URL](signing_the_url.md) guide to know how to sign your URLs. Otherwise, use any string here.

### Processing options

Processing options should be specified as URL parts divided by slashes (`/`). Processing option has the following format:

```
%option_name:%argument1:%argument2:...:argumentN
```

The list of processing options does not define imgproxy's processing pipeline. Instead, imgproxy already comes with a specific, built-in image processing pipeline for the maximum performance. Read more about it in the [About processing pipeline](about_processing_pipeline.md) guide.

imgproxy supports the following processing options:

#### Resize

```
resize:%resizing_type:%width:%height:%enlarge:%extend
rs:%resizing_type:%width:%height:%enlarge:%extend
```

Meta-option that defines the [resizing type](#resizing-type), [width](#width), [height](#height), [enlarge](#enlarge), and [extend](#extend). All arguments are optional and can be omitted to use their default values.

#### Size

```
size:%width:%height:%enlarge:%extend
s:%width:%height:%enlarge:%extend
```

Meta-option that defines the [width](#width), [height](#height), [enlarge](#enlarge), and [extend](#extend). All arguments are optional and can be omitted to use their default values.

#### Resizing type

```
resizing_type:%resizing_type
rt:%resizing_type
```

Defines how imgproxy will resize the source image. Supported resizing types are:

* `fit`: resizes the image while keeping aspect ratio to fit given size;
* `fill`: resizes the image while keeping aspect ratio to fill given size and cropping projecting parts;
* `auto`: if both source and resulting dimensions have the same orientation (portrait or landscape), imgproxy will use `fill`. Otherwise, it will use `fit`.

Default: `fit`

#### Resizing algorithm<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=resizing-algorithm

```
resizing_algorithm:%algorithm
ra:%algorithm
```

Defines the algorithm that imgproxy will use for resizing. Supported algorithms are `nearest`, `linear`, `cubic`, `lanczos2`, and `lanczos3`.

Default: `lanczos3`

#### Width

```
width:%width
w:%width
```

Defines the width of the resulting image. When set to `0`, imgproxy will calculate the resulting width using the defined height and source aspect ratio.

Default: `0`

#### Height

```
height:%height
h:%height
```

Defines the height of the resulting image. When set to `0`, imgproxy will calculate resulting height using the defined width and source aspect ratio.

Default: `0`

#### Dpr

```
dpr:%dpr
```

When set, imgproxy will multiply the image dimensions according to this factor for HiDPI (Retina) devices. The value must be greater than 0.

Default: `1`

#### Enlarge

```
enlarge:%enlarge
el:%enlarge
```

When set to `1`, `t` or `true`, imgproxy will enlarge the image if it is smaller than the given size.

Default: false

#### Extend

```
extend:%extend:%gravity
ex:%extend:%gravity
```

* When `extend` is set to `1`, `t` or `true`, imgproxy will extend the image if it is smaller than the given size.
* `gravity` _(optional)_ accepts the same values as [gravity](#gravity) option, except `sm`. When `gravity` is not set, imgproxy will use `ce` gravity without offsets.

Default: `false:ce:0:0`

#### Gravity

```
gravity:%type:%x_offset:%y_offset
g:%type:%x_offset:%y_offset
```

When imgproxy needs to cut some parts of the image, it is guided by the gravity.

* `type` - specifies the gravity type. Available values:
  * `no`: north (top edge);
  * `so`: south (bottom edge);
  * `ea`: east (right edge);
  * `we`: west (left edge);
  * `noea`: north-east (top-right corner);
  * `nowe`: north-west (top-left corner);
  * `soea`: south-east (bottom-right corner);
  * `sowe`: south-west (bottom-left corner);
  * `ce`: center.
* `x_offset`, `y_offset` - (optional) specify gravity offset by X and Y axes.

Default: `ce:0:0`

**Special gravities**:

* `gravity:sm` - smart gravity. `libvips` detects the most "interesting" section of the image and considers it as the center of the resulting image. Offsets are not applicable here;
* `gravity:fp:%x:%y` - focus point gravity. `x` and `y` are floating point numbers between 0 and 1 that define the coordinates of the center of the resulting image. Treat 0 and 1 as right/left for `x` and top/bottom for `y`.

#### Crop

```
crop:%width:%height:%gravity
c:%width:%height:%gravity
```

Defines an area of the image to be processed (crop before resize).

* `width` and `height` define the size of the area:
  * When `width` or `height` is greater than or equal to `1`, imgproxy treats it as an absolute value.
  * When `width` or `height` is less than `1`, imgproxy treats it as a relative value.
  * When `width` or `height` is set to `0`, imgproxy will use the full width/height of the source image.
* `gravity` _(optional)_ accepts the same values as [gravity](#gravity) option. When `gravity` is not set, imgproxy will use the value of the [gravity](#gravity) option.

#### Padding

```
padding:%top:%right:%bottom:%left
pd:%top:%right:%bottom:%left
```

Defines padding size in css manner. All arguments are optional but at least one dimension must be set. Padded space is filled according to [background](#background) option.

* `top` - top padding (and all other sides if they won't be set explicitly);
* `right` - right padding (and left if it won't be set explicitly);
* `bottom` - bottom padding;
* `left` - left padding.

**📝Note:** Padding is applied after all image transformations (except watermark) and enlarges generated image which means that if your resize dimensions were 100x200px and you applied `padding:10` option then you will get 120x220px image.

**📝Note:** Padding follows [dpr](#dpr) option so it will be scaled too if you set it.

#### Trim

```
trim:%threshold:%color:%equal_hor:%equal_ver
t:%threshold:%color:%equal_hor:%equal_ver
```

Removes surrounding background.

* `threshold` - color similarity tolerance.
* `color` - _(optional)_ hex-coded value of the color that needs to be cut off.
* `equal_hor` - _(optional)_ set to `1`, `t` or `true`, imgproxy will cut only equal parts from left and right sides. That means that if 10px of background can be cut off from left and 5px from right then 5px will be cut off from both sides. For example, it can be useful if objects on your images are centered but have non-symmetrical shadow.
* `equal_ver` - _(optional)_ acts like `equal_hor` but for top/bottom sides.

**⚠️Warning:** Trimming requires an image to be fully loaded into memory. This disables scale-on-load and significantly increases memory usage and processing time. Use it carefully with large images.

**📝Note:** If you know background color of your images then setting it explicitly via `color` will also save some resources because imgproxy won't detect it automatically.

**📝Note:** Use `color` value of `FF00FF` for trimming transparent background as imgproxy uses magenta as a transparency key.

**📝Note:** Trimming of animated images is not supported.

#### Rotate

```
rotate:%angle
rot:%angle
```

Rotates the image on the specified angle. The orientation from the image metadata is applied before the rotation unless autorotation is disabled.

**📝Note:** Only 0/90/180/270/etc degrees angles are supported.

Default: 0.

#### Quality

```
quality:%quality
q:%quality
```

Redefines quality of the resulting image, percentage. When `0`, quality is assumed based on `IMGPROXY_QUALITY` and `IMGPROXY_FORMAT_QUALITY`.

Default: 0.

#### Max Bytes

```
max_bytes:%bytes
mb:%bytes
```

When set, imgproxy automatically degrades the quality of the image until the image is under the specified amount of bytes.

**📝Note:** Applicable only to `jpg`, `webp`, `heic`, and `tiff`.

**⚠️Warning:** When `max_bytes` is set, imgproxy saves image multiple times to achieve specified image size.

Default: 0

#### Background

```
background:%R:%G:%B
bg:%R:%G:%B

background:%hex_color
bg:%hex_color
```

When set, imgproxy will fill the resulting image background with the specified color. `R`, `G`, and `B` are red, green and blue channel values of the background color (0-255). `hex_color` is a hex-coded value of the color. Useful when you convert an image with alpha-channel to JPEG.

With no arguments provided, disables any background manipulations.

Default: disabled

#### Background alpha<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=background-alpha

```
background_alpha:%alpha
bga:%alpha
```

Adds alpha channel to `background`. `alpha` is a positive floating point number between `0` and `1`.

Default: 1

#### Adjust<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=adjust

```
adjust:%brightness:%contrast:%saturation
a:%brightness:%contrast:%saturation
```

Meta-option that defines the [brightness](#brightness), [contrast](#contrast), and [saturation](#saturation). All arguments are optional and can be omitted to use their default values.

#### Brightness<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=brightness

```
brightness:%brightness
br:%brightness
```

When set, imgproxy will adjust brightness of the resulting image. `brightness` is an integer number in range from `-255` to `255`.

Default: 0

#### Contrast<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=contrast

```
contrast:%contrast
co:%contrast
```

When set, imgproxy will adjust contrast of the resulting image. `contrast` is a positive floating point number, where `1` keeps the contrast unchanged.

Default: 1

#### Saturation<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=saturation

```
saturation:%saturation
sa:%saturation
```

When set, imgproxy will adjust saturation of the resulting image. `saturation` is a positive floating point number, where `1` keeps the saturation unchanged.

Default: 1

#### Blur

```
blur:%sigma
bl:%sigma
```

When set, imgproxy will apply the gaussian blur filter to the resulting image. `sigma` defines the size of a mask imgproxy will use.

Default: disabled

#### Sharpen

```
sharpen:%sigma
sh:%sigma
```

When set, imgproxy will apply the sharpen filter to the resulting image. `sigma` the size of a mask imgproxy will use.

As an approximate guideline, use 0.5 sigma for 4 pixels/mm (display resolution), 1.0 for 12 pixels/mm and 1.5 for 16 pixels/mm (300 dpi == 12 pixels/mm).

Default: disabled

#### Pixelate<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=pixelate

```
pixelate:%size
pix:%size
```

When set, imgproxy will apply the pixelate filter to the resulting image. `size` is the size of a pixel.

Default: disabled

#### Unsharpening<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=unsharpening

```
unsharpening:%mode:%weight:%dividor
ush:%mode:%weight:%dividor
```

Allows redefining unsharpening options. All arguments have the same meaning as [Unsharpening](configuration.md#unsharpening) configs. All arguments are optional and can be omitted.

#### Watermark

```
watermark:%opacity:%position:%x_offset:%y_offset:%scale
wm:%opacity:%position:%x_offset:%y_offset:%scale
```

Puts watermark on the processed image.

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

Default: disabled

#### Watermark URL<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=watermark

```
watermark_url:%url
wmu:%url
```

When set, imgproxy will use the image from the specified URL as a watermark. `url` is Base64-encoded URL of the custom watermark.

Default: blank

#### Style<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=style

```
style:%style
st:%style
```

When set, imgproxy will prepend `<style>` node with provided content to the `<svg>` node of source SVG image. `%style` is url-safe Base64-encoded CSS-style.

Default: blank

#### JPEG options<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=jpeg-options

```
jpeg_options:%progressive:%no_subsample:%trellis_quant:%overshoot_deringing:%optimize_scans:%quant_table
jpgo:%progressive:%no_subsample:%trellis_quant:%overshoot_deringing:%optimize_scans:%quant_table
```

Allows redefining JPEG saving options. All arguments have the same meaning as [Advanced JPEG compression](configuration.md#advanced-jpeg-compression) configs. All arguments are optional and can be omitted.

#### PNG options<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=png-options

```
png_options:%interlaced:%quantize:%quantization_colors
pngo:%interlaced:%quantize:%quantization_colors
```

Allows redefining PNG saving options. All arguments have the same meaning as [Advanced PNG compression](configuration.md#advanced-png-compression) configs. All arguments are optional and can be omitted.

#### GIF options<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=gif-options

```
gif_options:%optimize_frames:%optimize_transparency
gifo:%optimize_frames:%optimize_transparency
```

Allows redefining GIF saving options. All arguments have the same meaning as [Advanced GIF compression](configuration.md#advanced-gif-compression) configs. All arguments are optional and can be omitted.

#### Page<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=page

```
page:%page
pg:%page
```

When source image supports pagination (PDF, TIFF) or animation (GIF, WebP), this option allows specifying the page to use. Pages numeration starts from zero.

Default: 0

#### Video thumbnail second<img class='pro-badge' src='assets/pro.svg' alt='pro' /> :id=video-thumbnail-second

```
video_thumbnail_second:%second
vts:%second
```

Allows redefining `IMGPROXY_VIDEO_THUMBNAIL_SECOND` config.

#### Preset

```
preset:%preset_name1:%preset_name2:...:%preset_nameN
pr:%preset_name1:%preset_name2:...:%preset_nameN
```

Defines a list of presets to be used by imgproxy. Feel free to use as many presets in a single URL as you need.

Read more about presets in the [Presets](presets.md) guide.

Default: empty

#### Cache buster

```
cachebuster:%string
cb:%string
```

Cache buster doesn't affect image processing but it's changing allows to bypass CDN, proxy server and browser cache. Useful when you have changed some things that are not reflected in the URL like image quality settings, presets or watermark data.

It's highly recommended to prefer `cachebuster` option over URL query string because the option can be properly signed.

Default: empty

#### Strip Metadata

```
strip_metadata:%strip_metadata
sm:%strip_metadata
```

When set to `1`, `t` or `true`, imgproxy will strip the metadata (EXIF, IPTC, etc.) on JPEG and WebP output images. Normally this is controlled by the [IMGPROXY_STRIP_METADATA](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

#### Strip Color Profile

```
strip_color_profile:%strip_color_profile
scp:%strip_color_profile
```

When set to `1`, `t` or `true`, imgproxy will transform the embedded color profile (ICC) to sRGB and remove it from the image. Otherwise, imgproxy will try to keep it as is. Normally this is controlled by the [IMGPROXY_STRIP_COLOR_PROFILE](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

#### Auto Rotate

```
auto_rotate:%auto_rotate
ar:%auto_rotate
```

When set to `1`, `t` or `true`, imgproxy will automatically rotate images based onon the EXIF Orientation parameter (if available in the image meta data). The orientation tag will be removed from the image anyway. Normally this is controlled by the [IMGPROXY_AUTO_ROTATE](configuration.md#miscellaneous) configuration but this procesing option allows the configuration to be set for each request.

#### Filename

```
filename:%string
fn:%string
```

Defines a filename for `Content-Disposition` header. When not specified, imgproxy will get filename from the source url.

Default: empty

#### Format

```
format:%extension
f:%extension
ext:%extension
```

Specifies the resulting image format. Alias for [extension](#extension) URL part.

Default: `jpg`

### Source URL

There are two ways to specify source url:

#### Plain

The source URL can be provided as is, prendended by `/plain/` part:

```
/plain/http://example.com/images/curiosity.jpg
```

**📝Note:** If the source URL contains query string or `@`, you need to escape it.

When using plain source URL, you can specify the [extension](#extension) after `@`:

```
/plain/http://example.com/images/curiosity.jpg@png
```

#### Base64 encoded

The source URL can be encoded with URL-safe Base64. The encoded URL can be split with `/` for your needs:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn
```

When using encoded source URL, you can specify the [extension](#extension) after `.`:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

### Extension

Extension specifies the format of the resulting image. Read about image formats support [here](image_formats_support.md).

The extension part can be omitted. In this case, imgproxy will use source image format as resulting one. If source image format is not supported as resulting, imgproxy will use `jpg`. You also can [enable WebP support detection](configuration.md#webp-support-detection) to use it as default resulting format when possible.

## Example

Signed imgproxy URL that uses `sharp` preset, resizes `http://example.com/images/curiosity.jpg` to fill `300x400` area with smart gravity without enlarging, and then converts the image to `png`:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/preset:sharp/resize:fill:300:400:0/gravity:sm/plain/http://example.com/images/curiosity.jpg@png
```

The same URL with shortcuts will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/pr:sharp/rs:fill:300:400:0/g:sm/plain/http://example.com/images/curiosity.jpg@png
```

The same URL with Base64-encoded source URL will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/pr:sharp/rs:fill:300:400:0/g:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```
