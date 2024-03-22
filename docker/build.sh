#!/bin/bash

set -e

# This is pretty dirty hack. Building imgproxy under Qemu is pretty slow.
# So we install Go binary native for the BUILDARCH.
if [[ $BUILDARCH != $TARGETARCH ]]; then
  GOLANG_VERSION=$(go version | sed -E 's/.*go([0-9]+\.[0-9]+(\.[0-9]+)?).*/\1/')

  rm -rf /usr/local/go

  apt-get update
  apt-get install -y --no-install-recommends libstdc++6:${BUILDARCH}

  curl -Ls https://golang.org/dl/go${GOLANG_VERSION}.linux-${BUILDARCH}.tar.gz \
    | tar -xzC /usr/local

  export CGO_ENABLED=1
  export GOOS=linux
  export GOARCH=$TARGETARCH
fi

go build -v -ldflags "-s -w" -o /usr/local/bin/imgproxy
