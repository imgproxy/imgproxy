# Getting started

This guide will show you how to get your first image resized with imgproxy quickly.

## Install

Let's assume you have Docker installed on your machine. Then you can pull an official imgproxy image, and you're done!

```bash
docker pull darthsim/imgproxy:latest
docker run -p 8080:8080 -it darthsim/imgproxy
```

If you don't have docker, you can use Heroku for a quick start.

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/DarthSim/imgproxy)

Check out our [installation guide](installation.md) for more details and instructions.

That's it! No further configuration is needed, but if you want to unleash the full power of imgproxy, read our [configuration guide](configuration.md).

## Resize an image

After you've successfully installed imgproxy, you might want to see if it is working correctly. To check that, you can use the following URL to get the resized image of Matt Damon from "The Martian" movie (replace `localhost:8080` with your domain if you installed imgproxy on a remote server):

[http://localhost:8080/insecure/rs:fill:300:400/g:sm/aHR0cHM6Ly9tLm1l/ZGlhLWFtYXpvbi5j/b20vaW1hZ2VzL00v/TVY1Qk1tUTNabVk0/TnpZdFkyVm1ZaTAw/WkRSbUxUZ3lPREF0/WldZelpqaGxOemsx/TnpVMlhrRXlYa0Zx/Y0dkZVFYVnlOVGMz/TWpVek5USUAuanBn.jpg](http://localhost:8080/insecure/rs:fill:300:400/g:sm/aHR0cHM6Ly9tLm1l/ZGlhLWFtYXpvbi5j/b20vaW1hZ2VzL00v/TVY1Qk1tUTNabVk0/TnpZdFkyVm1ZaTAw/WkRSbUxUZ3lPREF0/WldZelpqaGxOemsx/TnpVMlhrRXlYa0Zx/Y0dkZVFYVnlOVGMz/TWpVek5USUAuanBn.jpg)

Here's [the original image](https://m.media-amazon.com/images/M/MV5BMmQ3ZmY4NzYtY2VmYi00ZDRmLTgyODAtZWYzZjhlNzk1NzU2XkEyXkFqcGdeQXVyNTc3MjUzNTI@.jpg), just for reference. Using the URL above, imgproxy is told to resize it to fill the area of `300x400` size with "smart" gravity. "Smart" means that the `libvips` library chooses the most "interesting" part of the image.

Learn more on how to generate imgproxy URLs in the [Generating the URL](generating_the_url.md) guide.

## Security

Note that the URL in the above example is not signed. It is highly recommended to sign URLs in production. Read our [Signing the URL](signing_the_url.md) guide to know how to secure your imgproxy installation from attackers.
