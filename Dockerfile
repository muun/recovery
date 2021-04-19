# Building this Dockerfile executes a full build of the Recovery Tool, cross-compiling inside 
# the container. The resulting executable is copied to the output directory using Docker BuildKit.

# You need to pass 3 parameters via --build-arg:
# 1. `os`  : the GOOS env var -- `linux`, `windows` or `darwin`.
# 2. `arch`: the GOARCH env var -- `386` or `amd64` (note that darwin/386 is not a thing).
# 3. `out` : the name of the resulting executable, placed in the output directory on the host.

# For example, to build a linux/386 binary into `bin/rt`:
#   docker build . --output bin --build-arg os=linux --build-arg arch=386 --build-arg out=rt

# Note that the --output <dir> flag refers to the host, outside the container.

FROM golang:1.16.0-alpine3.13 AS build
ARG os
ARG arch

RUN apk add --no-cache build-base=0.5-r2

WORKDIR /src
COPY . .

ENV CGO_ENABLED=0
RUN env GOOS=${os} GOARCH=${arch} go build -mod=vendor -a -trimpath -o /out .

# ---

FROM scratch
ARG out

COPY --from=build /out ${out}