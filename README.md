# Imgproxy

Fast and secure micro-service for resizing and converting remote images.

Imgproxy does one thing, and it does it well: resizing of remote images. It works great when you need to resize some images on the fly to make them look good on your web page. The main principles of Imgproxy are simplicity, speed, and security.

#### Simlicity

One of the things I believe in is: "The best feature is the one you don't need to implement." That's why I implemented only features that most of us need. Rotation, flip, flop, etc. are cool, but I don't think that you want to process your web page images that ways, especially when you can do this with CSS.

#### Speed

Imgproxy uses probably the most efficient image processing library - libvips. It's fast and requires low memory footprint. Thus it allows processing a massive amount of images on the fly.

Also, imgproxy uses native Go's net/http routing for an absolute speed.

#### Security

Processing of remote images is a quite vulnerable thing. There are many ways to attack you, so it's a good idea to take measures to prevent attacks. There is what imgproxy does:

* It checks image type and dimensions while downloading, so the image won't be fully downloaded if it has an unknown format or too big dimensions. Thus imgproxy protects you from image bombs like https://www.bamsoftware.com/hacks/deflate.html

* Imgproxy protects its URL path with a signature, so it can't be easily compromised by an attacker. Thus imgproxy doesn't allow to use itself by third-party applications.

* Imgproxy supports authorization by HTTP header. This prevents using imgproxy directly by an attacker but allows to use it through CDN or a caching server.

### How to install

1. Install [vips](https://github.com/jcupitt/libvips). On macOS you can do:

  ```
  $ brew tap homebrew/science
  $ brew install vips
  ```

2. Install imgproxy itself

  ```
  $ go get github.com/DarthSim/imgproxy
  ```

### How to configure

Imgproxy is 12factor-ready and can be configured with env variables:

* IMGPROXY_BIND - TCP address to listen on. Default: :8080;
* IMGPROXY_READ_TIMEOUT - the maximum duration (seconds) for reading the entire request, including the body. Default: 10;
* IMGPROXY_WRITE_TIMEOUT - the maximum duration (seconds) for writing the response. Default: 10;
* IMGPROXY_MAX_SRC_DIMENSION - the maximum dimension of source image. Images with larger size will be rejected. Default: 4096;
* IMGPROXY_QUALITY - quality of a result image. Default: 80;
* IMGPROXY_GZIP_COMPRESSION - GZip compression level. Default: 5;
* IMGPROXY_KEY - hex-encoded key
* IMGPROXY_SALT - hex-encoded salt

You can also specify paths to a files with hex-encoded key and salt:

```
imgproxy -keypath /path/to/file/with/key -saltpath /path/to/file/with/salt
```

### How to generate url path

Here is a short ruby sample which shows how to generate url path for imgproxy.

```ruby
require 'openssl'
require 'base64'

# Key and salt. Since they're hex-encoded, we should decode it.
key = ['943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881'].pack("H*")
salt = ['520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5'].pack("H*")

# This is remote url with requested image
url = "http://img.example.com/pretty/image.jpg"

# Url should be encoded with base64 and could be splitted
encodedUrl = Base64.urlsafe_encode64(url).tr("=", "").scan(/.{1,16}/).join("/")

# Allowed values for resize are: fill, fit, crop and resize
resize = 'fill'
width = 300
height = 300
# Allowed values for gravity are: no (north), so (south), ea (east), we (west)
# ce (center) and sm (smart). "sm" works correctly only with resize == crop.
gravity = 'no'
# Should we enlarge small images? 1 for yes, and 0 for no.
enlarge = 1
# Allowed extensions are png and jpg/jpeg.
extension = 'png'

path = "/#{resize}/#{width}/#{height}/#{gravity}/#{enlarge}/#{encodedUrl}.#{extension}"

# Now we need to sign path with HMAC (SHA256)
digest = OpenSSL::Digest.new('sha256')
hmac = Base64.urlsafe_encode64(OpenSSL::HMAC.digest(digest, key, "#{salt}#{path}")).tr('=', '')

signed_path = "/#{hmac}#{path}"
```
