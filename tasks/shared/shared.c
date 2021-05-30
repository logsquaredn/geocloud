#include "shared.h"

int openVectorDataset(GDALDatasetH *dataset, const char *filePath) {
    *dataset = GDALOpenEx(filePath, GDAL_OF_VECTOR, NULL, NULL, NULL);
	if(*dataset == NULL) {
		return 1;
	}
	
    return 0;
}

const char *getDriverName(const char *filePath) {
    char *driverName;
	const char *ext = strrchr(filePath, '.');
	if(ext != NULL && !strcmp(ext, ".shp")) {
		driverName = "ESRI Shapefile";
	} else {
		driverName = "GeoJSON";
	}

    return driverName;
}

int getDriver(GDALDriverH **driver, const char *filePath) {
    const char *driverName = getDriverName(filePath);

    *driver = (GDALDriverH*) GDALGetDriverByName(driverName);
    if(*driver == NULL) {
        return 1;
    }

    return 0;
}

int deleteExistingDataset(GDALDriverH driver, const char* filePath) {
    struct stat statBuffer;
    if(stat(filePath, &statBuffer) == 0) {
        if(GDALDeleteDataset(driver, filePath) != CE_None) {
			return 1;
        }
    }

    return 0;
}

int createVectorDataset(GDALDatasetH *dataset, GDALDriverH driver, const char *filePath) {
    if(deleteExistingDataset(driver, filePath)) {
		return 1;
	}

    *dataset = GDALCreate(driver, 
                          filePath, 
                          0, 0, 0, 
                          GDT_Unknown, 
                          NULL);
    if(*dataset == NULL) {
        return 1;
    }
    
    return 0;
}
