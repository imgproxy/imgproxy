FROM alpine:edge
MAINTAINER Sergey Aleksandrovich <darthsim@gmail.com>

ENV GO_DOWNLOAD_URL https://golang.org/dl/go1.8.3.linux-amd64.tar.gz
ENV GO_DOWNLOAD_SHA_256 1862f4c3d3907e59b04a757cfda0ea7aa9ef39274af99a784f5be843c80c6772
ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH /usr/local/go/bin:$PATH

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ADD . /go/src/github.com/DarthSim/imgproxy

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
  && apk add --no-cache --update bash vips \
  && apk add --no-cache --virtual .build-deps curl gcc musl-dev fftw-dev vips-dev \
  && curl -L "$GO_DOWNLOAD_URL" -o /golang.tar.gz \
  && echo "$GO_DOWNLOAD_SHA_256  /golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf /golang.tar.gz \
  && cd /go/src/github.com/DarthSim/imgproxy \
  && go build -v -o /usr/local/bin/imgproxy \
  && apk del --purge .build-deps \
  && rm -rf /golang.tar.gz /usr/local/go /var/cache/apk*

CMD ["imgproxy"]

EXPOSE 8080
