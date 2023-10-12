# Signing the URL

imgproxy allows you to sign your URLs with a key and salt, so an attacker won’t be able to perform a denial-of-service attack by requesting multiple different image resizes.

### Configuring URL signature

URL signature checking is disabled by default, but it is highly recommended to enable it in a production environment. To do so, define a key/salt pair by setting the following environment variables:

* `IMGPROXY_KEY`: hex-encoded key
* `IMGPROXY_SALT`: hex-encoded salt

Read our [Configuration](configuration.md#url-signature) guide to learn more ways of setting keys and salts.

If you need a random key/salt pair in a hurry, you can quickly generate one using the following snippet:

```bash
echo $(xxd -g 2 -l 64 -p /dev/random | tr -d '\n')
```

### Calculating URL signature

A signature is a URL-safe Base64-encoded HMAC digest of the rest of the path, including the leading `/`. Here’s how it’s calculated:


* Take the part of the path after the signature:
  * For [processing URLs](generating_the_url.md): `/%processing_options/%encoded_url.%extension`, `/%processing_options/plain/%plain_url@%extension`, or `/%processing_options/enc/%encrypted_url.%extension`
  * For [info URLs](getting_the_image_info.md): `/%encoded_url`, `/plain/%plain_url`, or `/enc/%encrypted_url`
* Add a salt to the beginning.
* Calculate the HMAC digest using SHA256.
* Encode the result with URL-safe Base64.

### Example

**You can find helpful code snippets in various programming languages the [examples](https://github.com/imgproxy/imgproxy/tree/master/examples) folder. There's a good chance you'll find a snippet in your favorite programming language that you'll be able to use right away.**

And here is a step-by-step example of URL signature creation:

Assume that you have the following unsigned URL:

```
http://imgproxy.example.com/insecure/rs:fill:300:400:0/g:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

To sign it, you need to configure imgproxy to use your key/salt pair. Let's say, your key and salt are `secret` and `hello`, respectively — that translates to `736563726574` and `68656C6C6F` in hex encoding. This key/salt pair is quite weak for production purposes but will do for this example. Run imgproxy using this key/salt pair, like so:

```bash
IMGPROXY_KEY=736563726574 IMGPROXY_SALT=68656C6C6F imgproxy
```

Note that all your unsigned URL will stop working since imgproxy now checks all URL signatures.

First, you need to take the path after the signature and add the salt to the beginning:

```
hello/rs:fill:300:400:0/g:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

Then calculate the HMAC digest of this string using SHA256 and encode it with URL-safe Base64:

```
oKfUtW34Dvo2BGQehJFR4Nr0_rIjOtdtzJ3QFsUcXH8
```

And finally, add the signature to your URL:

```
http://imgproxy.example.com/oKfUtW34Dvo2BGQehJFR4Nr0_rIjOtdtzJ3QFsUcXH8/rs:fill:300:400:0/g:sm/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

Now you have a URL that you can use to securely resize the image.
