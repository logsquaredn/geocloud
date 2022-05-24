#ifndef SHARED_H
#define SHARED_H

#include "gdal.h"
#include <ogr_srs_api.h>
#include <stdlib.h>
#include <libgen.h>
#include <dirent.h>

extern const char *ENV_VAR_INPUT_FILEPATH;
extern const char *ENV_VAR_OUTPUT_DIRECTORY;
extern int MAX_UNZIPPED_FILES;
extern int ONE_KB;

struct GDALHandles {
    GDALDatasetH *inputDataset;
    GDALDatasetH *outputDataset;
    OGRLayerH *inputLayer;
    OGRSpatialReferenceH *inputSpatialRef;
    OGRLayerH *outputLayer;
    OGRFeatureDefnH *outputFeatureDefn;
};

void info(const char *msg);
void error(const char *message, const char *file, int line);
void fatalError(const char *message, const char *file, int line);

int isZip(const char *fp);
// result of unzip must be free()'d
char **unzip(const char *fp);

GDALDatasetH initRaster(const char *fp);


// inputGeoFilePath needs free()'d
char *getInputGeoFilePath(const char *inputFilePath);

int vectorInitialize(struct GDALHandles *gdalHandles, const char *inputFilePath, const char *outputDir);
int buildOutputVectorFeature(struct GDALHandles *gdalHandles, OGRFeatureH *inputFeature, OGRGeometryH *geometry);

int dumpToGeojson(const char *outputDir);
int zipShp(const char *outputDir);
int cleanup(const char *outputDir);
int splitGeometries(OGRGeometryH[], int, OGRGeometryH);

#endif
