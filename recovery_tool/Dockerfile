# Building this Dockerfile executes a full build of the Recovery Tool, cross-compiling inside 
# the container. The resulting executable is copied to the output directory using Docker BuildKit.

# You need to pass 3 parameters via --build-arg:
# 1. `os`  : the GOOS env var -- `linux`, `windows` or `darwin`.
# 2. `arch`: the GOARCH env var -- `386` or `amd64` (note that darwin/386 is not a thing).
# 3. `cc`  : the CC env var -- a C compiler for CGO to use, empty to use the default.
# 4. `out` : the name of the resulting executable, placed in the output directory on the host.

# For example, to build a linux/386 binary into `bin/rt`:
#   docker build . --output bin --build-arg os=linux --build-arg arch=386 --build-arg out=rt

# Note that the --output <dir> flag refers to the host, outside the container.

# --------------------------------------------------------------------------------------------------

# We have to different base images: one for arm and one for amd.
# Turns out that debian gcc packages assume the current is always the "gcc" package, but 
# for other archs it's "gcc-multilib-${arch}-linux-gnu". This makes it impossible to build
# one command that downloads all the archs we want for both amd64 and arm64 runners.
# The solution is to have one base layer per arch.

FROM golang:1.22.6-bookworm AS rtool-build-base-arm64

# Avoid prompts during package installation:
ENV DEBIAN_FRONTEND="noninteractive"

# Upgrade indices:
RUN apt-get update

# Install the various compilers we're going to use, with specific versions:
RUN apt-get install -y \
  gcc-mingw-w64 \
  gcc-12-multilib-i686-linux-gnu \
  gcc-12-multilib-x86-64-linux-gnu

FROM golang:1.22.6-bookworm AS rtool-build-base-amd64

# Avoid prompts during package installation:
ENV DEBIAN_FRONTEND="noninteractive"

# Upgrade indices:
RUN apt-get update

# Install the various compilers we're going to use, with specific versions:
RUN apt-get install -y \
  gcc-mingw-w64 \
  gcc-12-multilib-i686-linux-gnu \
  gcc-12-aarch64-linux-gnu

FROM rtool-build-base-${TARGETARCH} AS rtool-build-base

# Copy the source code into the container:
WORKDIR /src
COPY . .

RUN /bin/bash

# --------------------------------------------------------------------------------------------------

FROM rtool-build-base AS rtool-build
ARG os
ARG arch
ARG cc

# Enable and configure C support:
ENV CGO_ENABLED=1
ENV GO386=softfloat

# Do the thing:
RUN env GOOS=${os} GOARCH=${arch} CC=${cc} go build -mod=vendor -a -trimpath -o /out ./recovery_tool/

# --------------------------------------------------------------------------------------------------

FROM scratch
ARG out

# Copy the resulting executable back to the host:
COPY --from=rtool-build /out ${out}
