ARG BASE_IMAGE_VERSION="v3.10.0"

FROM --platform=${BUILDPLATFORM} darthsim/imgproxy-base:${BASE_IMAGE_VERSION} as base-native

FROM darthsim/imgproxy-base:${BASE_IMAGE_VERSION} as build

ARG BUILDARCH
ARG TARGETARCH

RUN dpkg --add-architecture ${BUILDARCH} \
  && apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends libstdc++6:${BUILDARCH}

# This is pretty dirty hack. Building imgproxy under Qemu is pretty slow.
# So we install Go binary native for the BUILDARCH.
RUN rm -rf /usr/local/go
COPY --from=base-native /usr/local/go /usr/local/go

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=$TARGETARCH

COPY . .
RUN bash -c 'go build -v -ldflags "-s -w" -o /usr/local/bin/imgproxy'

# ==================================================================================================
# Final image

FROM ubuntu:noble
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN apt-get update \
  && apt-get upgrade -y \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    libstdc++6 \
    liblzma5 \
    libzstd1 \
    fontconfig-config \
    fonts-dejavu-core \
    media-types \
    libjemalloc2 \
    libtcmalloc-minimal4 \
  && ln -s /usr/lib/$(uname -m)-linux-gnu/libjemalloc.so.2 /usr/local/lib/libjemalloc.so \
  && ln -s /usr/lib/$(uname -m)-linux-gnu/libtcmalloc_minimal.so.4 /usr/local/lib/libtcmalloc_minimal.so \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=build /usr/local/lib /usr/local/lib

COPY docker/entrypoint.sh /usr/local/bin/

# AWS Lambda adapter
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.3 /lambda-adapter /opt/extensions/lambda-adapter

COPY NOTICE /usr/local/share/doc/imgproxy/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV LD_LIBRARY_PATH /usr/local/lib
ENV IMGPROXY_MALLOC malloc
ENV AWS_LWA_READINESS_CHECK_PATH /health
ENV AWS_LWA_INVOKE_MODE response_stream

# Disable SVE on ARM64. SVE is slower than NEON on Amazon Graviton 3
ENV VIPS_VECTOR=167772160

RUN groupadd -r imgproxy \
  && useradd -r -u 999 -g imgproxy imgproxy \
  && mkdir -p /var/cache/fontconfig \
  && chmod 777 /var/cache/fontconfig
USER 999

ENTRYPOINT [ "entrypoint.sh" ]
CMD ["imgproxy"]

EXPOSE 8080
