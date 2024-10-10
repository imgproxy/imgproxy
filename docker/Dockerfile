ARG BASE_IMAGE_VERSION="v3.11.0"

FROM darthsim/imgproxy-base:${BASE_IMAGE_VERSION} as build

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
    libsm6 \
    liblzma5 \
    libzstd1 \
    libpcre3 \
    fontconfig-config \
    media-types \
    libjemalloc2 \
    libtcmalloc-minimal4 \
  && ln -s /usr/lib/$(uname -m)-linux-gnu/libjemalloc.so.2 /usr/local/lib/libjemalloc.so \
  && ln -s /usr/lib/$(uname -m)-linux-gnu/libtcmalloc_minimal.so.4 /usr/local/lib/libtcmalloc_minimal.so \
  && rm -rf /var/lib/apt/lists/*

# ====== PUSHD DITHER ========
# This is a temporary solution to install pushd-dither
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      python3 \
      python3-pip \
      python3-dev \
      gfortran \
      libopencv-dev \
      curl \
      zstd \
      build-essential \
      cmake

RUN pip3 install scikit-image==0.22.0 scipy==1.11.4 opencv-python numpy Pillow dlib==19.24.4 pyamg --break-system-packages

COPY plugins/pushd-dither /opt/pushd-dither/

RUN cd /opt/pushd-dither \
    && python3 -m pip install -e . --break-system-packages --target=/usr/local/lib/python3.11/

# decompress LUTs
RUN cd /opt/pushd-dither/lut_dither \
    && zstd -fd --rm lab_13_hack.interpol_clarabel.npy.zst \
    && zstd -fd --rm rgb_13.interpol_clarabel.npy.zst \
    && zstd -fd --rm hpminde_rgb.npy.zst

# provision vendored tools
RUN /opt/pushd-dither/vendor/provision-vendored.sh

# ====== END PUSHD DITHER ========

COPY --from=build /opt/imgproxy/ /opt/imgproxy/
COPY --from=build /usr/local/bin/imgproxy /usr/local/bin/imgproxy
COPY docker/entrypoint.sh /usr/local/bin/

COPY NOTICE /usr/local/share/doc/imgproxy/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV FONTCONFIG_PATH /etc/fonts
ENV IMGPROXY_MALLOC malloc

RUN groupadd -r imgproxy \
  && useradd -r -u 999 -g imgproxy imgproxy \
  && mkdir -p /var/cache/fontconfig \
  && chmod 777 /var/cache/fontconfig

# Disable SVE on ARM64. SVE is slower than NEON on Amazon Graviton 3
ENV VIPS_VECTOR=167772160

# run as root to allow vendored tools to run
#USER 999

ENTRYPOINT [ "entrypoint.sh" ]
CMD ["imgproxy"]

EXPOSE 8080
