# Installation

There are four ways you can install imgproxy:

## Docker

imgproxy can (and this is highly recommended) be used as a standalone application inside a Docker container. Just pull the official image from Docker Hub:

```bash
docker pull darthsim/imgproxy:latest
docker run -p 8080:8080 -it darthsim/imgproxy
```

You can also build your own image. imgproxy is ready to be dockerized out of the box:

```bash
docker build -f docker/Dockerfile -t imgproxy .
docker run -p 8080:8080 -it imgproxy
```

## Helm

imgproxy can be easily deployed to your Kubernetes cluster using Helm and our official Helm chart:

```bash
helm repo add imgproxy https://helm.imgproxy.net/

# With Helm 3
helm upgrade -i imgproxy imgproxy/imgproxy

# With Helm 2
helm upgrade -i --name imgproxy imgproxy/imgproxy
```

Check out the [chart's README](https://github.com/imgproxy/imgproxy-helm) for more info.

## Heroku

imgproxy can be deployed to Heroku with the click of a button:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/imgproxy/imgproxy)

That being said, you can also do it manually in just a few steps:

```bash
git clone https://github.com/imgproxy/imgproxy.git && cd imgproxy
heroku create your-application
heroku stack:set container
git push heroku master
```

## Packages

### Arch Linux and derivatives

[imgproxy](https://aur.archlinux.org/packages/imgproxy/) package is available from AUR.

### macOS + Homebrew

[imgproxy](https://formulae.brew.sh/formula/imgproxy) is available from Homebrew:
```bash
brew install imgproxy
```

## From the source

You can get the imgproxy source code by cloning the GitHub repo:

```bash
git clone https://github.com/imgproxy/imgproxy.git
cd imgproxy
```

...or by downloading the source tarball:

```bash
mkdir imgproxy
cd imgproxy
curl -Ls https://github.com/imgproxy/imgproxy/archive/master.tar.gz \
  | tar -xz --strip-components 1 -C .
```

You can also download a specific version:

```bash
mkdir imgproxy
cd imgproxy
curl -Ls https://github.com/imgproxy/imgproxy/archive/v2.13.1.tar.gz \
  | tar -xz --strip-components 1 -C .
```

### Ubuntu

First, install [libvips](https://github.com/libvips/libvips).

The Ubuntu apt repository contains a pretty old version of libvips. You can use PPA to access a  more recent version of libvips:

```bash
sudo add-apt-repository ppa:dhor/myway
sudo apt-get update
sudo apt-get install libvips-dev
```

But if you want to use all the features of imgproxy, it's recommended to build libvips from the source: [https://github.com/libvips/ libvips/wiki/Build-for-Ubuntu](https://github.com/libvips/libvips/wiki/Build-for-Ubuntu)

Next, install the latest version of Go:

```bash
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install golang-go
```

And finally, install imgproxy itself:

```bash
CGO_LDFLAGS_ALLOW="-s|-w" \
  go build -o /usr/local/bin/imgproxy
```

### macOS + Homebrew

```bash
brew install vips go
PKG_CONFIG_PATH="$(brew --prefix libffi)/lib/pkgconfig" \
  CGO_LDFLAGS_ALLOW="-s|-w" \
  CGO_CFLAGS_ALLOW="-Xpreprocessor" \
  go build -o /usr/local/bin/imgproxy
```
