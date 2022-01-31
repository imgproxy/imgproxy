#!/bin/bash

set -e

TAG_NAME=${1:-latest}
REPO_NAME="darthsim/imgproxy"

IMAGE_NAME="$REPO_NAME:$TAG_NAME"

echo "Image name: $IMAGE_NAME"

export DOCKER_CLI_EXPERIMENTAL=enabled

docker push $IMAGE_NAME-amd64
docker push $IMAGE_NAME-arm64

push_manifest() {
  docker manifest create $1 -a $2-amd64 -a $2-arm64
  docker manifest annotate $1 $2-amd64 --arch amd64
  docker manifest annotate $1 $2-arm64 --arch arm64 --variant v8
  docker manifest push $1
}

push_manifest $IMAGE_NAME $IMAGE_NAME

re="^v([0-9]+)\.([0-9]+)\.([0-9]+)$"

if [[ $TAG_NAME =~ $re ]]; then
  MINOR_IMAGE_NAME="$REPO_NAME:v${BASH_REMATCH[1]}.${BASH_REMATCH[2]}"
  push_manifest $MINOR_IMAGE_NAME $IMAGE_NAME

  MAJOR_IMAGE_NAME="$REPO_NAME:v${BASH_REMATCH[1]}"
  push_manifest $MAJOR_IMAGE_NAME $IMAGE_NAME
fi
