FROM golang:1.23.2-alpine3.20

# Do not use MOCKERY_VERSION, as it causes mockery to fail with a version parsing error.
ENV MOCKERY_INSTALL_VERSION=2.50.0

RUN apk add --no-cache make curl

RUN apk add --no-cache \
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
    aom-dev \
    ffmpeg

RUN mkdir -p /tmp/mockery && \
    curl -L -o /tmp/mockery/mockery.tar.gz "https://github.com/vektra/mockery/releases/download/v${MOCKERY_INSTALL_VERSION}/mockery_${MOCKERY_INSTALL_VERSION}_Linux_x86_64.tar.gz" && \
    tar -xzf /tmp/mockery/mockery.tar.gz -C /tmp/mockery && \
    mv /tmp/mockery/mockery /usr/local/bin/ && \
    chmod +x /usr/local/bin/mockery && \
    rm -rf /tmp/mockery

WORKDIR /go/src/github.com/chistyakoviv/converter

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.23.0 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2

COPY ./docker/common/wait-for-it.sh /usr/local/bin/wait-for-it
RUN chmod 555 /usr/local/bin/wait-for-it

RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

USER app