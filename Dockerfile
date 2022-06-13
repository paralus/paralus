FROM golang:1.17 as build
LABEL description="Build container"

ENV CGO_ENABLED 0
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

FROM scratch as runtime
LABEL description="Run container"

COPY --from=build /build/paralus /usr/bin/paralus
WORKDIR /usr/bin
# Copying data for running migrations
# TODO: Support paralus binary to run migrations
COPY ./persistence/migrations/admindb /data/migrations/admindb

# RPC port
EXPOSE 10000
# RPC relay peering port
EXPOSE 10001
# HTTP port
EXPOSE 11000
