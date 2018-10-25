FROM alpine:edge
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$PATH

ADD . /go/src/github.com/DarthSim/imgproxy

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
  && apk --no-cache upgrade \
  && apk add --no-cache --virtual .build-deps go gcc musl-dev fftw-dev vips-dev \
  && cd /go/src/github.com/DarthSim/imgproxy \
  && CGO_LDFLAGS_ALLOW="-s|-w" go build -v -o /usr/local/bin/imgproxy \
  && apk del --purge .build-deps \
  && rm -rf /var/cache/apk*

FROM alpine:edge
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
  && apk --no-cache upgrade \
  && apk add --no-cache ca-certificates bash vips \
  && rm -rf /var/cache/apk*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin

CMD ["imgproxy"]

EXPOSE 8080
