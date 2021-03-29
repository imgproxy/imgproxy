#!/bin/bash -xe
# This script is for building and pushing to pushd ECR docker repo.

docker build -t $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pushd/imgproxy:$(git rev-parse HEAD)  -f docker/Dockerfile .
docker push $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pushd/imgproxy:$(git rev-parse HEAD)