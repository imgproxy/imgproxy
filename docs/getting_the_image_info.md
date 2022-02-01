# Getting the image info<i class='badge badge-pro'></i>

imgproxy can fetch and return a source image info without downloading the whole image.

## URL format

To get the image info, use the following URL format:

```
/info/%signature/plain/%source_url
/info/%signature/%encoded_source_url
```

### Signature

A signature protects your URL from being modified by an attacker. It is highly recommended to sign imgproxy URLs in a production environment.

Once you set up your [URL signature](configuration.md#url-signature), check out the [Signing the URL](signing_the_url.md) guide to learn about how to sign your URLs. Otherwise, since the signature is required, feel free to use any string here.

### Source URL

There are two ways to specify source url:

#### Plain

The source URL can be provided as is, prepended by `/plain/` part:

```
/plain/http://example.com/images/curiosity.jpg
```

**📝Note:** If the source URL contains a query string or `@`, you'll need to escape it.

#### Base64 encoded

The source URL can be encoded with URL-safe Base64. The encoded URL can be split with `/` as desired:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn
```

## Response format

imgproxy responses with a JSON body and returns the following info:

* `format`: source image/video format. In case of video - list of predicted formats divided by comma
* `width`: image/video width
* `height`: image/video height
* `size`: file size. Can be zero if the image source doesn't set `Content-Length` header properly
* `exif`: Exif data
* `iptc`: IPTC data
* `video_meta`: metadata from the video

**📝Note:** There are lots of IPTC tags in the spec, but imgproxy supports only a few of them. If you need some tags to be supported, just contact us.

#### Example (JPEG)

```json
{
  "format": "jpeg",
  "width": 7360,
  "height": 4912,
  "size": 28993664,
  "exif": {
    "Aperture": "8.00 EV (f/16.0)",
    "Contrast": "Normal",
    "Date and Time": "2016:09:11 22:15:03",
    "Model": "NIKON D810",
    "Software": "Adobe Photoshop Lightroom 6.1 (Windows)"
  },
  "iptc": {
    "Name": "Spider-Man",
    "Caption": "Spider-Man swings on the web",
    "Copyright Notice": "Daily Bugle",
    "Keywords": ["spider-man", "menance", "offender"]
  }
}
```

#### Example (mp4)

```json
{
  "format": "mov,mp4,m4a,3gp,3g2,mj2",
  "width": 1178,
  "height": 730,
  "size": 984963,
  "exif": {},
  "video_meta": {
    "com.android.version": "9",
    "compatible_brands": "isommp42",
    "creation_time": "2022-01-12T15:04:10.000000Z",
    "location": "+46.4845+030.6848/",
    "location-eng": "+46.4845+030.6848/",
    "major_brand": "mp42",
    "minor_version": "0"
  }
}
```
