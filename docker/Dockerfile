FROM golang:1-buster
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

ARG VIPS_VERSION=8.9.2
ARG IMAGEMAGICK_VERSION=7.0.9-20
ARG QUANTIZR_VERSION=0.2.0
ARG LIBHEIF_VERSION=1.6.2
ARG LIBTIFF_VERSION=4.1.0

ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH /lib64:/usr/lib64:/usr/local/lib
ENV CGO_LDFLAGS_ALLOW "-s|-w"

# Install dependencies
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    curl \
    git \
    ca-certificates \
    build-essential \
    cmake \
    libtool \
    libglib2.0-dev \
    libexpat1-dev \
    libjpeg62-turbo-dev \
    libpng-dev \
    libwebp-dev \
    libgif-dev \
    librsvg2-dev \
    libexif-dev \
    liblcms2-dev \
    libde265-dev \
    liblzma-dev \
    libzstd-dev

# Install quantizr
RUN cd /root \
  && curl -Ls https://github.com/DarthSim/quantizr/releases/download/v$QUANTIZR_VERSION/quantizr-imagequant-compat_$QUANTIZR_VERSION-0.deb \
    > quantizr.deb \
  && apt-get install -y ./quantizr.deb

# Build libheif
RUN cd /root \
  && curl -Ls https://github.com/strukturag/libheif/releases/download/v$LIBHEIF_VERSION/libheif-$LIBHEIF_VERSION.tar.gz | tar -xz \
  && cd libheif-$LIBHEIF_VERSION \
  && ./configure \
    --prefix=/usr/local \
    --disable-static \
    --disable-go \
  && make install-strip

# Build libtiff
RUN cd /root \
  && curl -Ls https://gitlab.com/libtiff/libtiff/-/archive/v$LIBTIFF_VERSION/libtiff-v$LIBTIFF_VERSION.tar.gz | tar -xz \
  && cd libtiff-v$LIBTIFF_VERSION \
  && cmake . \
  && make \
  && make install

# Build ImageMagick
RUN cd /root \
  && curl -Ls https://github.com/ImageMagick/ImageMagick/archive/$IMAGEMAGICK_VERSION.tar.gz | tar -xz \
  && cd ImageMagick-$IMAGEMAGICK_VERSION \
  && ./configure \
    --enable-silent-rules \
    --disable-static \
    --disable-openmp \
    --disable-deprecated \
    --disable-docs \
    --with-threads \
    --without-magick-plus-plus \
    --without-utilities \
    --without-perl \
    --without-bzlib \
    --without-dps \
    --without-freetype \
    --without-fontconfig \
    --without-jbig \
    --without-jpeg \
    --without-lcms \
    --without-lzma \
    --without-png \
    --without-tiff \
    --without-wmf \
    --without-xml \
    --without-webp \
    --without-heic \
    --without-pango \
  && make install-strip \
  && rm -rf /usr/local/lib/libMagickWand-7.*

# Build libvips
RUN cd /root \
  && curl -Ls https://github.com/libvips/libvips/releases/download/v$VIPS_VERSION/vips-$VIPS_VERSION.tar.gz | tar -xz \
  && cd vips-$VIPS_VERSION \
  && ./configure \
    --without-python \
    --without-OpenEXR \
    --enable-debug=no \
    --disable-static \
    --disable-introspection \
    --enable-silent-rules \
  && make install-strip \
  && rm -rf /usr/local/lib/libvips-cpp.*

WORKDIR /app
COPY . .

# Build imgproxy
RUN go build -v -o /usr/local/bin/imgproxy

# ==================================================================================================
# Final image

FROM debian:buster-slim
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    libsm6 \
    libglib2.0-0 \
    libexpat1 \
    libjpeg62-turbo \
    libpng16-16 \
    libwebp6 \
    libwebpmux3 \
    libwebpdemux2 \
    libgif7 \
    librsvg2-2 \
    libexif12 \
    liblcms2-2 \
    libde265-0 \
    liblzma5 \
    libzstd1 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=0 /usr/local/lib /usr/local/lib
COPY --from=0 /usr/lib/libquantizr.so /usr/local/lib

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV LD_LIBRARY_PATH /lib64:/usr/lib64:/usr/local/lib

CMD ["imgproxy"]

EXPOSE 8080
