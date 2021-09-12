ARG base_image=osgeo/gdal:latest
ARG build_image=osgeo/gdal:latest

# does base_image need libc6-dev?
FROM ${base_image} AS base_image
ENV DEBIAN_FRONTEND noninteractive

FROM ${build_image} AS build_image
WORKDIR /src/
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        gcc \
        libc6-dev \
    && rm -rf /var/lib/apt/lists/*

FROM build_image AS build
# the name of the task to build
#  badGeometry, buffer, etc.
ARG task
ARG task_folder=${task}
COPY ${task_folder}/ ${task_folder}/
COPY shared/ shared/
RUN mkdir /assets \
    && gcc ${task_folder}/${task}.c shared/shared.c -l gdal -o /assets/${task}

FROM base_image AS task
COPY --from=build /assets/ /usr/local/bin/task
ENTRYPOINT ["task"]
