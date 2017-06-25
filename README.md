# Imgproxy

Tiny, fast and secure server for processing remote images.

Full README is on the way.

### How to configure

Imgproxy is 12factor-ready and can be configured with env variables:

* IMGPROXY_BIND - TCP address to listen on. Default: :8080;
* IMGPROXY_READ_TIMEOUT - the maximum duration (seconds) for reading the entire request, including the body. Default: 10;
* IMGPROXY_WRITE_TIMEOUT - the maximum duration (seconds) for writing the response. Default: 10;
* IMGPROXY_MAX_SRC_DIMENSION - the maximum dimension of source image. Images with larger size will be rejected. Default: 4096;
* IMGPROXY_QUALITY - quality of a result image. Default: 80;
* IMGPROXY_COMPRESSION - compression of a result image. Default: 6;
* IMGPROXY_KEY - hex-encoded key
* IMGPROXY_SALT - hex-encoded salt

You can also specify paths to a files with hex-encoded key and salt:

```
imgproxy -keypath /path/to/file/with/key -salt /path/to/file/with/salt
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
