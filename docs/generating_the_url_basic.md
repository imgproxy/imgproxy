# Generating the URL (Basic)

This guide describes the simple URL format that is easy to use but doesn't support the whole range of imgproxy features. This URL format is mostly supported for backwards compatibility with imgproxy 1.x. Please read our [Generating the URL (Advanced)](./generating_the_url_advanced.md) guide to learn about the advanced URL format.

### Format definition

The basic URL should contain the signature, resize parameters, and source URL, like this:

```
/%signature/%resizing_type/%width/%height/%gravity/%enlarge/plain/%source_url@%extension
/%signature/%resizing_type/%width/%height/%gravity/%enlarge/%encoded_source_url.%extension
```

Check out the [example](#example) at the end of this guide.

#### Signature

Signature protects your URL from being modified by an attacker. It is highly recommended to sign imgproxy URLs in a production environment.

Once you set up your [URL signature](./configuration.md#url-signature), check out the [Signing the URL](./signing_the_url.md) guide to learn about how to sign your URLs. Otherwise, use any string here.

#### Resizing types

imgproxy supports the following resizing types:

* `fit`: resizes the image while keeping aspect ratio to fit given size;
* `fill`: resizes the image while keeping aspect ratio to fill given size and cropping projecting parts;
* `auto`: if both source and resulting dimensions have the same orientation (portrait or landscape), imgproxy will use `fill`. Otherwise, it will use `fit`.

#### Width and height

Width and height parameters define the size of the resulting image in pixels. Depending on the resizing type applied, the dimensions may differ from the requested ones.

#### Gravity

When imgproxy needs to cut some parts of the image, it is guided by the gravity. The following values are supported:

* `no`: north (top edge);
* `so`: south (bottom edge);
* `ea`: east (right edge);
* `we`: west (left edge);
* `noea`: north-east (top-right corner);
* `nowe`: north-west (top-left corner);
* `soea`: south-east (bottom-right corner);
* `sowe`: south-west (bottom-left corner);
* `ce`: center;
* `sm`: smart. `libvips` detects the most "interesting" section of the image and considers it as the center of the resulting image;
* `fp:%x:%y` - focus point. `x` and `y` are floating point numbers between 0 and 1 that describe the coordinates of the center of the resulting image. Treat 0 and 1 as right/left for `x` and top/bottom for `y`.

#### Enlarge

When set to `0`, imgproxy will not enlarge the image if it is smaller than the given size. With any other value, imgproxy will enlarge the image.

#### Source URL

There are two ways to specify source url:

##### Plain

The source URL can be provided as is, prendended by `/plain/` part:

```
/plain/http://example.com/images/curiosity.jpg
```

**Note:** If the sorce URL contains query string or `@`, you need to escape it.

When using plain source URL, you can specify the [extension](#extension) after `@`:

```
/plain/http://example.com/images/curiosity.jpg@png
```

##### Base64 encoded

The source URL can be encoded with URL-safe Base64. The encoded URL can be split with `/` for your needs:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn
```

When using encoded source URL, you can specify the [extension](#extension) after `.`:

```
/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

#### Extension

Extension specifies the format of the resulting image. At the moment, imgproxy supports only `jpg`, `png`, `webp`, `gif`, and `ico`, them being the most popular and useful image formats on the Web.

**Note:** Read about GIF support [here](./image_formats_support.md#gif-support).

The extension part can be omitted. In this case, imgproxy will use source image format as resulting one. If source image format is not supported as resulting, imgproxy will use `jpg`. You also can [enable WebP support detection](./configuration.md#webp-support-detection) to use it as default resulting format when possible.

### Example

Signed imgproxy URL that resizes `http://example.com/images/curiosity.jpg` to fill `300x400` area with smart gravity without enlarging, and converts the image to `png`:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/fill/300/400/sm/0/plain/http://example.com/images/curiosity.jpg@png
```

The same URL with Base64-encoded source URL will look like this:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/fill/300/400/sm/0/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```
