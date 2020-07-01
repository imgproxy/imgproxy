# Signing the URL

imgproxy allows you to sign your URLs with key and salt, so an attacker won't be able to cause a denial-of-service attack by requesting multiple different image resizes.

### Configuring URL signature

URL signature checking is disabled by default, but it is highly recommended to enable it in a production environment. To do so, the define key/salt pair by setting the following environment variables:

* `IMGPROXY_KEY`: hex-encoded key;
* `IMGPROXY_SALT`: hex-encoded salt;

Read our [Configuration](configuration.md#url-signature) guide to find more ways to set key and salt.

If you need a random key/salt pair real fast, you can quickly generate it using, for example, the following snippet:

```bash
echo $(xxd -g 2 -l 64 -p /dev/random | tr -d '\n')
```

### Calculating URL signature

Signature is an URL-safe Base64-encoded HMAC digest of the rest of the path, including the leading `/`. Here is how it is calculated:

* Take the path part after the signature:
  * For [basic URL format](generating_the_url_basic.md): `/%resizing_type/%width/%height/%gravity/%enlarge/%encoded_url.%extension` or `/%resizing_type/%width/%height/%gravity/%enlarge/plain/%plain_url@%extension`;
  * For [advanced URL format](generating_the_url_advanced.md): `/%processing_options/%encoded_url.%extension` or `/%processing_options/plain/%plain_url@%extension`;
  * For [info URL](getting_the_image_info.md): `/%encoded_url` or `/plain/%plain_url`;
* Add salt to the beginning;
* Calculate the HMAC digest using SHA256;
* Encode the result with URL-safe Base64.

### Example

**You can find helpful code snippets in various programming languages the [examples](https://github.com/imgproxy/imgproxy/tree/master/examples) folder. There is a good chance you will find a snippet in your favorite programming language that you can use right away.**

And here is a step-by-step example of calculating the URL signature:

Assume that you have the following unsigned URL:

```
http://imgproxy.example.com/insecure/fill/300/400/sm/0/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

To sign it, you need to configure imgproxy to use your key/salt pair. Let's say, your key and salt are `secret` and `hello` — that translates to `736563726574` and `68656C6C6F` in hex encoding. This key/salt pair is quite weak for production use but will do for this example. Run your imgproxy using this key/salt pair:

```bash
IMGPROXY_KEY=736563726574 IMGPROXY_SALT=68656C6C6F imgproxy
```

Note that all your unsigned URL will stop working since imgproxy now checks signatures of all URLs.

First, you need to take the path after the signature and add the salt to the beginning:

```
hello/fill/300/400/sm/0/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

Then calculate the HMAC digest of this string using SHA256 and encode it with URL-safe Base64:

```
AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM
```

And finally put the signature to your URL:

```
http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/fill/300/400/sm/0/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

Now you got the URL that you can use to resize the image securely.
