#!/bin/bash

set -e

# This is pretty dirty hack. Building imgproxy under Qemu is pretty slow.
# So we install Go binary native for the BUILDPLATFORM.
if [[ $BUILDPLATFORM != $TARGETPLATFORM ]]; then
  case "$BUILDPLATFORM" in
    amd64 | "linux/amd64")
      BUILD_ARCH="amd64"
      ;;

    arm64 | "arm64/v8" | "linux/arm64" | "linux/arm64/v8")
      BUILD_ARCH="arm64"
      ;;

    *)
      echo "Unknown platform: $BUILDPLATFORM"
      exit 1
  esac

  case "$TARGETPLATFORM" in
    amd64 | "linux/amd64")
      TARGET_ARCH="amd64"
      ;;

    arm64 | "arm64/v8" | "linux/arm64" | "linux/arm64/v8")
      TARGET_ARCH="arm64"
      ;;

    *)
      echo "Unknown platform: $TARGETPLATFORM"
      exit 1
  esac

  GOLANG_VERSION=$(go version | sed -E 's/.*go([0-9]+\.[0-9]+(\.[0-9]+)?).*/\1/')

  rm -rf /usr/local/go

  dpkg --add-architecture ${BUILD_ARCH}
  apt-get update
  apt-get install -y --no-install-recommends libstdc++6:${BUILD_ARCH}

  curl -Ls https://golang.org/dl/go${GOLANG_VERSION}.linux-${BUILD_ARCH}.tar.gz \
    | tar -xzC /usr/local

  export CGO_ENABLED=1
  export GOOS=linux
  export GOARCH=$TARGET_ARCH
fi

go build -v -ldflags "-s -w" -o /usr/local/bin/imgproxy
