#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 5) {
		error("filter requires four arguments. Input file, output file, filter column, and filter value");
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputFilePath = argv[2];
	fprintf(stdout, "output file path: %s\n", outputFilePath);

	const char *filterColumn = argv[3];
	fprintf(stdout, "filter column: %s\n", filterColumn);

	const char *filterValue = argv[4];
	fprintf(stdout, "filter value: %s\n", filterValue);

	struct GDALHandles gdalHandles;
	gdalHandles.inputLayer = NULL;
	if(vectorInitialize(&gdalHandles, inputFilePath, outputFilePath)) {
		error("failed to initialize");
		fatalError();
	}
	
	if(gdalHandles.inputLayer != NULL) {	
        char attrFilterQuery[strlen(filterColumn) + strlen(filterValue) + 4];
        int retCode = snprintf(attrFilterQuery, sizeof(attrFilterQuery), "%s='%s'", filterColumn, filterValue);
        if(retCode < 0) {
            error("failed to build attribute filter query");
			fatalError();
        }
        fprintf(stdout, "attribute filter query: %s\n", attrFilterQuery);

        if(OGR_L_SetAttributeFilter(gdalHandles.inputLayer, attrFilterQuery) != OGRERR_NONE) {
			error("failed to set attribute filter on input layer");
			fatalError();
		}

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
			OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			
			if(buildOutputVectorFeature(&gdalHandles, &inputGeometry, &inputFeature)) {
				error(" failed to build output vector feature");
				fatalError();
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

	fprintf(stdout, "filter complete successfully\n");
	return 0;
}
