ARG BASE_IMAGE_VERSION="v1.4.0"

FROM darthsim/imgproxy-base:${BASE_IMAGE_VERSION} as build
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN mkdir /imgproxy_build
COPY . /imgproxy_build
RUN cd /imgproxy_build; go mod tidy
RUN cd /imgproxy_build; go build -v -o /usr/local/bin/imgproxy

# ==================================================================================================
# Final image

FROM debian:bullseye-slim
LABEL maintainer="Sergey Alexandrovich <darthsim@gmail.com>"

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    libsm6 \
    liblzma5 \
    libzstd1 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/bin/imgproxy /usr/local/bin/
COPY --from=build /usr/local/lib /usr/local/lib

COPY NOTICE /usr/local/share/doc/imgproxy/

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV LD_LIBRARY_PATH /usr/local/lib

CMD ["imgproxy"]

EXPOSE 8080
