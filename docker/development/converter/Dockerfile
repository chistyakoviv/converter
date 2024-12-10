FROM golang:1.23.2-alpine3.20

# Allow to run programs with -race flag
ENV CGO_ENABLED=1

# build-base allows to run programs with -race flag
RUN apk add --update --no-cache build-base entr git openssl

# Dependencies for VIPS
# aom-dev library (for AV1 support)
# libheif library (for HEIF/AVIF decoding)
RUN apk add --update --no-cache \
    vips \
    vips-dev \
    gcc \
    g++ \
    libc-dev \
    make \
    pkgconfig \
    glib-dev \
    libjpeg-turbo-dev \
    libpng-dev \
    libwebp-dev \
    libexif-dev \
    tiff-dev \
    libheif-dev \
    aom-dev

# Dependencies for FFmpeg
# Verify FFmpeg installation: ffmpeg -codecs | grep -E "libvpx|libopus"
# No need to install liboupus-dev and libvpx-dev, because the Alpine ffmpeg package already includes support for them
RUN apk add --update --no-cache ffmpeg

# Change working directory after installing dependencies
# (in case some deps are needed to be dowlaoded, unpacked and copied to the right place)
WORKDIR /go/src/github.com/chistyakoviv/converter

# Do not trigger reinstalling packages after updating dependencies
COPY ./go.mod ./go.sum ./

# Download dependencies required by go.mod
RUN go mod download

# Error: JSON arguments recommended for ENTRYPOINT/CMD to prevent unintended behavior related to OS signals
# Description:
# ENTRYPOINT and CMD instructions both support two different syntaxes for arguments:
# Shell form: CMD my-cmd start
# Exec form: CMD ["my-cmd", "start"]
# When you use shell form, the executable runs as a child process to a shell, which doesn't pass signals.
# This means that the program running in the container can't detect OS signals like SIGTERM and SIGKILL and respond to them correctly.
# see https://docs.docker.com/reference/build-checks/json-args-recommended/
CMD find . -name "*.go" | entr -r go run ./cmd/converter/main.go