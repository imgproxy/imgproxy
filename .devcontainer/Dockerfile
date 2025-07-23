ARG BASE_IMAGE=ghcr.io/imgproxy/imgproxy-base:latest

# Git build stage
FROM ${BASE_IMAGE} AS build

ARG LYCHEE_VERSION="0.19.1"
ARG GIT_VERSION="2.49.1"

# lsb-release and gnupg are needed for the LLVM repository
# TCL and gettext are needed for git
RUN apt-get install -y lsb-release tcl gettext autoconf libz-dev libcurl4-openssl-dev

# Install updated git
RUN curl -O https://www.kernel.org/pub/software/scm/git/git-${GIT_VERSION}.tar.gz && \
  tar -zxf git-${GIT_VERSION}.tar.gz && \
  cd git-${GIT_VERSION} && \
  make configure && \
  ./configure prefix=/usr/git && \
  make all && \
  make install && \
  rm -rf git-${GIT_VERSION}

# Get lychee binary
RUN curl -L "https://github.com/lycheeverse/lychee/releases/download/lychee-v${LYCHEE_VERSION}/lychee-$(uname -m)-unknown-linux-gnu.tar.gz" | \
  tar -zxf -

# Development container
FROM ${BASE_IMAGE}

ENV GOPATH=$HOME/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# gnupg is requiered for clang-format
RUN apt-get install -y gnupg lsb-release ssh

# Install air
RUN go install github.com/air-verse/air@latest

# Install lefthook
RUN go install github.com/evilmartians/lefthook@latest

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6

# Copy git from the git build stage
COPY --from=build /usr/git /usr

# Add the LLVM GPG key
RUN curl https://apt.llvm.org/llvm-snapshot.gpg.key | gpg --dearmor -o /usr/share/keyrings/llvm.gpg

# Add the LLVM 20 APT repo (replace jammy with your distro if needed)
RUN echo "deb [signed-by=/usr/share/keyrings/llvm.gpg] https://apt.llvm.org/$(lsb_release -cs)/ llvm-toolchain-$(lsb_release -cs)-20 main" \
  | tee /etc/apt/sources.list.d/llvm.list

# Install clang-format-20
RUN apt-get -y update && apt-get -y install clang-format-20

# Use clang-format-20 as the default clang-format
RUN ln -sf /usr/bin/clang-format-20 /usr/local/bin/clang-format

# Copy the lychee binary from the git build stage
COPY --from=build /app/lychee /usr/bin
