FROM golang:1.24.0 AS build
LABEL description="Build container"

ENV CGO_ENABLED 0
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

FROM alpine:latest AS runtime
LABEL description="Run container"

WORKDIR /usr/bin
COPY --from=build /build/paralus /usr/bin/paralus

# RPC port
EXPOSE 10000
# RPC relay peering port
EXPOSE 10001
# HTTP port
EXPOSE 11000

CMD [ "/usr/bin/paralus" ]
