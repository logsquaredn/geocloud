ARG base_image=ubuntu:bionic
ARG build_image=osgeo/gdal:latest

FROM ${base_image} AS base
ENV DEBIAN_FRONTEND=noninteractive

FROM ${build_image} AS build
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        gcc \
    && rm -rf /var/lib/apt/lists/*
