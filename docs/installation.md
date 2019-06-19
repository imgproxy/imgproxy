# Installation

There are four ways you can install imgproxy:

### Docker

imgproxy can (and should) be used as a standalone application inside a Docker container. Just pull the official image from Docker Hub:

```bash
$ docker pull darthsim/imgproxy:latest
$ docker run -p 8080:8080 -it darthsim/imgproxy
```

You can also build your own image. imgproxy is ready to be dockerized, plug and play:

```bash
$ docker build -t imgproxy .
$ docker run -p 8080:8080 -it imgproxy
```

### Heroku

imgproxy can be deployed to Heroku with a click of a button:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/imgproxy/imgproxy)

However, you can do it manually with a few steps:

```bash
$ git clone https://github.com/imgproxy/imgproxy.git && cd imgproxy
$ heroku create your-application
$ heroku stack:set container
$ git push heroku master
```

### Packages

#### Arch Linux and derivatives

[imgproxy](https://aur.archlinux.org/packages/imgproxy/) package is available from AUR.

### From the source

#### Ubuntu

First, install [libvips](https://github.com/libvips/libvips).

Ubuntu apt repository contains a pretty old version of libvips. You can use PPA with more recent version of libvips:

```bash
$ sudo add-apt-repository ppa:dhor/myway
$ sudo apt-get update
$ sudo apt-get install libvips-dev
```

But if you want to use all the features of imgproxy, it's recommended to build libvips from the source: [https://github.com/libvips/ libvips/wiki/Build-for-Ubuntu](https://github.com/libvips/libvips/wiki/Build-for-Ubuntu)

Next, install the latest Go:

```bash
$ sudo add-apt-repository ppa:longsleep/golang-backports
$ sudo apt-get update
$ sudo apt-get install golang-go
```

And finally, install imgproxy itself:

```bash
$ CGO_LDFLAGS_ALLOW="-s|-w" go get -f -u github.com/imgproxy/imgproxy
```

#### macOS + Homebrew

```bash
$ brew install vips go
$ PKG_CONFIG_PATH="$(brew --prefix libffi)/lib/pkgconfig" \
  CGO_LDFLAGS_ALLOW="-s|-w" \
  CGO_CFLAGS_ALLOW="-Xpreprocessor" \
  go get -f -u github.com/imgproxy/imgproxy
```
