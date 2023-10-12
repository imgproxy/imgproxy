# Encrypting the source URL![pro](./assets/pro.svg)

If you don't want to reveal your source URLs, you can encrypt them with the AES-CBC algorithm.

### Configuring source URL encryption

The only thing needed for source URL encryption is a key:

* `IMGPROXY_SOURCE_URL_ENCRYPTION_KEY`: hex-encoded key used for source URL encryption. Default: blank

The key should be either 16, 24, or 32 bytes long for AES-128-CBC, AES-192-CBC, or AES-256-CBC, respectively.

If you need a random key in a hurry, you can quickly generate one using the following snippet:

```bash
echo $(xxd -g 2 -l 32 -p /dev/random | tr -d '\n')
```

### Encrypting the source URL

* Pad your source URL using the [PKCS #7](https://en.wikipedia.org/wiki/Padding_(cryptography)#PKCS#5_and_PKCS#7) method so it becomes 16-byte aligned. Some libraries like Ruby's `openssl` do the message padding for you
* Generate a 16-byte long initialization vector (IV)
* Encrypt the padded source URL with the AES-CBC algorithm using the configured key and the IV generated in the previous step
* Create the following string: IV + encrypted URL
* Encode the result of the previous step with URL-safe Base64

#### IV generation

AES-CBC requires IV to be unique between unencrypted messages (source URLs in our case). Usually, it's recommended to use a counter when generating an IV to be sure it never repeats. However, in our case, this leads to a major drawback: using a unique IV every time you encrypt the same source URL will lead to different cipher texts and thus different imgproxy URLs. And this leads to a situation where requests to imgproxy will never hit the CDNs cache.

On the other hand, reusing the IV with the same message is safe but ONLY while with this message. Thus, there are some tradeoffs:

1. Cache IVs. Store IV somewhere so you need to generate it only once for each source URL and extract it if needed. Depending on the level of security you need, you may also want to encrypt stored IVs with a different key so a DB leak won't reveal the message-IV pairs.
2. Use a deterministic method of generation. For example, you can calculate an HMAC hash of the plain source URL with a different key and truncate it to the IV size. Though this method doesn't guarantee that it will always generate unique IVs, the chances of generating repeatable IVs with it are considerably rare.

### Example

**You can find helpful code snippets in various programming languages in the [examples](https://github.com/imgproxy/imgproxy/tree/master/examples) folder. There's a good chance you'll find a snippet in your favorite programming language that you'll be able to use right away.**

And here is a step-by-step example of a source URL encryption:

Before we start, we need an encryption key. We will use the AES-256-CBC algorithm in this example, so we need a 32-byte key. Let's assume we used a random generator and got the following hex-encoded key:

```
1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1
```

Run imgproxy using this encryption key, like so:

```bash
IMGPROXY_SOURCE_URL_ENCRYPTION_KEY="1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1" imgproxy
```

Next, assume that you have the following source URL:

```
http://example.com/images/curiosity.jpg
```

It's 39-byte long, so we should align it to 16 bytes using the PKCS #7 method:

```
http://example.com/images/curiosity.jpg\09\09\09\09\09\09\09\09\09
```

**📝 Note:** From this point on, we'll show unprintable characters in `\NN` format where `NN` is a hex representation of the byte.

Next, we need an initialization vector (IV). Let's assume we generated the following IV:

```
\A7\95\63\A2\B3\5D\86\CE\E6\45\1C\3C\80\0F\53\5A
```

We'll use our encription key and IV encrypt a 16-byte-aligned source URL with the AES-256-CBC algorithm:

```
\84\65\19\C8\B7\97\59\2E\CE\A3\78\DD\44\25\45\A4\48\43\4A\AD\04\A5\B7\A8\50\01\22\CC\7E\65\1C\FF\71\57\3C\89\54\D8\6E\1B\0D\B3\13\41\2F\50\47\69
```

Add the IV to the beginning:

```
\A7\95\63\A2\B3\5D\86\CE\E6\45\1C\3C\80\0F\53\5A\84\65\19\C8\B7\97\59\2E\CE\A3\78\DD\44\25\45\A4\48\43\4A\AD\04\A5\B7\A8\50\01\22\CC\7E\65\1C\FF\71\57\3C\89\54\D8\6E\1B\0D\B3\13\41\2F\50\47\69
```

And finally, encode the result with URL-safe Base64:

```
p5VjorNdhs7mRRw8gA9TWoRlGci3l1kuzqN43UQlRaRIQ0qtBKW3qFABIsx-ZRz_cVc8iVTYbhsNsxNBL1BHaQ
```

Now you can put this encrypted URL in the imgproxy URL path, prepending it with the `/enc/` segment:

```
/unsafe/rs:fit:300:300/enc/p5VjorNdhs7mRRw8gA9TWoRlGci3l1kuzqN43UQlRaRIQ0qtBKW3qFABIsx-ZRz_cVc8iVTYbhsNsxNBL1BHaQ
```

**📝 Note:** The imgproxy URL in this example is not signed but signing URLs is especially important when using encrypted source URLs to prevent a padding oracle attack.
