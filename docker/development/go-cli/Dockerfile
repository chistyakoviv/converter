FROM golang:1.24.2-alpine3.21

# Do not use MOCKERY_VERSION, as it causes mockery to fail with a version parsing error.
ENV MOCKERY_INSTALL_VERSION=2.50.0

RUN apk add --no-cache \
    curl \
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
    ffmpeg \
    git

RUN mkdir -p /tmp/mockery && \
    curl -L -o /tmp/mockery/mockery.tar.gz "https://github.com/vektra/mockery/releases/download/v${MOCKERY_INSTALL_VERSION}/mockery_${MOCKERY_INSTALL_VERSION}_Linux_x86_64.tar.gz" && \
    tar -xzf /tmp/mockery/mockery.tar.gz -C /tmp/mockery && \
    mv /tmp/mockery/mockery /usr/local/bin/ && \
    chmod +x /usr/local/bin/mockery && \
    rm -rf /tmp/mockery

# See https://golangci-lint.run/welcome/install/
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6

WORKDIR /go/src/github.com/chistyakoviv/converter

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.23.0

COPY ./go.mod ./go.sum ./
RUN go mod tidy

COPY ./docker/common/wait-for-it.sh /usr/local/bin/wait-for-it
RUN chmod 555 /usr/local/bin/wait-for-it

RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app
RUN chown -R app:app /go/pkg

USER app