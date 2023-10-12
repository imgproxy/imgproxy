const crypto = require('crypto');

const KEY = Buffer.from(
  '1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1',
  'hex',
)

const encrypt = (target, key) => {
  const data = Buffer.from(target).toString('binary');

	// We use a random iv generation, but you'll probably want to use some
	// deterministic method
  const iv = crypto.randomBytes(16)

  const cipher = crypto.createCipheriv('aes-256-cbc', key, iv);

  let encrypted = Buffer.from(
    cipher.update(data, 'utf8', 'binary') + cipher.final('binary'),
    'binary',
  );

  return Buffer.concat([iv, encrypted]).toString('base64').replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_')
}

const url = 'http://img.example.com/pretty/image.jpg'
const encrypted_url = encrypt(url, KEY)

// We don't sign the URL in this example but it is highly recommended to sign
// imgproxy URLs when imgproxy is being used in production.
// Signing URLs is especially important when using encrypted source URLs to
// prevent a padding oracle attack
const path = `/unsafe/rs:fit:300:300/enc/${encrypted_url}.jpg`

console.log(path)
