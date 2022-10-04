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

FROM ubuntu:22.04 AS rtool-build-base

# Avoid prompts during package installation:
ENV DEBIAN_FRONTEND="noninteractive"

# Upgrade indices:
RUN apt-get update

# Install the various compilers we're going to use, with specific versions:
RUN apt-get install -y \
  golang-1.18-go=1.18.1-1ubuntu1 \
  gcc-mingw-w64=10.3.0-14ubuntu1+24.3 \
  gcc-multilib=4:11.2.0-1ubuntu1

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
RUN env GOOS=${os} GOARCH=${arch} CC=${cc} /usr/lib/go-1.18/bin/go build -mod=vendor -a -trimpath -o /out .

# --------------------------------------------------------------------------------------------------

FROM scratch
ARG out

# Copy the resulting executable back to the host:
COPY --from=rtool-build /out ${out}
