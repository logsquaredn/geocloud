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
COPY api/ api/
COPY cmd/ cmd/
COPY shared/ shared/
COPY tools/ tools/
COPY worker/ worker/
COPY *.go .

FROM build_image AS build
ARG ldflags
RUN go build -ldflags "${ldflags}" -o /assets/api ./api/cmd/ \
    && go build -ldflags "${ldflags}" -o /assets/worker ./worker/cmd/ \
    && go build -ldflags "${ldflags}" -o /assets/geocloud ./cmd/

FROM base_image AS api
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        dumb-init \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /assets/api /usr/local/geocloud/bin/geocloud
ENV PATH=/usr/local/geocloud/bin:$PATH
ENTRYPOINT ["dumb-init", "geocloud"]

FROM base_image AS containerd
ARG containerd_release=https://github.com/containerd/containerd/releases/download/v1.5.2/containerd-1.5.2-linux-amd64.tar.gz
# when its dest is a remote .tgz, ADD does not unpack the tarball
# when its dest is a local .tgz, the tarball is unpacked
ADD ${containerd_release} /tmp/
# this conditional handles that difference in ADD's functionality between remote and local
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

FROM api AS worker
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        pigz \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /assets/worker /usr/local/geocloud/bin/geocloud
COPY --from=containerd /assets/ /usr/local/geocloud/bin/
COPY --from=runc /assets/ /usr/local/geocloud/bin/
VOLUME /var/lib/geocloud/containerd
ENV GEOCLOUD_CONTAINERD_ROOT /var/lib/geocloud/containerd
ENV GEOCLOUD_CONTAINERD_PROMETHEUS_IP 0.0.0.0
ENV GEOCLOUD_WORKER_IP 0.0.0.0

FROM worker AS geocloud
COPY --from=build /assets/geocloud /usr/local/geocloud/bin/geocloud

FROM geocloud AS final
