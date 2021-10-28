#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 5) {
		error("filter requires four arguments. Input file, output directory, filter column, and filter value", __FILE__, __LINE__);
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputDir = argv[2];
	fprintf(stdout, "output directory: %s\n", outputDir);

	const char *filterColumn = argv[3];
	fprintf(stdout, "filter column: %s\n", filterColumn);

	const char *filterValue = argv[4];
	fprintf(stdout, "filter value: %s\n", filterValue);

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
        char attrFilterQuery[strlen(filterColumn) + strlen(filterValue) + 4];
        int retCode = snprintf(attrFilterQuery, sizeof(attrFilterQuery), "%s='%s'", filterColumn, filterValue);
        if(retCode < 0) {
            error("failed to build attribute filter query", __FILE__, __LINE__);
			fatalError();
        }
        fprintf(stdout, "attribute filter query: %s\n", attrFilterQuery);

        if(OGR_L_SetAttributeFilter(gdalHandles.inputLayer, attrFilterQuery) != OGRERR_NONE) {
			error("failed to set attribute filter on input layer", __FILE__, __LINE__);
			fatalError();
		}

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
			OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			
			if(buildOutputVectorFeature(&gdalHandles, &inputGeometry, &inputFeature)) {
				error(" failed to build output vector feature", __FILE__, __LINE__);
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

	// TODO this seg faults on some geojson input
	// GDALClose(gdalHandles.inputDataset);

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

	fprintf(stdout, "filter complete successfully\n");
	return 0;
}
