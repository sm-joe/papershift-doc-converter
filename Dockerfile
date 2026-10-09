# syntax=docker/dockerfile:1.7

FROM golang:1.27-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/papershift-api \
    ./apps/api

FROM debian:trixie-slim

ENV PAPERSHIFT_ADDR=:8080 \
    PAPERSHIFT_WORK_DIR=/tmp/papershift

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends \
        libreoffice \
        libpcre2-8-0 \
        ca-certificates \
        curl \
        fonts-dejavu \
        default-jre \
        libreoffice-java-common \
        imagemagick \
        librsvg2-bin \
        pandoc \
        poppler-utils \
        ghostscript \
    && rm -rf /var/lib/apt/lists/* \
    && useradd \
        --system \
        --uid 10001 \
        --create-home \
        --home-dir /home/papershift \
        papershift \
    && mkdir -p /tmp/papershift \
    && chown -R papershift:papershift /tmp/papershift \
    && chmod 700 /tmp/papershift

COPY --from=builder /out/papershift-api /usr/local/bin/papershift-api

RUN chmod 0755 /usr/local/bin/papershift-api

USER 10001:10001

WORKDIR /home/papershift

EXPOSE 8080

HEALTHCHECK --interval=30s \
    --timeout=5s \
    --start-period=30s \
    --retries=3 \
    CMD curl --fail --silent http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/papershift-api"]
