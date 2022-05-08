ARG base_image=osgeo/gdal:alpine-normal-3.4.3
ARG build_image=golang:1.18-alpine3.15
ARG build_tasks_image=osgeo/gdal:alpine-normal-3.4.3

FROM ${base_image} AS base_image

FROM base_image AS install

FROM install AS zip
ARG zip=assets/zip_3.0_x86_64.tgz
ADD ${zip} /assets

FROM ${build_image} as build_image
ENV CGO_ENABLED 0
WORKDIR $GOPATH/src/github.com/logsquaredn/geocloud
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

FROM ${build_tasks_image} AS build_tasks_image
WORKDIR /src/github.com/logsquaredn/geocloud/tasks
RUN apk add --no-cache gcc libc-dev
COPY tasks/ .

FROM build_tasks_image AS build_tasks
RUN mkdir -p /assets/removebadgeometry /assets/buffer /assets/filter /assets/reproject
RUN gcc -Wall removebadgeometry/removebadgeometry.c shared/shared.c -l gdal -o /assets/removebadgeometry/task
RUN gcc -Wall buffer/buffer.c shared/shared.c -l gdal -o /assets/buffer/task
RUN gcc -Wall filter/filter.c shared/shared.c -l gdal -o /assets/filter/task
RUN gcc -Wall reproject/reproject.c shared/shared.c -l gdal -o /assets/reproject/task

FROM build_image AS build
ARG version=0.0.0
ARG prerelease=
ARG build=
RUN go build -ldflags "-s -w -X github.com/logsquaredn/geocloud.Version=${verision} -X github.com/logsquaredn/geocloud.Prerelease=${prerelease} -X github.com/logsquaredn/geocloud.Build=${build}" -o /assets/geocloud ./cmd/geocloud/

FROM base_image AS geocloud
ENV PATH=/usr/local/geocloud/bin:$PATH
RUN apk add --no-cache ca-certificates
RUN apk del ca-certificates
VOLUME /var/lib/geocloud
ENTRYPOINT ["geocloud"]
CMD ["--help"]
COPY --from=zip /assets /usr/local/geocloud/bin
COPY --from=build_tasks /assets /usr/local/geocloud/bin
COPY --from=build /assets /usr/local/geocloud/bin
