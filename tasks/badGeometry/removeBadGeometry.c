#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 3) {
		error("remove bad geometry requires two arguments. Input file and output directory", __FILE__, __LINE__);
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputDir = argv[2];
	fprintf(stdout, "output directory: %s\n", outputDir);

	char *inputGeoFilePath;
	if(getInputGeoFilePath(inputFilePath, &inputGeoFilePath)) {
		error("failed to find input geo file path", __FILE__, __LINE__);
	}
	fprintf(stdout, "input geo file path: %s\n", inputGeoFilePath);

	struct GDALHandles gdalHandles;
	gdalHandles.inputLayer = NULL;
	if(vectorInitialize(&gdalHandles, inputGeoFilePath, outputDir)) {
		error("failed to initialize", __FILE__, __LINE__);
		fatalError();
	}
	free(inputGeoFilePath);

	if(gdalHandles.inputLayer != NULL) {
		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
            OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			if(OGR_G_IsValid(inputGeometry)) {	
				if(buildOutputVectorFeature(&gdalHandles, &inputGeometry, &inputFeature)) {
					error(" failed to build output vector feature", __FILE__, __LINE__);
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

	if(zipShp(outputDir)) {
		error("failed to zip up shp", __FILE__, __LINE__);
		fatalError();
	}

	if(dumpToGeojson(outputDir)) {
		error("failed to convert shp to geojson", __FILE__, __LINE__);
		fatalError();
	}

	if(cleanup(outputDir)) {
		error("failed to cleanup output", __FILE__, __LINE__);
		fatalError();
	}

	fprintf(stdout, "removed bad geometry successfully\n");
	return 0;
}
