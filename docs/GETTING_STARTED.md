# Getting started

This guide will show you how to quickly get you first image resized with imgproxy.

## Install

Let's assume you have Docker installed on your machine. Then you can just pull official imgproxy image, and you're done!

```bash
$ docker pull darthsim/imgproxy:latest
$ docker run -p 8080:8080 -it darthsim/imgproxy
```

If you don't have docker, you can use Heroku for a quick start.

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

Check out our [installation guide](../docs/installation.md) for more details and instructions.

That's it! No further configuration is needed, but if you want to unleash the full power of imgproxy, read our [configuration guide](../docs/configuration.md).

## Resize an image

After you installed imgproxy, you can use the following URL to get the resized image of Matt Damon from "The Martian" movie (replace `localhost:8080` with your domain if you installed imgproxy on a remote server):
[http://localhost:8080/insecure/fill/300/400/sm/0/aHR0cHM6Ly9tLm1l/ZGlhLWFtYXpvbi5j/b20vaW1hZ2VzL00v/TVY1Qk1tUTNabVk0/TnpZdFkyVm1ZaTAw/WkRSbUxUZ3lPREF0/WldZelpqaGxOemsx/TnpVMlhrRXlYa0Zx/Y0dkZVFYVnlOVGMz/TWpVek5USUAuanBn.jpg](http://localhost:8080/insecure/fill/300/400/sm/0/aHR0cHM6Ly9tLm1l/ZGlhLWFtYXpvbi5j/b20vaW1hZ2VzL00v/TVY1Qk1tUTNabVk0/TnpZdFkyVm1ZaTAw/WkRSbUxUZ3lPREF0/WldZelpqaGxOemsx/TnpVMlhrRXlYa0Zx/Y0dkZVFYVnlOVGMz/TWpVek5USUAuanBn.jpg)

[The original image](https://m.media-amazon.com/images/M/MV5BMmQ3ZmY4NzYtY2VmYi00ZDRmLTgyODAtZWYzZjhlNzk1NzU2XkEyXkFqcGdeQXVyNTc3MjUzNTI@.jpg) is resized to fill `300x400` with smart gravity. `libvips` chose the most interesting part of the image.

Get more info about generation imgproxy URLs in the [Generating the URL](../docs/generating_the_url_basic.md) guide.

## Security

Note that this URL is not signed. It's highly recommended to sign URLs in production. Read our [Signing the URL](../docs/signing_the_url.md) guide to know how to secure your imgproxy from attackers.
