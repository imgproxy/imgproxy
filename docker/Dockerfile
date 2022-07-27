ARG BASE_IMAGE_VERSION="v3.3.0"

FROM darthsim/imgproxy-base:${BASE_IMAGE_VERSION}

ARG BUILDPLATFORM
ARG TARGETPLATFORM

COPY . .
RUN docker/build.sh

# ==================================================================================================
# Final image

FROM debian:bullseye-slim
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN apt-get update \
  && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    libsm6 \
    liblzma5 \
    libzstd1 \
    libpcre3 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=0 /usr/local/lib /usr/local/lib

COPY NOTICE /usr/local/share/doc/imgproxy/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV LD_LIBRARY_PATH /usr/local/lib

RUN groupadd -r imgproxy && useradd -r -u 999 -g imgproxy imgproxy
USER 999

CMD ["imgproxy"]

EXPOSE 8080
