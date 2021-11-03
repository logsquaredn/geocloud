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
int vectorInitialize(struct GDALHandles *gdalHandles, const char *inputFilePath, const char *outputDir);
int buildOutputVectorFeature(struct GDALHandles *gdalHandles, OGRFeatureH *inputFeature, OGRGeometryH *geometry);
// inputGeoFilePath needs free()'d
char *getInputGeoFilePath(const char *inputFilePath);
int dumpToGeojson(const char *outputDir);
int zipShp(const char *outputDir);
int cleanup(const char *outputDir);

#endif
