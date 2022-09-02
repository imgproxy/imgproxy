import base64
import hashlib
import hmac


key = bytes.fromhex("943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881")
salt = bytes.fromhex("520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5")


path = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg".encode()

digest = hmac.new(key, msg=salt+path, digestmod=hashlib.sha256).digest()
signature = base64.urlsafe_b64encode(digest).rstrip(b"=")

url = b'/%s%s' % (
    signature,
    path,
)

print(url.decode())
