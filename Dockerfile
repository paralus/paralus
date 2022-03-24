FROM golang:1.17 as build
LABEL description="Build container"

ENV CGO_ENABLED 0
COPY . /build
WORKDIR /build
RUN go build github.com/RafayLabs/rcloud-base

FROM alpine:latest as runtime
LABEL description="Run container"

COPY --from=build /build/rcloud-base /usr/bin/rcloud-base
WORKDIR /usr/bin
CMD ./rcloud-base

EXPOSE 10000
EXPOSE 11000
