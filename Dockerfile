ARG base_image=ubuntu:bionic
ARG build_image=golang:latest

FROM ${base_image} AS base
ENV DEBIAN_FRONTEND=noninteractive

FROM ${build_image} as build
ENV CGO_ENABLED 0
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM build AS builder
COPY api/ api/
COPY cmd/ cmd/
COPY tasks/mock/ tasks/mock/
COPY worker/ worker/
COPY *.go .
RUN go build -o /assets/geocloud ./cmd

FROM base AS geocloud
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        dumb-init \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /assets/ /usr/local/bin/
ENTRYPOINT ["geocloud"]
