#include <stdio.h>

#include "../shared/shared.h"

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
	
	struct GDALHandles gdalHandles;
	gdalHandles.inputLayer = NULL;
	if(vectorInitialize(&gdalHandles, inputFilePath, outputFilePath)) {
		error("failed to initialize");
		fatalError();
	}
	
	if(gdalHandles.inputLayer != NULL) {
		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
			OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			if(inputGeometry == NULL) {
				error("failed to get input geometry");	
				fatalError();	
			}

			OGRGeometryH bufferedGeometry = OGR_G_Buffer(inputGeometry, bufferDistanceDouble, 50);
			if(bufferedGeometry == NULL) {
				error("failed to buffer input geometry");
				fatalError();
			}

			if(buildOutputVectorFeature(&gdalHandles, &bufferedGeometry, &inputFeature)) {
				error(" failed to build output vector feature");
				fatalError();
			}

			OGR_G_DestroyGeometry(inputGeometry);
			OGR_G_DestroyGeometry(bufferedGeometry);
		}

		OGR_F_Destroy(inputFeature);
		OSRDestroySpatialReference(gdalHandles.inputSpatialRef);
		GDALClose(gdalHandles.outputDataset);
	} else {
		fprintf(stdout, "no layers found in input file\n");
	}

	GDALClose(gdalHandles.inputDataset);

	fprintf(stdout, "buffer complete successfully\n");
	return 0;
}
