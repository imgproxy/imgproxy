FROM golang:1-alpine3.10
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$PATH

# Install dependencies
RUN apk --no-cache upgrade \
  && apk add --no-cache \
    curl \
    git \
    ca-certificates \
    gcc \
    g++ \
    make \
    musl-dev \
    fftw-dev \
    glib-dev \
    libtool \
    expat-dev \
    libjpeg-turbo-dev \
    libpng-dev \
    libwebp-dev \
    librsvg-dev \
    libexif-dev \
    lcms2-dev \
    libheif-dev \
    tiff-dev \
  && apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/main \
    giflib-dev \
  && apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing \
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
  && make install-strip

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
    --enable-silent-rules \
  && make install-strip

ADD . /app
WORKDIR /app

# Build imgproxy
RUN cd /app \
  && CGO_LDFLAGS_ALLOW="-s|-w" go build -v -o /usr/local/bin/imgproxy

# Copy compiled libs here to copy them to the final image
RUN cd /root \
  && mkdir libs \
  && ldd /usr/local/bin/imgproxy | grep /usr/local/lib/ | awk '{print $3}' | xargs -I '{}' cp '{}' libs/

# ==================================================================================================
# Final image

FROM alpine:3.10
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN apk --no-cache upgrade \
  && apk add --no-cache \
    bash \
    ca-certificates \
    fftw \
    glib \
    libltdl \
    expat \
    libjpeg-turbo \
    libpng \
    libwebp \
    librsvg \
    libgsf \
    libexif \
    lcms2 \
    libheif \
    tiff \
  && apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/main \
    giflib \
  && apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing \
    libimagequant \
  && rm -rf /var/cache/apk*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=0 /root/libs/* /usr/local/lib/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=4

CMD ["imgproxy"]

EXPOSE 8080
