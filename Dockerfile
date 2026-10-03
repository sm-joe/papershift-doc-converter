# syntax=docker/dockerfile:1.7

FROM golang:1.27-bookworm AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/papershift-api \
    ./apps/api

FROM debian:bookworm-slim

ENV PAPERSHIFT_ADDR=:8080 \
    PAPERSHIFT_WORK_DIR=/tmp/papershift

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        libreoffice \
        ca-certificates \
        fonts-dejavu \
        default-jre libreoffice-java-common \
        imagemagick librsvg2-bin \
    && rm -rf /var/lib/apt/lists/* \
    && useradd \
        --system \
        --uid 10001 \
        --create-home \
        --home-dir /home/papershift \
        papershift \
    && mkdir -p /tmp/papershift \
    && chown -R papershift:papershift /tmp/papershift

COPY --from=builder /out/papershift-api /usr/local/bin/papershift-api

USER 10001:10001

WORKDIR /home/papershift

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/papershift-api"]