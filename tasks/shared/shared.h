#ifndef SHARED_H
#define SHARED_H

#include "gdal.h"
#include <ogr_srs_api.h>

struct GDALHandles {
    GDALDatasetH *inputDataset;
    GDALDatasetH *outputDataset;
    OGRLayerH *inputLayer;
    OGRSpatialReferenceH *inputSpatialRef;
    OGRLayerH *outputLayer;
    OGRFeatureDefnH *outputFeatureDefn;
};

void error(const char *message);
void fatalError();
int vectorInitialize(struct GDALHandles *gdalHandles, const char *inputFilePath, const char *outputFilePath);
int buildOutputVectorFeature(struct GDALHandles *gdalHandles, OGRFeatureH *inputFeature, OGRGeometryH *geometry);
const char *getInputGeoFilePath(const char *inputFilePath);

#endif
