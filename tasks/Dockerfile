ARG base_image=osgeo/gdal:ubuntu-small-latest
ARG build_image=osgeo/gdal:latest

FROM ${base_image} AS base_image

FROM base_image AS install

FROM install AS zip
ARG zip=assets/zip_3.0_x86_64.tgz
ADD ${zip} /assets/

FROM ${build_image} AS build_image
WORKDIR /src/github.com/logsquaredn/geocloud/tasks
RUN apt-get update
RUN apt-get install -y --no-install-recommends \
        gcc \
        libc6-dev

FROM build_image AS build
COPY . .
RUN mkdir /assets/
RUN gcc -Wall badGeometry/removeBadGeometry.c shared/shared.c -l gdal -o /assets/removebadgeometry
RUN gcc -Wall buffer/buffer.c shared/shared.c -l gdal -o /assets/buffer
RUN gcc -Wall filter/filter.c shared/shared.c -l gdal -o /assets/filter
RUN gcc -Wall reproject/reproject.c shared/shared.c -l gdal -o /assets/reproject

FROM base_image AS removebadgeometry
COPY --from=build /assets/removebadgeometry /usr/local/bin/
COPY --from=zip /assets/ /usr/local/bin/
ENTRYPOINT ["removebadgeometry"]

FROM base_image AS buffer
COPY --from=build /assets/buffer /usr/local/bin/
COPY --from=zip /assets/ /usr/local/bin/
ENTRYPOINT ["buffer"]

FROM base_image AS filter
COPY --from=build /assets/filter /usr/local/bin/
COPY --from=zip /assets/ /usr/local/bin/
ENTRYPOINT ["filter"]

FROM base_image AS reproject
COPY --from=build /assets/reproject /usr/local/bin/
COPY --from=zip /assets/ /usr/local/bin/
ENTRYPOINT ["reproject"]
