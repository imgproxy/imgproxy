# Installation

There are three ways you can install imgproxy:

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

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/DarthSim/imgproxy)

However, you can do it manually with a few steps:

```bash
$ git clone https://github.com/DarthSim/imgproxy.git && cd imgproxy
$ heroku create your-application
$ heroku stack:set container
$ git push heroku master
```

### From the source

1. First, install [libvips](https://github.com/libvips/libvips).

  ```bash
  # macOS
  $ brew install vips

  # Ubuntu
  # Ubuntu apt repository contains a pretty old version of libvips.
  # It's recommended to use PPA with an up to date version.
  $ sudo add-apt-repository ppa:dhor/myway
  $ sudo apt-get install libvips-dev

  # FreeBSD
  $ pkg install -y pkgconf vips
  ```

2. Next, install imgproxy itself:

  ```bash
  $ CGO_LDFLAGS_ALLOW="-s|-w" go get -f -u github.com/DarthSim/imgproxy
  ```

**Note:** imgproxy build is automatically tested with the latest three minor versions of Go and libvips. Successfully build with the older versions is not guaranteed but yet may be possible.
