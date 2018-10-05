# Signing the URL

imgproxy allows you to sign your URLs with key and salt, so an attacker won't be able to cause a denial-of-service attack by requesting multiple image resizes.

### Configuring URL signature

URL signature checking is disabled by default, but it's highly recommended to enable it in production. To do so, define key/salt pair by setting the environment variables:

* `IMGPROXY_KEY` — hex-encoded key;
* `IMGPROXY_SALT` — hex-encoded salt;

Read our [Configuration](./configuration.md#url-signature) guide to find more ways to set key and salt.

If you need a random key/salt pair real fast, you can quickly generate it using, for example, the following snippet:

```bash
$ echo $(xxd -g 2 -l 64 -p /dev/random | tr -d '\n')
```

### Calculating URL signature

Signature is a URL-safe Base64-encoded HMAC digest of the rest of the path including the leading `/`. Here's how it is calculated:

* Take the path after the signature:
  * For [basic URL format](./generating_the_url_basic.md) - `/%resizing_type/%width/%height/%gravity/%enlarge/%encoded_url.%extension`;
  * For [advanced URL format](./generating_the_url_advanced.md) - `/%processing_options/%encoded_url.%extension`;
* Add salt to the beginning;
* Calculate the HMAC digest using SHA256;
* Encode the result with URL-safe Base64.

### Example

You can find helpful code snippets in the [examples](../examples) folder. And here is a step-by-step example of calculating URL signature:

Assume that you have the following unsigned URL:

```
http://imgproxy.example.com/insecure/fill/300/400/sm/0/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png
```

To sign it, you need to configure imgproxy to use your key/salt pair. Let's say, your key and salt are `secret` and `hello` that will be `736563726574` and `68656C6C6F` in hex encoding. This key/salt pair is obviously weak for production but ok for this example. Run your imgproxy using this key and salt:

```bash
$ IMGPROXY_KEY=736563726574 IMGPROXY_SALT=68656C6C6F imgproxy
```

Note that your unsigned URL will stop work because imgproxy now checks signatures of all URLs.

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

Now you got the URL that you can use to securely resize the image.
