FROM golang:1.17 as build
LABEL description="Build container"

ENV CGO_ENABLED 0
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build github.com/paralus/paralus

FROM alpine:latest as runtime
LABEL description="Run container"

COPY --from=build /build/paralus /usr/bin/paralus
WORKDIR /usr/bin
# Copying data for running migrations
# TODO: Support rcloud-base binary to run migrations
COPY ./persistence/migrations/admindb /data/migrations/admindb

EXPOSE 10000
EXPOSE 11000
