#include <stdio.h>

#include "../shared/shared.h"

void error(const char *message) {
	fprintf(stderr, "%s\n", message);
	exit(1);
}

int main(int argc, char *argv[]) {
	if(argc != 3) {
		error("remove bad geometry requires two arguments. Input file and output file");
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputFilePath = argv[2];
	fprintf(stdout, "output file path: %s\n", outputFilePath);

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
        		
		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(inputLayer)) != NULL) {
            OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			if(OGR_G_IsValid(inputGeometry)) {	
				OGRFeatureH outputFeature = OGR_F_Create(outputFeatureDef);
				if(outputFeature == NULL) {
					error("failed to create output feature");
				}

				if(OGR_F_SetGeometry(outputFeature, inputGeometry) != OGRERR_NONE) {
					error("failed to set geometry on output feature");
				}
            
				if(OGR_L_CreateFeature(outputLayer, outputFeature) != OGRERR_NONE) {
					error("failed to create feature in output layer");
				}

				OGR_F_Destroy(outputFeature);
			}

			OGR_G_DestroyGeometry(inputGeometry);
		}

		OSRDestroySpatialReference(inputSpatialRef);
		OGR_F_Destroy(inputFeature);
		GDALClose(outputDataset);
	}

	GDALClose(inputDataset);

	fprintf(stdout, "removed bad geometry successfully\n");
	return 0;
}
