#!/bin/bash

gcc buffer/buffer.c shared/shared.c -l gdal -o bin/buffer
gcc filter/filter.c shared/shared.c -l gdal -o bin/filter
gcc reproject/reproject.c shared/shared.c -l gdal -o bin/reproject