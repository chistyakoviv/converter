FROM golang:1.23.2-alpine3.20

WORKDIR /app

RUN apk add --update --no-cache make bash

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY ./docker/common/wait-for-it.sh /usr/local/bin/wait-for-it
RUN chmod 555 /usr/local/bin/wait-for-it

RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

USER app