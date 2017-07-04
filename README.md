# Imgproxy

Fast and secure microservice for resizing and converting remote images.

<a href="https://evilmartians.com/?utm_source=overmind">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54">
</a>

Imgproxy does one thing, and it does it well: resizing of remote images. It works great when you need to resize some images on the fly to make them look good on your web page. The main principles of Imgproxy are simplicity, speed, and security.

#### Simplicity

One of the things I believe in is: "The best feature is the one you don't need to implement." That's why I implemented only features that most of us need.

* Rotation, flip, flop, etc. are good, but it's better to do this with CSS;
* Caching is good, but it's better to use CDN or caching server for this;
* HTTPS is good, but it's better to use TCP proxy-server like NGINX.

#### Speed

Imgproxy uses probably the most efficient image processing library - libvips. It's fast and requires low memory footprint. Thus it allows processing a massive amount of images on the fly.

Also, imgproxy uses native Go's net/http routing for an absolute speed.

#### Security

Processing of remote images is a quite vulnerable thing. There are many ways to attack you, so it's a good idea to take measures to prevent attacks. There is what imgproxy does:

* It checks image type and dimensions while downloading, so the image won't be fully downloaded if it has an unknown format or too big dimensions. Thus imgproxy protects you from image bombs like https://www.bamsoftware.com/hacks/deflate.html

* Imgproxy protects its URL path with a signature, so it can't be easily compromised by an attacker. Thus imgproxy doesn't allow to use itself by third-party applications.

* Imgproxy supports authorization by HTTP header. This prevents using imgproxy directly by an attacker but allows to use it through CDN or a caching server.

## Installation

There are two ways you can currently install imgproxy:

#### From the source

1. Install [vips](https://github.com/jcupitt/libvips)

  ```bash
  # macOS
  $ brew tap homebrew/science
  $ brew install vips

  # Ubintu
  $ sudo apt-get install libvips
  ```

2. Install imgproxy itself

  ```bash
  $ go get github.com/DarthSim/imgproxy
  ```

#### Docker

```bash
$ docker build -t imgproxy .
```

## Configuration

Imgproxy is 12factor-ready and can be configured with env variables.

#### Path signature

Imgproxy requires all paths to be signed with key and salt:

* IMGPROXY_KEY - (**required**) hex-encoded key;
* IMGPROXY_SALT - (**required**) hex-encoded salt;

You can also specify paths to a files with hex-encoded key and salt (useful in a development evironment):

```bash
$ imgproxy -keypath /path/to/file/with/key -saltpath /path/to/file/with/salt
```

You can easily generate key and salt with `openssl enc -aes-256-cbc -P -md sha256`.

#### Server

* IMGPROXY_BIND - TCP address to listen on. Default: :8080;
* IMGPROXY_READ_TIMEOUT - the maximum duration (seconds) for reading the entire request, including the body. Default: 10;
* IMGPROXY_WRITE_TIMEOUT - the maximum duration (seconds) for writing the response. Default: 10;

#### Security

Imgproxy protects you from image bombs. Here you can specify maximum image dimension which you're ready to process:

* IMGPROXY_MAX_SRC_DIMENSION - the maximum dimension of source image. Images with larger size will be rejected. Default: 4096;

Also you can specify secret to enable authorization with HTTP `Authorization` header:

* IMGPROXY_SECRET - auth token. If specified, request should contain `Authorization: Bearer %secret%` header;

#### Compression

* IMGPROXY_QUALITY - quality of a result image. Default: 80;
* IMGPROXY_GZIP_COMPRESSION - GZip compression level. Default: 5;

## Generating url

Url path should contain signature and resizing params like this:

```
/%signature/%resizing_type/%width/%height/%gravity/%enlarge/%encoded_url.%extension
```

#### Resizing type

Imgproxy supports the following resizing types:

* `fit` - resizes image keeping aspect ratio to fit given size;
* `fill` - resizes image keeping aspect ratio to fill given size and crops projecting parts;
* `crop` - crops image to given size;
* `force` - resizes image to given size breaking aspect ratio.

#### Width and height

Width and height define size of the result image. The result dimensions may be not equal to the given depending on what resizing type was applied.

#### Gravity

When imgproxy needs to cut some parts of the image, it's guided by gravity. The following values are supported:

* `no` - north (top edge);
* `so` - south (bottom edge);
* `ea` - east (right edge);
* `we` - west (left edge);
* `ce` - center;
* `sm` - smart. Vips detects the most interesting section of the image and considers it as the center of the result image. **Note:** This value applicable only to the crop resizing.

#### Enlarge

This param is `0`, imgproxy won't enlarge image if it's smaller that given size. With any other value imgproxy will enlarge image.

#### Encoded url

The source url should be encoded with url-safe base64. Encoded url can be splitted with `/` for your needs.

#### Extension

Extension specifies the format of the result image. Imgproxy supports only `jpg` and `png` as the most popular web-image formats.

#### Signature

Signature is a url-safe base64-encoded HMAC digest of the rest of the path including leading `/`.

* Take the path after signature - `/%resizing_type/%width/%height/%gravity/%enlarge/%encoded_url.%extension`;
* Add salt to the beginning;
* Calc HMAC digest using SHA256;
* Encode the result with url-secure base64.

You can find code snippets in the `examples` folder.

## Source images formats support

Imgproxy supports only three most popular images formats: PNG, JPEG and GIF.

**Known issue:** Libvips may not support some kinds of JPEG, if you met this issue, you may need to build libvips with ImageMagick or GraphicMagick support. See https://github.com/jcupitt/libvips#imagemagick-or-optionally-graphicsmagick

## Special thanks

Special thanks to [h2non](https://github.com/h2non) and all authors and contributors of [bimg](https://github.com/h2non/bimg).

## Author

Sergey "DarthSim" Aleksandrovich

## License

Imgproxy is licensed under the MIT license.

See LICENSE for the full license text.
