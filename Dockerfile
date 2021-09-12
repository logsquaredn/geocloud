ARG base_image=ubuntu:bionic
ARG build_image=golang:latest

FROM ${base_image} AS base_image

FROM ${build_image} as build_image
ENV CGO_ENABLED 0
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY api/ api/
COPY cmd/ cmd/
COPY infrastructure/ infrastructure/
COPY migrate/ migrate/
COPY shared/ shared/
COPY worker/ worker/
COPY *.go ./

FROM build_image AS build
ARG ldflags
RUN go build -ldflags "${ldflags}" -o /assets/api ./api/cmd/ \
    && go build -ldflags "${ldflags}" -o /assets/worker ./worker/cmd/ \
    && go build -ldflags "${ldflags}" -o /assets/geocloud ./cmd/

FROM base_image AS ca_certificates
ARG ca_certificates_release=assets/ca_certificates_x86_64.tgz
# when its src is a remote .tgz, ADD does not unpack the tarball
# when its src is a local .tgz, ADD unpacks the tarball
ADD ${ca_certificates_release} /tmp/
# this conditional handles that difference in ADD's functionality between remote and local
RUN TGZ=/tmp/$(basename ${ca_certificates_release}); \
    if [ -f $TGZ ]; then \
        tar \
            -C /tmp/ \
            -xzf \
            $TGZ \
        && rm $TGZ; \
    fi; \
    mkdir /assets/ \
    && mv /tmp/* /assets/

FROM base_image AS dumb_init
ARG dumb_init_release=https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64
ADD ${dumb_init_release} /assets/dumb-init
RUN chmod +x /assets/dumb-init

FROM base_image AS api
COPY --from=build /assets/api /usr/local/geocloud/bin/geocloud
# COPY --from=ca_certificates /assets/ /usr/local/geocloud/bin/
COPY --from=dumb_init /assets/ /usr/local/geocloud/bin/
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
ENV PATH=/usr/local/geocloud/bin:$PATH
ENTRYPOINT ["dumb-init", "geocloud"]

FROM base_image AS containerd
ARG containerd_release=https://github.com/containerd/containerd/releases/download/v1.5.5/containerd-1.5.5-linux-amd64.tar.gz
ADD ${containerd_release} /tmp/
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

FROM base_image AS pigz
ARG pigz_release=assets/pigz_2.4_x86_64.tgz
ADD ${pigz_release} /tmp/
RUN TGZ=/tmp/$(basename ${pigz_release}); \
    if [ -f $TGZ ]; then \
        tar \
            -C /tmp/ \
            -xzf \
            $TGZ \
        && rm $TGZ; \
    fi; \
    mkdir /assets/ \
    && mv /tmp/* /assets/

FROM base_image AS runc
ARG runc_release=https://github.com/opencontainers/runc/releases/download/v1.0.1/runc.amd64
ADD ${runc_release} /assets/runc
RUN chmod +x /assets/runc

FROM api AS worker
COPY --from=build /assets/worker /usr/local/geocloud/bin/geocloud
COPY --from=containerd /assets/ /usr/local/geocloud/bin/
COPY --from=pigz /assets/ /usr/local/geocloud/bin/
COPY --from=runc /assets/ /usr/local/geocloud/bin/
VOLUME /var/lib/geocloud/containerd/
ENV GEOCLOUD_CONTAINERD_ROOT /var/lib/geocloud/containerd/
ENV GEOCLOUD_CONTAINERD_PROMETHEUS_IP 0.0.0.0
ENV GEOCLOUD_WORKER_IP 0.0.0.0

FROM worker AS geocloud
COPY --from=build /assets/geocloud /usr/local/geocloud/bin/geocloud

FROM geocloud AS final
