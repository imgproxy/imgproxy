require "openssl"
require "base64"

key = ["1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1"].pack("H*")

url = "http://img.example.com/pretty/image.jpg"

# The key is 32 bytes long, so we use AES-256-CBC
cipher = OpenSSL::Cipher::AES.new(256, :CBC)
cipher.encrypt

# We use a random iv generation, but you'll probably want to use some
# deterministic method
iv = cipher.random_iv

cipher.key = key
cipher.iv = iv

encrypted_url = Base64.urlsafe_encode64(iv + cipher.update(url) + cipher.final).tr("=", "")

# We don't sign the URL in this example but it is highly recommended to sign
# imgproxy URLs when imgproxy is being used in production.
# Signing URLs is especially important when using encrypted source URLs to
# prevent a padding oracle attack
path = "/unsafe/rs:fit:300:300/enc/#{encrypted_url}.jpg"
