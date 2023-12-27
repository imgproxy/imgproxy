var crypto = require('crypto');

var KEY = '943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881'
var SALT = '520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5'

var hexDecode = (hex) => String.bytesFrom(hex, 'hex')

var sign = (salt, target, secret) => {
    var hmac = crypto.createHmac('sha256', hexDecode(secret))
    hmac.update(hexDecode(salt))
    hmac.update(target)
    
    return hmac.digest('base64url')
}

var path = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg"
var signature = sign(SALT,path,KEY)
var result = `/${signature}${path}`
console.log(result)
