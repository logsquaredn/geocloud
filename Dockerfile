ARG base_image=osgeo/gdal:alpine-normal-3.4.3
ARG build_image=golang:1.19-alpine3.15
ARG build_tasks_image=osgeo/gdal:alpine-normal-3.4.3

FROM ${base_image} AS base_image

FROM ${build_image} as build_image
ENV CGO_ENABLED 0
WORKDIR $GOPATH/src/github.com/logsquaredn/rototiller
RUN apk add --no-cache git

FROM ${build_tasks_image} AS build_tasks_image
WORKDIR /src/github.com/logsquaredn/rototiller/tasks
RUN apk add --no-cache gcc libc-dev

FROM build_tasks_image AS build_tasks
RUN mkdir -p /assets
COPY tasks/ .
RUN gcc -Wall removebadgeometry/removebadgeometry.c shared/shared.c -l gdal -o /assets/removebadgeometry
RUN gcc -Wall buffer/buffer.c shared/shared.c -l gdal -o /assets/buffer
RUN gcc -Wall filter/filter.c shared/shared.c -l gdal -o /assets/filter
RUN gcc -Wall reproject/reproject.c shared/shared.c -l gdal -o /assets/reproject
RUN gcc -Wall lookup/vectorlookup.c shared/shared.c -l gdal -o /assets/vectorlookup
RUN gcc -Wall lookup/rasterlookup.c shared/shared.c -l gdal -o /assets/rasterlookup
RUN gcc -Wall lookup/polygonVectorLookup.c shared/shared.c -l gdal -o /assets/polygonVectorLookup

FROM build_image AS build
ARG semver=0.0.0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-s -w -X github.com/logsquaredn/rototiller.Semver=${semver}" -o /assets/rotoctl ./cmd/rotoctl
RUN go build -ldflags "-s -w -X github.com/logsquaredn/rototiller.Semver=${semver}" -o /assets/rotoproxy ./cmd/rotoproxy
RUN go build -ldflags "-s -w -X github.com/logsquaredn/rototiller.Semver=${semver}" -o /assets/rototiller ./cmd/rototiller

FROM base_image AS rototiller
ARG zip=assets/zip_3.0_x86_64.tgz
ADD ${zip} /usr/local/bin
VOLUME /var/lib/rototiller
ENTRYPOINT ["rototiller"]
CMD ["--help"]
COPY --from=build_tasks /assets /usr/local/bin
COPY --from=build /assets /usr/local/bin
