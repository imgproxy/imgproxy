FROM debian:bullseye-slim

RUN apt-get -qq update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    build-essential \
    curl \
    git \
    libglib2.0-dev \
    libxml2-dev \
    libjpeg-dev \
    libpng-dev \
    libwebp-dev \
    librsvg2-dev \
    libexif-dev \
    liblcms2-dev \
    libavcodec-dev \
    libavformat-dev \
    libavutil-dev \
    libswscale-dev \
    libopencv-core-dev \
    libopencv-imgproc-dev \
    libopencv-dnn-dev

RUN \
  mkdir /root/vips \
  && cd /root/vips \
  && curl -s -S -L -o vips_releases.json "https://api.github.com/repos/libvips/libvips/releases" \
  && for VIPS_VERSION in "8.10" "8.11" "8.12"; do \
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

RUN echo "Name: OpenCV\n" \
  "Description: Open Source Computer Vision Library\n" \
  "Version: 4.5.1\n" \
  "Libs: -L/usr/lib/x86_64-linux-gnu -lopencv_dnn -lopencv_imgproc -lopencv_core\n" \
  "Libs.private: -ldl -lm -lpthread -lrt\n" \
  "Cflags: -I/usr/include/opencv4\n" \
  > /usr/lib/x86_64-linux-gnu/pkgconfig/opencv4.pc

WORKDIR /go/src

ENTRYPOINT [ "/bin/bash" ]
