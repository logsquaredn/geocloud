ARG base_image=ubuntu:bionic
ARG build_image=golang:latest

FROM ${base_image} AS base_image
ENV DEBIAN_FRONTEND=noninteractive

FROM ${build_image} as build_image
ENV CGO_ENABLED 0
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM base_image AS containerd
ARG containerd_release=https://github.com/containerd/containerd/releases/download/v1.5.2/containerd-1.5.2-linux-amd64.tar.gz
# when its dest is a remote .tgz, ADD does not unpack the tarball
# when its dest is a local .tgz, the tarvall is unpacked
ADD ${containerd_release} /tmp/
# this conditional handles that difference in ADD's functionality
RUN TGZ=/tmp/$(basename ${containerd_release}); \
    if [ -f $TGZ ]; then \
        tar \
            -C /tmp/ \
            -xzf \
            $TGZ \
        && rm $TGZ; \
    fi; \
    mkdir /assets/ \
        && mv /tmp/bin/* /assets/

FROM base_image AS runc
ARG runc_release=https://github.com/opencontainers/runc/releases/download/v1.0.0-rc95/runc.amd64
ADD ${runc_release} /assets/runc
RUN chmod +x /assets/runc

FROM build_image AS build
COPY api/ api/
COPY cmd/ cmd/
COPY runners/ runners/
COPY tasks/mock/ tasks/mock/
COPY worker/ worker/
COPY *.go .
ARG ldflags
RUN go build -ldflags "$ldflags" -o /assets/geocloud ./cmd/

FROM base_image AS geocloud
COPY --from=build /assets/ /usr/local/geocloud/bin/
COPY --from=containerd /assets/ /usr/local/geocloud/bin/
COPY --from=runc /assets/ /usr/local/geocloud/bin/
ENV PATH=/usr/local/geocloud/bin:$PATH
VOLUME /var/lib/geocloud/containerd
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        dumb-init \
        pigz \
    && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["dumb-init", "geocloud"]
CMD ["--help"]
