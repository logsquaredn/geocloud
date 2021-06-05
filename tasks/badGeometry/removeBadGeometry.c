#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 3) {
		error("remove bad geometry requires two arguments. Input file and output file");
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputFilePath = argv[2];
	fprintf(stdout, "output file path: %s\n", outputFilePath);

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
			if(OGR_G_IsValid(inputGeometry)) {	
				if(buildOutputVectorFeature(&gdalHandles, &inputGeometry, &inputFeature)) {
					error(" failed to build output vector feature");
					fatalError();
				}
			}

			OGR_G_DestroyGeometry(inputGeometry);
		}

		OGR_F_Destroy(inputFeature);
		OSRDestroySpatialReference(gdalHandles.inputSpatialRef);
		GDALClose(gdalHandles.outputDataset);
	} else {
		fprintf(stdout, "no layers found in input file\n");
	}

	GDALClose(gdalHandles.inputDataset);

	fprintf(stdout, "removed bad geometry successfully\n");
	return 0;
}
