#ifndef SHARED_H
#define SHARED_H

#include "gdal.h"
#include <ogr_srs_api.h>

int openVectorDataset(GDALDatasetH *dataset, const char *filePath);
const char *getDriverName(const char *filePath);
int getDriver(GDALDriverH **driver, const char *driverName);
int deleteExistingDataset(GDALDriverH driver, const char* filePath);
int createVectorDataset(GDALDatasetH *dataset, GDALDriverH driver, const char *filePath);

#endif
