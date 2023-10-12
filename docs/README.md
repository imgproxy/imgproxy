# imgproxy

imgproxy is a fast and secure standalone server for resizing and converting remote images. The guiding principles behind imgproxy are security, speed, and simplicity.

imgproxy is able to quickly and easily resize images on the fly, and it's well-equipped to handle a large amount of image resizing. imgproxy is a fast, secure replacement for all the image resizing code inside your web application (such as resizing libraries, or code that calls ImageMagick or GraphicsMagic). It's also an indispensable tool for processing images from a remote source. With imgproxy, you don’t need to repeatedly prepare images to fit your design every time it changes.

To get an even better introduction, and to dive deeper into the nitty gritty details, check out this article: [imgproxy: Resize your images instantly and securely](https://evilmartians.com/chronicles/introducing-imgproxy)

<a href="https://evilmartians.com/?utm_source=imgproxy" target="_blank">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians_v2.0_for-dark-bg.svg" alt="Sponsored by Evil Martians" width="236" height="54">
</a>

#### Simplicity

> "No code is better than no code."

imgproxy only includes the must-have features for image processing, fine-tuning and security. Specifically,

* It would be great to be able to rotate, flip and apply masks to images, but in most of the cases, it is possible — and is much easier — to do that using CSS3.
* It may be great to have built-in HTTP caching of some kind, but it is way better to use a Content-Delivery Network or a caching proxy server for this, as you will have to do this sooner or later in the production environment.
* It might be useful to have everything built in — such as HTTPS support — but an easy way to solve that would be just to use a proxying HTTP server such as nginx.

#### Speed

imgproxy takes advantage of probably the most efficient image processing library out there – `libvips`. It’s scary fast and comes with a very low memory footprint. Thanks to libvips, we can readily and extemporaneously process a massive amount of images.

imgproxy uses Go’s raw (no wrappers) native `net/http` package to omit any overhead while processing requests and provides the best possible HTTP support.

You can take a look at some benchmarking results and compare imgproxy with some well-known alternatives in our [benchmark report](https://github.com/imgproxy/imgproxy/blob/master/BENCHMARK.md).

#### Security

In terms of security, the massive processing of remote images is a potentially dangerous endeavor. There are a number of possible attack vectors, so it’s a good idea to take an approach that considers attack prevention measures as a priority. Here’s how imgproxy does this:

* imgproxy checks the image type and its “real” dimensions when downloading. The image will not be fully downloaded if it has an unknown format or if the dimensions are too big (you can set the max allowed dimensions). This is how imgproxy protects from so called "image bombs”, like those described in [this doc](https://www.bamsoftware.com/hacks/deflate.html).

* imgproxy protects image URLs with a signature, so an attacker cannot enact a denial-of-service attack by requesting multiple image resizes.

* imgproxy supports authorization by HTTP header. This prevents imgproxy from being used directly by an attacker, but allows it to be used via a CDN or a caching server — simply by adding a header to a proxy or CDN config.

## Author

Sergey "[DarthSim](https://github.com/DarthSim)" Alexandrovich

## Special thanks

Many thanks to:

* [Roman Shamin](https://github.com/romashamin) for the awesome logo.
* [Alena Kirdina](https://github.com/egodyston) and [Alexander Madyankin](https://github.com/madyankin) for the great website.
* [John Cupitt](https://github.com/jcupitt) for developing [libvips](https://github.com/libvips/libvips) and for helping me optimize its usage with imgproxy.
* [Kirill Kuznetsov](https://github.com/dragonsmith) for the [Helm chart](https://github.com/imgproxy/imgproxy-helm).
* [Travis Turner](https://github.com/Travis-Turner) for keeping the documentation in good shape.

## License

imgproxy is licensed under the MIT license.

See [LICENSE](https://github.com/imgproxy/imgproxy/blob/master/LICENSE) for the full license text.
