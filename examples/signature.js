const createHmac = require('create-hmac')

const KEY = '943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881'
const SALT = '520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5'

const urlSafeBase64 = (string) => {
  return Buffer.from(string).toString('base64').replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_')
}

const hexDecode = (hex) => Buffer.from(hex, 'hex')

const sign = (salt, target, secret) => {
  const hmac = createHmac('sha256', hexDecode(secret))
  hmac.update(hexDecode(salt))
  hmac.update(target)
  return urlSafeBase64(hmac.digest())
}

const url = 'http://img.example.com/pretty/image.jpg'
const resizing_type = 'fill'
const width = 300
const height = 300
const gravity = 'no'
const enlarge = 1
const extension = 'png'
const encoded_url = urlSafeBase64(url)
const path = `/rs:${resizing_type}:${width}:${height}:${enlarge}/g:${gravity}/${encoded_url}.${extension}`

const signature = sign(SALT, path, KEY)
const result = `/${signature}${path}`
console.log(result)
