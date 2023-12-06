ARG BASE_IMAGE_VERSION="v3.8.3"

FROM darthsim/imgproxy-base:${BASE_IMAGE_VERSION}

ARG BUILDPLATFORM
ARG TARGETPLATFORM

COPY . .
RUN docker/build.sh

# ==================================================================================================
# Final image

FROM debian:stable-slim
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN apt-get update \
  && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    libstdc++6 \
    liblzma5 \
    libzstd1 \
    fontconfig-config \
    media-types \
    libjemalloc2 \
    libtcmalloc-minimal4 \
  && ln -s /usr/lib/$(uname -m)-linux-gnu/libjemalloc.so.2 /usr/local/lib/libjemalloc.so \
  && ln -s /usr/lib/$(uname -m)-linux-gnu/libtcmalloc_minimal.so.4 /usr/local/lib/libtcmalloc_minimal.so \
  && rm -rf /var/lib/apt/lists/*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=0 /usr/local/lib /usr/local/lib

COPY docker/entrypoint.sh /usr/local/bin/

COPY NOTICE /usr/local/share/doc/imgproxy/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV LD_LIBRARY_PATH /usr/local/lib
ENV IMGPROXY_MALLOC malloc

RUN groupadd -r imgproxy \
  && useradd -r -u 999 -g imgproxy imgproxy \
  && mkdir -p /var/cache/fontconfig \
  && chmod 777 /var/cache/fontconfig
USER 999

ENTRYPOINT [ "entrypoint.sh" ]
CMD ["imgproxy"]

EXPOSE 8080
