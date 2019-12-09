FROM golang:1-buster
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

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
    libtool \
    libfftw3-dev \
    libglib2.0-dev \
    libexpat1-dev \
    libjpeg62-turbo-dev \
    libpng-dev \
    libwebp-dev \
    libgif-dev \
    librsvg2-dev \
    libexif-dev \
    liblcms2-dev \
    libheif-dev \
    libtiff-dev \
    libimagequant-dev

# Build ImageMagick
RUN cd /root \
  && mkdir ImageMagick \
  && curl -Ls https://imagemagick.org/download/ImageMagick.tar.gz | tar -xz -C ImageMagick --strip-components 1 \
  && cd ImageMagick \
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
  && export VIPS_VERSION=$(curl -s "https://api.github.com/repos/libvips/libvips/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/') \
  && echo "Vips version: $VIPS_VERSION" \
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
    libfftw3-3 \
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
    libheif1 \
    libtiff5 \
    libimagequant0 \
    libjemalloc2 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=0 /usr/local/lib /usr/local/lib

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=4
ENV LD_LIBRARY_PATH /lib64:/usr/lib64:/usr/local/lib
ENV LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libjemalloc.so.2

CMD ["imgproxy"]

EXPOSE 8080
