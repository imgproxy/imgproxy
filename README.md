# imgproxy

<img align="right" width="200" height="200" title="imgproxy logo"
     src="https://cdn.rawgit.com/DarthSim/imgproxy/master/logo.svg">


[![CircleCI branch](https://img.shields.io/circleci/project/github/imgproxy/imgproxy/master.svg?style=for-the-badge)](https://circleci.com/gh/DarthSim/imgproxy) [![Docker](https://img.shields.io/badge/docker-darthsim%2Fimgproxy-blue.svg?style=for-the-badge)](https://hub.docker.com/r/darthsim/imgproxy/) [![MicroBadger Size](https://img.shields.io/microbadger/image-size/darthsim/imgproxy.svg?style=for-the-badge)](https://hub.docker.com/r/darthsim/imgproxy/) [![Docker Pulls](https://img.shields.io/docker/pulls/darthsim/imgproxy.svg?style=for-the-badge)](https://hub.docker.com/r/darthsim/imgproxy/)


imgproxy is a fast and secure standalone server for resizing and converting remote images. The main principles of imgproxy are simplicity, speed, and security.

imgproxy can be used to provide a fast and secure way to replace all the image resizing code of your web application (like calling ImageMagick or GraphicsMagick, or using libraries), while also being able to resize everything on the fly, fast and easy. imgproxy is also indispensable when handling lots of image resizing, especially when images come from a remote source.

imgproxy does one thing — resizing remote images — and does it well. It works great when you need to resize multiple images on the fly to make them match your application design without preparing a ton of cached resized images or re-doing it every time the design changes.

imgproxy is a Go application, ready to be installed and used in any Unix environment — also ready to be containerized using Docker.

See this article for a good intro and all the juicy details! [imgproxy:
Resize your images instantly and securely](https://evilmartians.com/chronicles/introducing-imgproxy)

<a href="https://evilmartians.com/?utm_source=imgproxy">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54">
</a>

#### Simplicity

> "No code is better than no code."

imgproxy only includes the must-have features for image processing, fine-tuning and security. Specifically,

* It would be great to be able to rotate, flip and apply masks to images, but in most of the cases, it is possible — and is much easier — to do that using CSS3.
* It may be great to have built-in HTTP caching of some kind, but it is way better to use a Content-Delivery Network or a caching proxy server for this, as you will have to do this sooner or later in the production environment.
* It might be useful to have everything built in — such as HTTPS support — but an easy way to solve that would be just to use a proxying HTTP server such as nginx.

#### Speed

imgproxy uses probably the most efficient image processing library there is, called `libvips`. It is screaming fast and has a very low memory footprint; with it, we can handle the processing for a massive amount of images on the fly.

imgproxy also uses native Go's `net/http` routing for the best HTTP networking support.

You can see benchmarking results and comparison with some well-known alternatives in our [benchmark report](./BENCHMARK.md).

#### Security

Massive processing of remote images is a potentially dangerous thing, security-wise. There are many attack vectors, so it is a good idea to consider attack prevention measures first. Here is how imgproxy can help:

* imgproxy checks image type _and_ "real" dimensions when downloading, so the image will not be fully downloaded if it has an unknown format or the dimensions are too big (there is a setting for that). That is how imgproxy protects you from so called "image bombs" like those described at  https://www.bamsoftware.com/hacks/deflate.html.

* imgproxy protects image URLs with a signature, so an attacker cannot cause a denial-of-service attack by requesting multiple image resizes.

* imgproxy supports authorization by an HTTP header. That prevents using imgproxy directly by an attacker but allows to use it through a CDN or a caching server — just by adding a header to a proxy or CDN config.

## Documentation

1. [Getting started](./docs/GETTING_STARTED.md)
2. [Installation](./docs/installation.md)
   * [Docker](./docs/installation.md#docker)
   * [Heroku](./docs/installation.md#heroku)
   * [From the source](./docs/installation.md#from-the-source)
3. [Configuration](./docs/configuration.md)
   * [URL signature](./docs/configuration.md#url-signature)
   * [Server](./docs/configuration.md#server)
   * [Security](./docs/configuration.md#security)
   * [Compression](./docs/configuration.md#compression)
   * [WebP support detection](./docs/configuration.md#webp-support-detection)
   * [Client Hints support](./docs/configuration.md#client-hints-support)
   * [Watermark](./docs/configuration.md#watermark)
   * [Presets](./docs/configuration.md#presets)
   * [Serving local files](./docs/configuration.md#serving-local-files)
   * [Serving files from Amazon S3](./docs/configuration.md#serving-files-from-amazon-s3)
   * [Serving files from Google Cloud Storage](./docs/configuration.md#serving-files-from-google-cloud-storage)
   * [New Relic metrics](./docs/configuration.md#new-relic-metrics)
   * [Prometheus metrics](./docs/configuration.md#prometheus-metrics)
   * [Error reporting](./docs/configuration.md#error-reporting)
   * [Syslog](./docs/configuration.md#syslog)
   * [Memory usage tweaks](./docs/configuration.md#memory-usage-tweaks)
   * [Miscellaneous](./docs/configuration.md#miscellaneous)
4. [Generating the URL](./docs/generating_the_url_basic.md)
   * [Basic](./docs/generating_the_url_basic.md)
   * [Advanced](./docs/generating_the_url_advanced.md)
   * [Signing the URL](./docs/signing_the_url.md)
5. [Watermark](./docs/watermark.md)
6. [Presets](./docs/presets.md)
7. [Serving local files](./docs/serving_local_files.md)
8. [Serving files from Amazon S3](./docs/serving_files_from_s3.md)
9. [Serving files from Google Cloud Storage](./docs/serving_files_from_google_cloud_storage.md)
10. [New Relic](./docs/new_relic.md)
11. [Prometheus](./docs/prometheus.md)
12. [Image formats support](./docs/image_formats_support.md)
13. [About processing pipeline](./docs/about_processing_pipeline.md)
14. [Health check](./docs/healthcheck.md)
15. [Memory usage tweaks](./docs/memory_usage_tweaks.md)

## Author

Sergey "DarthSim" Alexandrovich

Many thanks to [Roman Shamin](https://github.com/romashamin) for the awesome logo.

Great bunch of kudos goes to [John Cupitt](https://github.com/jcupitt) who develops [libvips](https://github.com/libvips/libvips) and helps me to optimize its usage under the hood of imgproxy.

## License

imgproxy is licensed under the MIT license.

See LICENSE for the full license text.
