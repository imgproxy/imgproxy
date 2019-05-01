FROM alpine:edge
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$PATH

ADD . /go/src/github.com/DarthSim/imgproxy
WORKDIR /go/src/github.com/DarthSim/imgproxy

# Install dependencies
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
  && apk --no-cache upgrade \
  && apk add --no-cache curl ca-certificates go gcc g++ make musl-dev fftw-dev glib-dev expat-dev \
    libjpeg-turbo-dev libpng-dev libwebp-dev giflib-dev librsvg-dev libexif-dev lcms2-dev libimagequant-dev

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
    --without-jbig \
    --without-jpeg \
    --without-lcms \
    --without-lzma \
    --without-png \
    --without-tiff \
    --without-wmf \
    --without-xml \
    --without-webp \
  && make install-strip

# Build libvips
RUN cd /root \
  && export VIPS_VERSION=$(curl -s "https://api.github.com/repos/libvips/libvips/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/') \
  && echo "Vips version: $VIPS_VERSION" \
  && curl -Ls https://github.com/libvips/libvips/releases/download/v$VIPS_VERSION/vips-$VIPS_VERSION.tar.gz | tar -xz \
  && cd vips-$VIPS_VERSION \
  && ./configure \
    --disable-magickload \
    --without-python \
    --without-tiff \
    --without-OpenEXR \
    --enable-debug=no \
    --disable-static \
    --enable-silent-rules \
  && make install-strip

# Build imgproxy
RUN cd /go/src/github.com/DarthSim/imgproxy \
  && CGO_LDFLAGS_ALLOW="-s|-w" go build -v -o /usr/local/bin/imgproxy

# Copy compiled libs here to copy them to the final image
RUN cd /root \
  && mkdir libs \
  && ldd /usr/local/bin/imgproxy | grep /usr/local/lib/ | awk '{print $3}' | xargs -I '{}' cp '{}' libs/

# ==================================================================================================
# Final image

FROM alpine:edge
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
  && apk --no-cache upgrade \
  && apk add --no-cache bash ca-certificates fftw glib expat libjpeg-turbo libpng \
    libwebp giflib librsvg libgsf libexif lcms2 libimagequant\
  && apk add --no-cache \
      fontconfig \
      font-bh-100dpi \
      font-sun-misc \
      font-bh-lucidatypewriter-100dpi \
      font-adobe-utopia-type1 \
      font-cronyx-cyrillic \
      font-misc-cyrillic \
      font-schumacher-misc \
      font-daewoo-misc \
      font-screen-cyrillic \
      font-adobe-utopia-75dpi \
      font-bitstream-100dpi \
      font-xfree86-type1 \
      font-bitstream-75dpi \
      font-bh-ttf \
      font-arabic-misc \
      font-dec-misc \
      font-misc-ethiopic \
      font-micro-misc \
      font-alias \
      font-isas-misc \
      font-bh-lucidatypewriter-75dpi \
      font-winitzki-cyrillic \
      font-jis-misc \
      ttf-ubuntu-font-family \
      font-bitstream-type1 \
      font-mutt-misc \
      font-misc-misc \
      font-adobe-100dpi \
      font-bh-type1 \
      font-bh-75dpi \
      font-sony-misc \
      font-ibm-type1 \
      font-bitstream-speedo \
      font-adobe-utopia-100dpi \
      font-adobe-75dpi \
      font-misc-meltho \
      font-cursor-misc \
  && rm -rf /var/cache/apk*

COPY --from=0 /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=0 /root/libs/* /usr/local/lib/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=4

CMD ["imgproxy"]

EXPOSE 8080
