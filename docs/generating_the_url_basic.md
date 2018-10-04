# Generating the URL (Basic)

This guide describes the simple URL format that is easy to use but doesn't support the whole range of imgproxy features. This URL format is mostly backward-compatible with the first version of imgproxy. Read our [Generating the URL (Advanced)](../docs/generating_the_url_advanced.md) guide to get info about advanced URL format.

### Format definition

The basic URL should contain the signature, resize parameters, and encoded source URL, like this:

```
/%signature/%resizing_type/%width/%height/%gravity/%enlarge/%encoded_url.%extension
```

Check out the [example](#example) at the end of this guide.

#### Signature

Signature protects your URL from being changed by an attacker. It's highly recommended to sign imgproxy URLs in production.

If you set up [URL signature](../docs/configuration.md#url-signature), check out [Signing the URL](../docs/signing_the_url.md) guide to know how to sign your URLs. Otherwise, use any string here.

#### Resizing types

imgproxy supports the following resizing types:

* `fit` — resizes the image while keeping aspect ratio to fit given size;
* `fill` — resizes the image while keeping aspect ratio to fill given size and cropping projecting parts;
* `crop` — crops the image to a given size.

#### Width and height

Width and height parameters define the size of the resulting image. Depending on the resizing type applied, the dimensions may differ from the requested ones.

#### Gravity

When imgproxy needs to cut some parts of the image, it is guided by the gravity. The following values are supported:

* `no` — north (top edge);
* `so` — south (bottom edge);
* `ea` — east (right edge);
* `we` — west (left edge);
* `ce` — center;
* `sm` — smart. `libvips` detects the most "interesting" section of the image and considers it as the center of the resulting image;
* `fp:%x:%y` - focus point. `x` and `y` are floating point numbers between 0 and 1 that describe coordinates of the center of the resulting image. Trait 0 and 1 as right/left for `x` and top/bottom for `y`.

#### Enlarge

If set to `0`, imgproxy will not enlarge the image if it is smaller than the given size. With any other value, imgproxy will enlarge the image.

#### Encoded URL

The source URL should be encoded with URL-safe Base64. The encoded URL can be split with `/` for your needs.

#### Extension

Extension specifies the format of the resulting image. At the moment, imgproxy supports only `jpg`, `png` and `webp`, them being the most popular and useful web image formats.

The extension part can be omitted. In this case, imgproxy will use `jpg` by default. You also can [enable WebP support detection](../docs/configuration.md#webp-support-detection) to use it as default resulting format when possible.

### Example

Signed imgproxy URL that resizes `http://example.com/images/curiosity.jpg` to fill `300x400` area with smart gravity without enlarging, and converts the image to `png` will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/fill/300/400/sm/0/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```
