ARG base_image=ubuntu:jammy
ARG build_image=golang:latest

FROM ${base_image} AS base_image

FROM ${build_image} as build_image
ENV CGO_ENABLED 0
WORKDIR $GOPATH/src/github.com/logsquaredn/geocloud
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

FROM build_image AS build
ARG version=0.0.0
ARG revision=
RUN go build -ldflags "-X github.com/logsquaredn/geocloud.Version=${verision} -X github.com/logsquaredn/geocloud.Revision=${revision}" -o /assets/geocloud ./cmd/geocloud/

FROM base_image AS install
COPY bin/ /usr/local/bin/

FROM install AS containerd
ARG containerd=https://github.com/containerd/containerd/releases/download/v1.5.7/containerd-1.5.7-linux-amd64.tar.gz
ADD ${containerd} /tmp/
RUN if_tar_exists_xzf_rm_mv_assets /tmp/$(basename ${containerd})
RUN rm /assets/bin/ctr

FROM install AS pigz
ARG pigz=assets/pigz_2.4_x86_64.tgz
ADD ${pigz} /assets/

FROM install AS runc
ARG runc=https://github.com/opencontainers/runc/releases/download/v1.0.2/runc.amd64
ADD ${runc} /assets/runc
RUN chmod +x /assets/runc

FROM base_image AS geocloud
ENV PATH=/usr/local/geocloud/bin:$PATH
RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates
RUN apt-get remove -y ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*
RUN 
COPY --from=build /assets/ /usr/local/geocloud/bin/
COPY --from=containerd /assets/bin/ /usr/local/geocloud/bin/
COPY --from=pigz /assets/ /usr/local/geocloud/bin/
COPY --from=runc /assets/ /usr/local/geocloud/bin/
VOLUME /var/lib/containerd/
ENTRYPOINT ["geocloud"]
CMD ["--help"]
