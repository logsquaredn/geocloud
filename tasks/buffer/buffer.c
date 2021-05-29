#include <stdio.h>

#include "../shared/shared.h"

void error(const char *message) {
	fprintf(stderr, "%s\n", message);
	exit(1);
}

int main(int argc, char *argv[]) {
	if(argc != 4) {
		error("buffer requires three arguments. input file, output file, and a buffer distance");
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputFilePath = argv[2];
	fprintf(stdout, "output file path: %s\n", outputFilePath);

	const char *bufferDistance = argv[3];
	double bufferDistanceDouble = strtod(bufferDistance, NULL);
	if(bufferDistanceDouble == 0) {
		error("buffer distance must be a valid double greater than 0");
	}
	fprintf(stdout, "buffer distance: %f\n", bufferDistanceDouble);

	GDALAllRegister();

	GDALDatasetH inputDataset;
	if(openVectorDataset(&inputDataset, inputFilePath)) {
		error("failed to open input file");
	}

	const char *driverName = getDriverName(outputFilePath);
	fprintf(stdout, "driver name: %s\n", driverName);

	GDALDriverH *driver;
	if(getDriver(&driver, driverName)) {
		error("failed to create driver");
	}

	int numberOfLayers = GDALDatasetGetLayerCount(inputDataset);
	if(numberOfLayers > 0) {
		OGRLayerH inputLayer = GDALDatasetGetLayer(inputDataset, 0);
		if(inputLayer == NULL) {
			error("failed to get layer from intput dataset");
		}
		
		OGR_L_ResetReading(inputLayer);

	    if(deleteExistingDataset(driver, outputFilePath)) {
			error("failed to clean output location");
		}
		
		GDALDatasetH outputDataset;
		if(createVectorDataset(&outputDataset, driver, outputFilePath)) {
			error("failed to create output dataset");
		}

		OGRSpatialReferenceH inputSpatialRef = OGR_L_GetSpatialRef(inputLayer);
		OGRLayerH outputLayer = GDALDatasetCreateLayer(outputDataset, 
													   OGR_L_GetName(inputLayer),  
													   inputSpatialRef, 
													   wkbPolygon, 
													   NULL);
		if(outputLayer == NULL) {
			error("failed to create output layer");
		}

		OGRFeatureDefnH outputFeatureDef = OGR_L_GetLayerDefn(outputLayer);

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(inputLayer)) != NULL) {
			OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			if(inputGeometry == NULL) {
				error("failed to get input geometry");		
			}

			OGRGeometryH bufferedGeometry = OGR_G_Buffer(inputGeometry, bufferDistanceDouble, 50);
			if(bufferedGeometry == NULL) {
				error("failed to buffer input geometry");
			}

			OGRFeatureH outputBufferedFeature =  OGR_F_Create(outputFeatureDef);
			if(OGR_F_SetGeometry(outputBufferedFeature, bufferedGeometry) != OGRERR_NONE) {
				error("failed to set buffered geometry on buffered feature");
			}

			if(OGR_L_CreateFeature(outputLayer, outputBufferedFeature) != OGRERR_NONE) {
				error("failed to create buffered feature in output layer");
			}

			OGR_G_DestroyGeometry(inputGeometry);
			OGR_G_DestroyGeometry(bufferedGeometry);
			OGR_F_Destroy(outputBufferedFeature);
		}

		OSRDestroySpatialReference(inputSpatialRef);
		OGR_F_Destroy(inputFeature);
		GDALClose(outputDataset);
	}

	GDALClose(inputDataset);

	fprintf(stdout, "buffer complete successfully\n");
	return 0;
}
