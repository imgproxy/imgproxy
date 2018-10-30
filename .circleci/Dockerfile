FROM debian:stretch

RUN apt-get -qq update \
  && apt-get install -y --no-install-recommends bash unzip ca-certificates build-essential \
  curl git mercurial make binutils bison gcc autoconf libtool automake swig \
  gobject-introspection libglib2.0-dev libexpat1-dev libxml2-dev libfftw3-dev \
  libjpeg-dev libpng-dev libwebp-dev libgif-dev libexif-dev liblcms2-dev gtk-doc-tools

RUN curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | bash -

RUN \
  for VIPS_VERSION in "8.3" "8.4" "8.5" "8.6" "8.7"; do \
    mkdir -p /root/vips \
    && cd /root/vips \
    && mkdir /root/vips/$VIPS_VERSION \
    && curl -s -S -L -o $VIPS_VERSION.zip https://github.com/libvips/libvips/archive/$VIPS_VERSION.zip \
    && unzip $VIPS_VERSION.zip \
    && cd libvips-$VIPS_VERSION \
    && test -f autogen.sh && ./autogen.sh || ./bootstrap.sh \
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
    && make \
    && make install \
    && cd .. \
    && rm -rf $VIPS_VERSION.zip libvips-$VIPS_VERSION; \
  done

WORKDIR /go/src
ENV GOPATH=/go

ENTRYPOINT [ "/bin/bash" ]
