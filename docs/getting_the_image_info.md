# Getting the image info![pro](./assets/pro.svg)

imgproxy can fetch and return a source image info without downloading the whole image.

## URL format

To get the image info, use the following URL format:

```
/info/%signature/%info_options/plain/%source_url
/info/%signature/%info_options/%encoded_source_url
```

## Signature

A signature protects your URL from being modified by an attacker. It is highly recommended to sign imgproxy URLs in a production environment.

Once you set up your [URL signature](configuration.md#url-signature), check out the [Signing the URL](signing_the_url.md) guide to learn about how to sign your URLs. Otherwise, since the signature is required, feel free to use any string here.

## Info options

Info options should be specified as URL parts divided by slashes (`/`). An info option has the following format:

```
%option_name:%argument1:%argument2:...:argumentN
```

### Size

```
size:%size
s:%size
```

When set to `1`, `t`, or `true`, imgproxy will return the size of the image file. If the source URL is an HTTP(s) URL, imgproxy will determine the file size based on the `Content-Length` HTTP header.

Default: `true`.

**Response example:**

```json
{
  "size": 123456
}
```

### Format

```
format:%format
f:%format
```

When set to `1`, `t`, or `true`, imgproxy will return the image format.

**📝 Note:** For video files, imgproxy returns a list of predicted formats divided by comma.

Default: `true`.

**Response example:**

```json
{
  "format": "jpeg"
}
```

### Dimensions

```
dimensions:%dimensions
d:%dimensions
```

When set to `1`, `t`, or `true`, imgproxy will return the image dimensions.

Default: `true`.

**Response example:**

```json
{
  "width": 7360,
  "height": 4912
}
```

### EXIF

```
exif:%exif
```

When set to `1`, `t`, or `true`, imgproxy will return the image's EXIF metadata.

Default: `true`.

**Response example:**

```json
{
  "exif": {
    "Aperture": "8.00 EV (f/16.0)",
    "Contrast": "Normal",
    "Date and Time": "2016:09:11 22:15:03",
    "Model": "NIKON D810",
    "Software": "Adobe Photoshop Lightroom 6.1 (Windows)"
  }
}
```

### IPTC

```
iptc:%iptc
```

When set to `1`, `t`, or `true`, imgproxy will return the image's IPTC (IPTC-IIM) metadata and Photoshop metadata (currently, only the resolution data).

Default: `true`.

**Response example:**

```json
{
  "iptc": {
    "Name": "Spider-Man",
    "Caption": "Spider-Man swings on the web",
    "Copyright Notice": "Daily Bugle",
    "Keywords": ["spider-man", "menance", "offender"]
  },
  "photoshop": {
    "resolution": {
      "XResolution": 240,
      "XResolutionUnit": "inches",
      "WidthUnit": "inches",
      "YResolution": 240,
      "YResolutionUnit": "inches",
      "HeightUnit": "inches"
    }
  }
}
```

### XMP

```
xmp:%xmp
```

When set to `1`, `t`, or `true`, imgproxy will return the image's XMP metadata.

Default: `true`.

**Response example:**

```json
{
  "xmp": {
    "aux": {
      "ApproximateFocusDistance": "4294967295/1",
      "ImageNumber": "16604",
      "Lens": "16.0-35.0 mm f/4.0",
      "LensID": "163",
      "LensInfo": "160/10 350/10 40/10 40/10",
      "SerialNumber": "12345678"
    },
    "dc": {
      "creator": ["Peter B. Parker"],
      "publisher": ["Daily Bugle"],
      "subject": ["spider-man", "menance", "offender"],
      "format": "image/jpeg"
    },
    "photoshop": {
      "DateCreated": "2016-09-11T18:44:50.003"
    }
  }
}
```

### Video meta

```
video_meta:%video_meta
vm:%video_meta
```

When set to `1`, `t`, or `true`, imgproxy will return the video metadata and video streams info.

Default: `true`.

**Response example:**

```json
{
  "video_meta": {
    "com.android.version": "9",
    "compatible_brands": "isommp42",
    "creation_time": "2022-01-12T15:04:10.000000Z",
    "location": "+46.4845+030.6848/",
    "location-eng": "+46.4845+030.6848/",
    "major_brand": "mp42",
    "minor_version": "0"
  },
  "video_streams": [
    {
      "type": "video",
      "codec": "h264",
      "bps": 16910024,
      "fps": 24,
      "language": "eng"
    },
    {
      "type": "audio",
      "codec": "eac3",
      "bps": 768000,
      "frequency": 48000,
      "layout": "5.1(side)",
      "language": "eng"
    },
    {
      "type": "subtitle",
      "codec": "subrip",
      "language": "eng"
    }
  ]
}
```

### Detect objects

**⏳ Slow:** This option requires the image to be fully downloaded and processed.

```
detect_objects:%detect_objects
do:%detect_objects
```

When set to `1`, `t`, or `true`, imgproxy will return the info about the objects found in the image. Read the [object detection](object_detection.md) manual to learn how to configure object detection.

**📝 Note:** imgproxy returns the relative coordinates of the found objects.

Default: `false`.

**Response example:**

```json
{
  "objects": [
    {
      "class_id": 0,
      "class_name": "face",
      "confidence": 0.985792,
      "left": 0.6602726057171822,
      "top": 0.23434072732925415,
      "width": 0.11385439336299896,
      "height": 0.18671900033950806
    },
    {
      "class_id": 0,
      "class_name": "face",
      "confidence": 0.9810329,
      "left": 0.4354642778635025,
      "top": 0.3503067269921303,
      "width": 0.10691609978675842,
      "height": 0.18357203900814056
    }
  ]
}
```

### Crop coordinates :id=crop

**⏳ Slow:** This option requires the image to be fully downloaded and processed.

```
crop:%width:%height:%gravity
c:%width:%height:%gravity
```

When `width` and `height` are greater than zero, imgproxy will return the _relative_ crop coordinates for the defined crop parameters.

This option takes the same arguments as the [crop](generating_the_url.md#crop) processiong option.

Default: `0:0:ce`.

**Response example:**

```json
{
  "crop": {
    "left": 0.383203125,
    "top": 0.2603861907548274,
    "width": 0.1953125,
    "height": 0.3510825043885313
  }
}
```

### Palette

**⏳ Slow:** This option requires the image to be fully downloaded and processed.

```
palette:%colors
p:%colors
```

When `colors` is greater than zero, imgproxy will build and return the image's RGBA palette containing maximum `colors` colors.

**📝 Note:** When `colors` is greater than zero, its value should be between `2` and `256`.

Default: `0`.

**Response example:**

```json
{
  "palette": [
    { "R": 189, "G": 178, "B": 169, "A": 255 },
    { "R": 83, "G": 79, "B": 67, "A": 255 }
  ]
}
```

### Average color :id=average

**⏳ Slow:** This option requires the image to be fully downloaded and processed.

```
average:%average:%ignore_transparent
avg:%average:%ignore_transparent
```

* `average` – when set to `1`, `t`, or `true`, imgproxy will calculate and return the image's average color. Default: `false`
* `ignore_transparent` – _(optional)_ when set to `1`, `t`, or `true`, imgproxy will ignore fully transparent pixels. Default: `true`

Default: `false:true`

**Response example:**

```json
{
  "average": { "R": 139, "G": 132, "B": 121, "A": 255 }
}
```

### Dominant colors

**⏳ Slow:** This option requires the image to be fully downloaded and processed.

```
dominant_colors:%dominant_colors:%build_missed
dc:%dominant_colors:%build_missed
```

* `dominant_colors` – when set to `1`, `t`, or `true`, imgproxy will calculate and return the image's dominant colors (vibrant, light vibrant, dark vibrant, muted, light muted, and dark muted). Default: `false`
* `build_missed` – _(optional)_ when set to `1`, `t`, or `true`, imgproxy will build colors that were not found in the image based on the found ones. Default: `false`

Default: `false:false`

**Response example:**

```json
{
  "dominant_colors": {
    "dark_muted": { "R": 75, "G": 70, "B": 57 },
    "dark_vibrant": { "R": 90, "G": 78, "B": 43 },
    "light_muted": { "R": 167, "G": 156, "B": 130 },
    "light_vibrant": { "R": 212, "G": 198, "B": 165 },
    "muted": { "R": 155, "G": 146, "B": 120 },
    "vibrant": { "R": 172, "G": 146, "B": 83 }
  }
}
```

### BlurHash

**⏳ Slow:** This option requires the image to be fully downloaded and processed.

```
blurhash:%x_components:%y_components
bh:%x_components:%y_components
```

When `x_components` and `y_components` are greater than zero, imgproxy will calculate and return the image's [BlurHash](https://blurha.sh/). `x_components` and `y_components` is the numbers of horizontal and vertical components of BlurHash. The larger the numbers the more "detailed" will be the BlurHash.

The maximum value for `x_components` and `y_components` is `9`.

Default: `0:0`

**Response example:**

```json
{
  "blurhash": "LLH-}fox0fRQ%Do}9as9_3%2M{S2"
}
```

### Page

```
page:%page
pg:%page
```

When a source image supports pagination (PDF, TIFF) or animation (GIF, WebP), this option allows specifying the page to use. Page numeration starts from zero.

Default: 0

### Video thumbnail second

```
video_thumbnail_second:%second
vts:%second
```

Allows redefining `IMGPROXY_VIDEO_THUMBNAIL_SECOND` config.

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

### Preset

```
preset:%preset_name1:%preset_name2:...:%preset_nameN
pr:%preset_name1:%preset_name2:...:%preset_nameN
```

Defines a list of presets to be used by imgproxy. Feel free to use as many presets in a single URL as you need.

Read more about presets in the [Presets](presets.md) guide.

Default: empty

### Max src resolution

```
max_src_resolution:%resolution
msr:%resolution
```

Allows redefining `IMGPROXY_MAX_SRC_RESOLUTION` config.

**⚠️ Warning:** Since this option allows redefining a security restriction, its usage is not allowed unless the `IMGPROXY_ALLOW_SECURITY_OPTIONS` config is set to `true`.

### Max src file size

```
max_src_file_size:%size
msfs:%size
```

Allows redefining `IMGPROXY_MAX_SRC_FILE_SIZE` config.

**⚠️ Warning:** Since this option allows redefining a security restriction, its usage is not allowed unless the `IMGPROXY_ALLOW_SECURITY_OPTIONS` config is set to `true`.

## Source URL
### Plain

The source URL can be provided as is, prepended by `/plain/` part:

```
/plain/http://example.com/images/curiosity.jpg
```

**📝 Note:** If the source URL contains a query string or `@`, you'll need to escape it.

### Base64 encoded

The source URL can be encoded with URL-safe Base64. The encoded URL can be split with `/` as desired:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn
```

### Encrypted with AES-CBC![pro](./assets/pro.svg) :id=encrypted-with-aes-cbc

The source URL can be encrypted with the AES-CBC algorithm, prepended by the `/enc/` segment. The encrypted URL can be split with `/` as desired:

```
/enc/jwV3wUD9r4VBIzgv/ang3Hbh0vPpcm5cc/VO5rHxzonpvZjppG/2VhDnX2aariBYegH/jlhw_-dqjXDMm4af/ZDU6y5sBog
```
