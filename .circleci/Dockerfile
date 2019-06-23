FROM debian:stretch

RUN apt-get -qq update \
  && apt-get install -y --no-install-recommends bash ca-certificates build-essential \
  curl git mercurial make binutils bison gcc gobject-introspection libglib2.0-dev \
  libexpat1-dev libxml2-dev libfftw3-dev libjpeg-dev libpng-dev libwebp-dev libgif-dev \
  libexif-dev liblcms2-dev libavcodec-dev libavformat-dev libavutil-dev libswscale-dev

RUN curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | bash -

RUN \
  mkdir /root/vips \
  && cd /root/vips \
  && curl -s -S -L -o vips_releases.json "https://api.github.com/repos/libvips/libvips/releases" \
  && for VIPS_VERSION in "8.6" "8.7" "8.8"; do \
    mkdir $VIPS_VERSION \
    && export VIPS_RELEASE=$(grep -m 1 "\"tag_name\": \"v$VIPS_VERSION." vips_releases.json | sed -E 's/.*"v([^"]+)".*/\1/') \
    && echo "Building Vips $VIPS_RELEASE as $VIPS_VERSION" \
    && curl -s -S -L -o $VIPS_RELEASE.tar.gz https://github.com/libvips/libvips/releases/download/v$VIPS_RELEASE/vips-$VIPS_RELEASE.tar.gz \
    && tar -xzf $VIPS_RELEASE.tar.gz \
    && cd vips-$VIPS_RELEASE \
    && ./configure \
      --prefix=/root/vips/$VIPS_VERSION \
      --without-python \
      --without-gsf \
      --without-orc \
      --disable-debug \
      --disable-dependency-tracking \
      --disable-static \
      --enable-silent-rules \
      --enable-gtk-doc-html=no \
      --enable-gtk-doc=no \
      --enable-pyvips8=no \
    && make install \
    && cd .. \
    && rm -rf $VIPS_RELEASE.tar.gz vips-$VIPS_RELEASE; \
  done

WORKDIR /go/src
ENV GOPATH=/go

ENTRYPOINT [ "/bin/bash" ]
