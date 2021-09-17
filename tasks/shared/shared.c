#include "shared.h"

void error(const char *message, const char *file, int line) {
	fprintf(stderr, "%s:%d: %s\n", file, line, message);
}

void fatalError() {
	exit(1);
}


int buildOutputVectorFeature(struct GDALHandles *gdalHandles, OGRGeometryH *geometry, OGRFeatureH *inputFeature) {
    OGRFeatureH outputFeature =  OGR_F_Create(gdalHandles->outputFeatureDefn);
    if(outputFeature == NULL) {
        error("failed to create output feature", __FILE__, __LINE__);
        return 1;
    }

    if(OGR_F_SetGeometry(outputFeature, *geometry) != OGRERR_NONE) {
        error("failed to set geometry on output feature", __FILE__, __LINE__);
        return 1;
    }

    for(int i = 0; i < OGR_FD_GetFieldCount(OGR_L_GetLayerDefn(gdalHandles->inputLayer)); ++i) {
        const char *fieldValue = OGR_F_GetFieldAsString(*inputFeature, i);
        OGR_F_SetFieldString(outputFeature, i, fieldValue);
    }

    if(OGR_L_CreateFeature(gdalHandles->outputLayer, outputFeature) != OGRERR_NONE) {
        error("failed to create feature in output layer", __FILE__, __LINE__);
        return 1;
    }    

    OGR_F_Destroy(outputFeature);
}

int createOutputFields(OGRLayerH inLayer, OGRLayerH *outLayer, OGRFeatureDefnH outFeatureDefn) {
    OGRFeatureDefnH inFeatureDef = OGR_L_GetLayerDefn(inLayer);
    int inputFieldCount = OGR_FD_GetFieldCount(inFeatureDef);
    for(int i = 0; i < inputFieldCount; ++i) {
        OGRFieldDefnH fieldDefn = OGR_FD_GetFieldDefn(inFeatureDef, i);
        if(fieldDefn == NULL) {
            error("failed to get input feature definition", __FILE__, __LINE__);
            return 1;
        }

        if(OGR_L_CreateField(*outLayer, fieldDefn, i) != OGRERR_NONE) {
            error("failed to create field in output layer", __FILE__, __LINE__);
            return 1;
        }
    }
    
    return 0;
}

int deleteExistingDataset(GDALDriverH driver, const char* filePath) {
    struct stat statBuffer;
    if(stat(filePath, &statBuffer) == 0) {
        if(GDALDeleteDataset(driver, filePath) != CE_None) {
            error("failed to delete dataset at output location", __FILE__, __LINE__);
			return 1;
        }
    }

    return 0;
}

int createVectorDataset(GDALDatasetH *dataset, GDALDriverH driver, const char *filePath) {
    if(deleteExistingDataset(driver, filePath)) {
        error("failed to delete existing dataset at output location", __FILE__, __LINE__);
		return 1;
	}

    *dataset = GDALCreate(driver, 
                          filePath, 
                          0, 0, 0, 
                          GDT_Unknown, 
                          NULL);
    if(*dataset == NULL) {
        error("failed to create vector dataset", __FILE__, __LINE__);
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
        error("failed to get driver name", __FILE__, __LINE__);
        return 1;
    }

    return 0;
}

int openVectorDataset(GDALDatasetH *dataset, const char *filePath) {
    *dataset = GDALOpenEx(filePath, GDAL_OF_VECTOR, NULL, NULL, NULL);
	if(*dataset == NULL) {
        // TODO improve all error messaging
        // printf("%d\n", CPLGetErrorCounter());
        // printf("%d\n", CPLGetLastErrorNo());
        // printf("%s\n", CPLGetLastErrorMsg());
        error("failed to open vector dataset", __FILE__, __LINE__);
		return 1;
	}
	
    return 0;
}

int vectorInitialize(struct GDALHandles *gdalHandles, const char *inputFilePath, const char *outputFilePath) {
    GDALAllRegister();

    printf("debug: %s\n", inputFilePath);
	GDALDatasetH inputDataset;
	if(openVectorDataset(&inputDataset, inputFilePath)) {
		error("failed to open input vector dataset", __FILE__, __LINE__);
		return 1;
	}
    gdalHandles->inputDataset = inputDataset;

	GDALDriverH *driver;
	if(getDriver(&driver, outputFilePath)) {
		error("failed to create driver", __FILE__, __LINE__);
		return 1;
	}

	int numberOfLayers = GDALDatasetGetLayerCount(inputDataset);
	if(numberOfLayers > 0) {
		OGRLayerH inputLayer = GDALDatasetGetLayer(inputDataset, 0);
		if(inputLayer == NULL) {
			error("failed to get layer from intput dataset", __FILE__, __LINE__);
			return 1;
		}  
        gdalHandles->inputLayer = inputLayer;

		OGR_L_ResetReading(inputLayer);

		GDALDatasetH outputDataset;
		if(createVectorDataset(&outputDataset, driver, outputFilePath)) {
			error("failed to create output vector dataset", __FILE__, __LINE__);
			return 1;
		}
        gdalHandles->outputDataset = outputDataset;

		OGRSpatialReferenceH inputSpatialRef = OGR_L_GetSpatialRef(inputLayer);
        gdalHandles->inputSpatialRef = inputSpatialRef;
		OGRLayerH outputLayer = GDALDatasetCreateLayer(outputDataset, 
													   OGR_L_GetName(inputLayer),  
													   inputSpatialRef, 
													   wkbPolygon, 
													   NULL);
		if(outputLayer == NULL) {
			error("failed to create output layer", __FILE__, __LINE__);
			return 1;
		}
        gdalHandles->outputLayer = outputLayer;

		OGRFeatureDefnH outputFeatureDefn = OGR_L_GetLayerDefn(outputLayer);
		if(createOutputFields(inputLayer, &outputLayer, outputFeatureDefn)) {
			error("failed to create fields on output layer", __FILE__, __LINE__);
			return 1;
		}
        gdalHandles->outputFeatureDefn = outputFeatureDefn;
    }

    return 0;
}

int unzip(const char *inputFilePath, char **unzipDir) {
    char *dupeInputFilePath = strdup(inputFilePath);
    *unzipDir = dirname(dupeInputFilePath);
    char cmd[256];
    snprintf(cmd, sizeof(cmd), "%s%s%s%s", "unzip -o ", inputFilePath, " -d ", *unzipDir);    

    int unzipResult = system(cmd);
    if(unzipResult != 0) {
        error("failed to unzip input file", __FILE__, __LINE__);
        return 1;
    }

    return 0;
}

int isShpFile(const char *filename) {
    const char *suffix = ".shp";

    size_t filenameLength = strlen(filename);
    size_t suffixLength = strlen(suffix);

    return strncmp(filename + filenameLength - suffixLength, suffix, suffixLength) == 0;
}

int getShpFilePath(const char *unzipDir, char **shpFilePath) {
    DIR *d = opendir(unzipDir);
    struct dirent *dir;
    if(d) {
        int shpFileFound = 0;
        while((dir = readdir(d)) != NULL) {
            if(isShpFile(dir->d_name)) {
                shpFileFound = 1;
                char tmpShpFilePath[256];
                snprintf(tmpShpFilePath, sizeof(tmpShpFilePath), "%s/%s", unzipDir, dir->d_name);
                printf("debug1: %s\n", tmpShpFilePath);
                *shpFilePath = &tmpShpFilePath[0];
                printf("debug2: %s\n", *shpFilePath);
                break;
            }
        }

        if(!shpFileFound) {
            error("shp file not found in input", __FILE__, __LINE__);
            return 1;
        }
        closedir(d);
    } else {
        error("unable to open unzip directory", __FILE__, __LINE__);
        return 1;
    }

    return 0;
}

int getInputGeoFilePath(const char *inputFilePath, char **inputGeoFilePath) {
    char *ext = strrchr(inputFilePath, '.');
    if(ext && !strcmp(ext, ".geojson")) {
        *inputGeoFilePath = strdup(inputFilePath);
    } else if(ext && !strcmp(ext, ".zip")) {
        char *unzipDir;
        int unzipResult = unzip(inputFilePath, &unzipDir);
        if(unzipResult != 0) {
            error("failed to unzip input file", __FILE__, __LINE__);
            return 1;
        }

        char *shpFilePath;
        if(getShpFilePath(unzipDir, &shpFilePath)) {
            error("failed to get shp file path", __FILE__, __LINE__);
            return 1;
        }
        printf("debug3: %s\n", shpFilePath);
        *inputGeoFilePath = shpFilePath;
    } else {
        error("unrecognized input file", __FILE__, __LINE__);
        return 1;
    }

    return 0;
}
