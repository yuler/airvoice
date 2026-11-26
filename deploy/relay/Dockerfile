# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download github.com/gorilla/websocket

COPY cli/e2e ./cli/e2e
COPY relay ./relay
COPY cmd/airvoice-relay ./cmd/airvoice-relay

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/airvoice-relay \
    ./cmd/airvoice-relay

FROM alpine:3.21

RUN adduser -S -D -H airvoice

COPY --from=builder /out/airvoice-relay /usr/local/bin/airvoice-relay

USER airvoice
EXPOSE 8787

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8787/health >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/airvoice-relay"]
