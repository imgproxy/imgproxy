require "openssl"
require "base64"

key = ["943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881"].pack("H*")
salt = ["520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5"].pack("H*")

url = "http://img.example.com/pretty/image.jpg"

# You can trim padding spaces to get good-looking url
encoded_url = Base64.urlsafe_encode64(url).tr("=", "").scan(/.{1,16}/).join("/")

resize = "fill"
width = 300
height = 300
gravity = "no"
enlarge = 1
extension = "png"

path = "/rs:#{resize}:#{width}:#{height}:#{enlarge}/g:#{gravity}/#{encoded_url}.#{extension}"

digest = OpenSSL::Digest.new("sha256")
# You can trim padding spaces to get good-looking url
hmac = Base64.urlsafe_encode64(OpenSSL::HMAC.digest(digest, key, "#{salt}#{path}")).tr("=", "")

signed_path = "/#{hmac}#{path}"
