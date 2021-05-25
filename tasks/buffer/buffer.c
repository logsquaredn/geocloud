#include "gdal.h"
#include "ogr_srs_api.h"
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
		error("please pass a valid double greater than 0");
	}
	fprintf(stdout, "buffer distance: %f\n", bufferDistanceDouble);

	GDALAllRegister();
	fprintf(stdout, "registered gdal successfully\n");

	GDALDatasetH inputDataset;
	if(openVectorDataset(&inputDataset, inputFilePath)) {
		error("failed to open input file");
	}
	fprintf(stdout, "opened input file successfully\n");

	const char *driverName = getDriverName(outputFilePath);
	fprintf(stdout, "driver name: %s\n", driverName);

	GDALDriverH *driver;
	if(getDriver(&driver, driverName)) {
		error("failed to create driver");
	}
	fprintf(stdout, "created driver successfully\n");

	int numberOfLayers = GDALDatasetGetLayerCount(inputDataset);
	if(numberOfLayers > 0) {
		OGRLayerH inputLayer = GDALDatasetGetLayer(inputDataset, 0);
		if(inputLayer == NULL) {
			error("failed to get layer from intput dataset");
		}
		fprintf(stdout, "got layer from input dataset successfully\n");
		
		OGR_L_ResetReading(inputLayer);
		fprintf(stdout, "reset layer reading successfully\n");

	    if(deleteExistingDataset(driver, outputFilePath)) {
			error("failed to clean output location");
		}
		fprintf(stdout, "output location is clean\n");
		
		GDALDatasetH outputDataset;
		if(createVectorDataset(&outputDataset, driver, outputFilePath)) {
			error("failed to create output dataset");
		}
		fprintf(stdout, "created output dataset successfully\n");

		OGRLayerH outputLayer = GDALDatasetCreateLayer(outputDataset, 
													   OGR_L_GetName(inputLayer),  
													   OGR_L_GetSpatialRef(inputLayer), 
													   wkbPolygon, 
													   NULL);
		if(outputLayer == NULL) {
			error("failed to create output layer");
		}
		fprintf(stdout, "created output layer successfully\n");
		OGRFeatureDefnH outputFeatureDef = OGR_L_GetLayerDefn(outputLayer);

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(inputLayer)) != NULL) {
			OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			if(inputGeometry == NULL) {
				error("failed to get input geometry");		
			}
			fprintf(stdout, "got input geometry successfully\n");

			OGRGeometryH bufferedGeometry = OGR_G_Buffer(inputGeometry, bufferDistanceDouble, 50);
			if(bufferedGeometry == NULL) {
				error("failed to buffer input geometry");
			}
			fprintf(stdout, "buffered input geometry successfully\n");

			OGRFeatureH outputBufferedFeature =  OGR_F_Create(outputFeatureDef);
			if(OGR_F_SetGeometry(outputBufferedFeature, bufferedGeometry) != OGRERR_NONE) {
				error("failed to set buffered geometry on buffered feature");
			}
			fprintf(stdout, "set buffered geometry on output feature successfully\n");

			if(OGR_L_CreateFeature(outputLayer, outputBufferedFeature) != OGRERR_NONE) {
				error("failed to create buffered feature in output layer");
			}
			fprintf(stdout, "created buffered feature in output layer successfully\n");

			OGR_G_DestroyGeometry(inputGeometry);
			OGR_G_DestroyGeometry(bufferedGeometry);
			OGR_F_Destroy(outputBufferedFeature);
		}

		OGR_F_Destroy(inputFeature);
		GDALClose(outputDataset);
	}

	GDALClose(inputDataset);

	fprintf(stdout, "buffer complete successfully\n");
	return 0;
}
