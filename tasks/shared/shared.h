#ifndef SHARED_H
#define SHARED_H

#include "gdal.h"
#include <ogr_srs_api.h>
#include <stdlib.h>
#include <libgen.h>
#include <dirent.h>

struct GDALHandles {
    GDALDatasetH *inputDataset;
    GDALDatasetH *outputDataset;
    OGRLayerH *inputLayer;
    OGRSpatialReferenceH *inputSpatialRef;
    OGRLayerH *outputLayer;
    OGRFeatureDefnH *outputFeatureDefn;
};

void error(const char *message, const char *file, int line);
void fatalError();
int vectorInitialize(struct GDALHandles *gdalHandles, const char *inputFilePath, const char *outputFilePath);
int buildOutputVectorFeature(struct GDALHandles *gdalHandles, OGRFeatureH *inputFeature, OGRGeometryH *geometry);
// inputGeoFilePath needs free()'d
int getInputGeoFilePath(const char *inputFilePath, char **inputGeoFilePath);
int dumpToGeojson(struct GDALHandles *gdalHandles);

#endif
