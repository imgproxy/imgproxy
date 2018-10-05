# Generating the URL (Advanced)

This guide describes the advanced URL format that supports all the imgproxy features. Read our [Generating the URL (Basic)](./generating_the_url_basic.md) guide to get info about basic URL format that is compatible with the first version of imgproxy.

### Format definition

The advanced URL should contain the signature, processing options, and encoded source URL, like this:

```
/%signature/%processing_options/%encoded_url.%extension
```

Check out the [example](#example) at the end of this guide.

#### Signature

Signature protects your URL from being changed by an attacker. It's highly recommended to sign imgproxy URLs in production.

If you set up [URL signature](./configuration.md#url-signature), check out [Signing the URL](./signing_the_url.md) guide to know how to sign your URLs. Otherwise, use any string here.

#### Processing options

Processing options should be specified as URL parts divided by slashes (`/`). Processing option has the following format:

```
%option_name:%argument1:%argument2:...:argumentN
```

Processing options should not be treated as a processing pipeline. Processing pipeline of imgproxy is fixed to provide you a maximum performance. Read more about it in [About processing pipeline](./about_processing_pipeline.md) guide.

imgproxy supports the following processing options:

##### Resize

`resize:%resizing_type:%width:%height:%enlarge`
`rs:%resizing_type:%width:%height:%enlarge`

Meta-option that defines [resizing type](#resizing-type), [width](#width), [height](#height), and [enlarge](#enlarge). All arguments are optional and can be omited to use their default values.

##### Size

`size:%width:%height:%enlarge`
`s:%width:%height:%enlarge`

Meta-option that defines [width](#width), [height](#height), and [enlarge](#enlarge). All arguments are optional and can be omited to use their default values.

##### Resizing type

`resizing_type:%resizing_type`
`rt:%resizing_type`

Defines how imgproxy will resize the source image. Supported resizing types are:

* `fit` — resizes the image while keeping aspect ratio to fit given size;
* `fill` — resizes the image while keeping aspect ratio to fill given size and cropping projecting parts;
* `crop` — crops the image to a given size.

Default: `fit`

##### Width

`width:%width`
`w:%width`

Defines the width of the resulting image. When set to `0`, imgproxy will calculate resulting width by defined height and source aspect ratio. When set to `0` and `crop` resizing type is used, imgproxy will use the full width of the source image.

Default: `0`

##### Height

`height:%height`
`h:%height`

Defines the height of the resulting image. When set to `0`, imgproxy will calculate resulting height by defined width and source aspect ratio. When set to `0` and `crop` resizing type is used, imgproxy will use the full height of the source image.

Default: `0`

##### Enlarge

`enlarge:%enlarge`
`el:%enlarge`

If set to `0`, imgproxy will not enlarge the image if it is smaller than the given size. With any other value, imgproxy will enlarge the image.

Default: `0`

##### Gravity

`gravity:%gravity`
`g:%gravity`

When imgproxy needs to cut some parts of the image, it is guided by the gravity. The following values are supported:

* `no` — north (top edge);
* `so` — south (bottom edge);
* `ea` — east (right edge);
* `we` — west (left edge);
* `ce` — center;
* `sm` — smart. `libvips` detects the most "interesting" section of the image and considers it as the center of the resulting image;
* `fp:%x:%y` - focus point. `x` and `y` are floating point numbers between 0 and 1 that defines coordinates of the center of the resulting image. Trait 0 and 1 as right/left for `x` and top/bottom for `y`.

Default: `ce`

##### Background

`background:%R:%G:%B`
`bg:%R:%G:%B`

`background:%hex_color`
`bg:%hex_color`

When set, imgproxy will fill the resulting image background with specified color. `R`, `G`, and `B` are red, green and blue channel values of background color (0-255). `hex_color` is hex-coded color. Useful when you convert an image with alpha-channel to JPEG.

When no arguments is provided, disables background.

Default: disabled

##### Blur

`blur:%sigma`
`bl:%sigma`

When set, imgproxy will apply gaussian blur filter to the resulting image. `sigma` defines how large a mask imgproxy will use.

Default: disabled

##### Sharpen

`sharpen:%sigma`
`sh:%sigma`

When set, imgproxy will apply sharpen filter to the resulting image. `sigma` defines how large a mask imgproxy will use.

As an approximate guideline, use 0.5 sigma for 4 pixels/mm (display resolution), 1.0 for 12 pixels/mm and 1.5 for 16 pixels/mm (300 dpi == 12 pixels/mm).

Default: disabled

##### Preset

`preset:%preset_name1:%preset_name2:...:%preset_nameN`
`pr:%preset_name1:%preset_name2:...:%preset_nameN`

Defines presets to be used by imgproxy. Feel free to use as many presets in a single URL as you need.

Read more about presets in our [Presets](./presets.md) guide.

Default: empty

##### Format

`format:%extension`
`f:%extension`
`ext:%extension`

Specifies resulting image format. Alias for [extension](#extension) URL part.

Default: `jpg`

#### Encoded URL

The source URL should be encoded with URL-safe Base64. The encoded URL can be split with `/` for your needs.

#### Extension

Extension specifies the format of the resulting image. At the moment, imgproxy supports only `jpg`, `png` and `webp`, them being the most popular and useful web image formats.

The extension part can be omitted. In this case, if the format is not defined by processing options, imgproxy will use `jpg` by default. You also can [enable WebP support detection](./configuration.md#webp-support-detection) to use it as default resulting format when possible.

### Example

Signed imgproxy URL that uses `sharp` preset, resizes `http://example.com/images/curiosity.jpg` to fill `300x400` area with smart gravity without enlarging, and converts the image to `png` will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/preset:sharp/resize:fill:300:400:0/gravity:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

The same URL with shortcuts will look like this:


```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/pr:sharp/rs:fill:300:400:0/g:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```
