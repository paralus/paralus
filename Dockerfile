FROM golang:1.17 as build
LABEL description="Build container"

ENV CGO_ENABLED 0
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

# pinning to 3.14.10 which does not have any vulnerabilities
# track https://hub.docker.com/_/alpine/tags for vulnerability fixes in latest version and move back to using latest
FROM alpine:3.14.10 as runtime
LABEL description="Run container"

WORKDIR /usr/bin
COPY --from=build /build/paralus /usr/bin/paralus

# RPC port
EXPOSE 10000
# RPC relay peering port
EXPOSE 10001
# HTTP port
EXPOSE 11000
