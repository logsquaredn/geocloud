#include <stdio.h>

#include "../shared/shared.h"

void error(const char *message) {
	fprintf(stderr, "%s\n", message);
	exit(1);
}

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

	GDALAllRegister();

	GDALDatasetH inputDataset;
	if(openVectorDataset(&inputDataset, inputFilePath)) {
		error("failed to open input file");
	}

	GDALDriverH *driver;
	if(getDriver(&driver, outputFilePath)) {
		error("failed to create driver");
	}

	int numberOfLayers = GDALDatasetGetLayerCount(inputDataset);
	if(numberOfLayers > 0) {
		OGRLayerH inputLayer = GDALDatasetGetLayer(inputDataset, 0);
		if(inputLayer == NULL) {
			error("failed to get layer from intput dataset");
		}
		
		OGR_L_ResetReading(inputLayer);

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

        char attrFilterQuery[strlen(filterColumn) + strlen(filterValue) + 4];
        int retCode = snprintf(attrFilterQuery, sizeof(attrFilterQuery), "%s='%s'", filterColumn, filterValue);
        if(retCode < 0) {
            error("failed to build attribute filter query");
        }
        fprintf(stdout, "attribute filter query: %s\n", attrFilterQuery);

        if(OGR_L_SetAttributeFilter(inputLayer, attrFilterQuery) != OGRERR_NONE) {
			error("failed to set attribute filter on input layer");
		}

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(inputLayer)) != NULL) {
			if(OGR_L_CreateFeature(outputLayer, inputFeature) != OGRERR_NONE) {
				error("failed to create feature in output layer");
			}
		}

		OSRDestroySpatialReference(inputSpatialRef);
		OGR_F_Destroy(inputFeature);
		GDALClose(outputDataset);
	}

	GDALClose(inputDataset);

	fprintf(stdout, "filter complete successfully\n");
	return 0;
}
