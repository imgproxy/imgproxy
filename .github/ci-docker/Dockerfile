FROM public.ecr.aws/ubuntu/ubuntu:noble

ARG VIPS_VERSIONS="8.14 8.15 8.16"

RUN apt-get -qq update \
  && apt-get install -y --no-install-recommends \
    bash \
    curl \
    git \
    ca-certificates \
    build-essential \
    gobject-introspection \
    libgirepository1.0-dev \
    python3-pip \
    python3-venv \
    libssl-dev \
    libglib2.0-dev \
    libxml2-dev \
    libjpeg-dev \
    libpng-dev \
    libwebp-dev \
    librsvg2-dev \
    libexif-dev \
    liblcms2-dev \
  && python3 -m venv /root/.python \
  && /root/.python/bin/pip install meson ninja \
  && rm -rf /var/lib/apt/lists/*

RUN curl https://sh.rustup.rs -sSf | sh -s -- -y \
  && export PATH="/root/.cargo/bin:$PATH" \
  && cargo install cargo-c \
  && cd /root \
  && git clone --depth 1 https://github.com/DarthSim/quantizr.git \
  && cd quantizr \
  && cargo cinstall --release --library-type=cdylib --prefix=/usr/local --libdir=/usr/local/lib \
  && rm -rf /root/.rustup /root/.cargo

ENV PATH="/root/.python/bin:$PATH"
ENV LD_LIBRARY_PATH="/usr/local/lib"

RUN \
  mkdir /root/vips \
    && cd /root/vips \
    && curl -s -S -L -o vips_releases.json "https://api.github.com/repos/libvips/libvips/releases" \
    && for VIPS_VERSION in $VIPS_VERSIONS; do \
      mkdir $VIPS_VERSION \
      && export VIPS_RELEASE=$(grep -m 1 "\"tag_name\": \"v$VIPS_VERSION." vips_releases.json | sed -E 's/.*"v([^"]+)".*/\1/') \
      && echo "Building Vips $VIPS_RELEASE as $VIPS_VERSION" \
      && curl -s -S -L -o libvips-$VIPS_RELEASE.tar.gz https://github.com/libvips/libvips/archive/refs/tags/v$VIPS_RELEASE.tar.gz \
      && tar -xzf libvips-$VIPS_RELEASE.tar.gz \
      && cd libvips-$VIPS_RELEASE \
      && meson setup _build \
        --buildtype=release \
        --strip \
        --prefix=/root/vips/$VIPS_VERSION \
        --libdir=lib \
        -Dgtk_doc=false \
      && ninja -C _build \
      && ninja -C _build install \
      && cd .. \
      && rm -rf libvips-$VIPS_RELEASE.tar.gz libvips-$VIPS_RELEASE; \
    done

WORKDIR /go/src

ENTRYPOINT [ "/bin/bash" ]
