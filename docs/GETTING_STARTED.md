# Getting started

This guide will show you how to quickly resize your first image using imgproxy.

## Install

Let's assume you already have Docker installed on your machine — you can pull an official imgproxy Docker image, and you’re done!

```bash
docker pull darthsim/imgproxy:latest
docker run -p 8080:8080 -it darthsim/imgproxy
```

If you don't have docker, you can use Heroku for a quick start.

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/DarthSim/imgproxy)

Check out our [installation guide](installation.md) for more details and instructions.

In both cases, that's it! No further configuration is needed, but if you want to unleash the full power of imgproxy, read our [configuration guide](configuration.md).

## Resize an image

After you’ve successfully installed imgproxy, a good first step is to make sure that everything is working correctly. To do that, you can use the following URL to get a resized image of Matt Damon from “The Martian” (replace `localhost:8080` with your domain if you’ve installed imgproxy on a remote server):

[http://localhost:8080/insecure/rs:fill:300:400/g:sm/aHR0cHM6Ly9tLm1l/ZGlhLWFtYXpvbi5j/b20vaW1hZ2VzL00v/TVY1Qk1tUTNabVk0/TnpZdFkyVm1ZaTAw/WkRSbUxUZ3lPREF0/WldZelpqaGxOemsx/TnpVMlhrRXlYa0Zx/Y0dkZVFYVnlOVGMz/TWpVek5USUAuanBn.jpg](http://localhost:8080/insecure/rs:fill:300:400/g:sm/aHR0cHM6Ly9tLm1l/ZGlhLWFtYXpvbi5j/b20vaW1hZ2VzL00v/TVY1Qk1tUTNabVk0/TnpZdFkyVm1ZaTAw/WkRSbUxUZ3lPREF0/WldZelpqaGxOemsx/TnpVMlhrRXlYa0Zx/Y0dkZVFYVnlOVGMz/TWpVek5USUAuanBn.jpg)

Just for reference, here’s [the original image](https://m.media-amazon.com/images/M/MV5BMmQ3ZmY4NzYtY2VmYi00ZDRmLTgyODAtZWYzZjhlNzk1NzU2XkEyXkFqcGdeQXVyNTc3MjUzNTI@.jpg). Using the URL above, imgproxy is instructed to resize it to fill an area of `300x400` size with “smart” gravity. “Smart” means that the `libvips` library chooses the most “interesting” part of the image.

You can learn more on how to generate imgproxy URLs in the [Generating the URL](generating_the_url.md) guide.

## Security

Note that the URL in the above example is not signed. However, it’s highly recommended to use signed URLs in production. Read our [Signing the URL](signing_the_url.md) guide to learn how to secure your imgproxy installation from attackers.
