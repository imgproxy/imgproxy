# Getting the image info <img class="pro-badge" src="assets/pro.svg" alt="pro" />

imgproxy can fetch and return the source image info without downloading the whole image.

## URL format

To get the image info, use the following URL format:

```
/info/%signature/plain/%source_url
/info/%signature/%encoded_source_url
```

### Signature

Signature protects your URL from being modified by an attacker. It is highly recommended to sign imgproxy URLs in a production environment.

Once you set up your [URL signature](configuration.md#url-signature), check out the [Signing the URL](signing_the_url.md) guide to learn about how to sign your URLs. Otherwise, use any string here.

### Source URL

There are two ways to specify source url:

#### Plain

The source URL can be provided as is, prepended by `/plain/` part:

```
/plain/http://example.com/images/curiosity.jpg
```

**📝Note:** If the sorce URL contains query string or `@`, you need to escape it.

#### Base64 encoded

The source URL can be encoded with URL-safe Base64. The encoded URL can be split with `/` for your needs:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn
```

## Response format

imgproxy responses with JSON body and returns the following info:

* `format`: source image/video format. In case of video - list of predicted formats divided by comma;
* `width`: image/video width;
* `height`: image/video height;
* `size`: file size. Can be zero if the image source doesn't set `Content-Length` header properly;
* `exif`: JPEG exif data.

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
  }
}
```

#### Exampple (mp4)

```json
{
  "format": "mov,mp4,m4a,3gp,3g2,mj2",
  "width": 1178,
  "height": 730,
  "size": 984963,
  "exif": {}
}
```
