#include <stdio.h>

#include "../shared/shared.h"

void error(const char *message) {
	fprintf(stderr, "%s\n", message);
	exit(1);
}

int main(int argc, char *argv[]) {
	if(argc != 4) {
		error("reproject requires three arguments. Input file, output file, and target projection in EPSG code format");
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputFilePath = argv[2];
	fprintf(stdout, "output file path: %s\n", outputFilePath);

	const char *targetProjection = argv[3];	
	long targetProjectionLong = strtol(targetProjection, NULL, 10);
	if(targetProjectionLong == 0) {
		error("EPSG code must be an integer");
	}
	fprintf(stdout, "target projection: %ld\n", targetProjectionLong);

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
        		
		OGRSpatialReferenceH outputSpatialRef = OSRNewSpatialReference("");
		if(OSRImportFromEPSG(outputSpatialRef, targetProjectionLong) != OGRERR_NONE) {
			error("failed to create output spatial ref");
		}
		OGRCoordinateTransformationH transformer = OCTNewCoordinateTransformation(inputSpatialRef, outputSpatialRef);

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(inputLayer)) != NULL) {
            OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			
			if(OGR_G_Transform(inputGeometry, transformer) != OGRERR_NONE) {
				error("failed to transform geometry");
			}

			OGRFeatureH outputReprojectedFeature =  OGR_F_Create(outputFeatureDef);
			if(outputReprojectedFeature == NULL) {
				error("failed to create output feature");
			}

			if(OGR_F_SetGeometry(outputReprojectedFeature, inputGeometry) != OGRERR_NONE) {
				error("failed to set geometry on output feature");
			}
            
			if(OGR_L_CreateFeature(outputLayer, outputReprojectedFeature) != OGRERR_NONE) {
				error("failed to create feature in output layer");
			}

			OGR_G_DestroyGeometry(inputGeometry);
			OGR_F_Destroy(outputReprojectedFeature);
		}

		OSRDestroySpatialReference(inputSpatialRef);
		OSRDestroySpatialReference(outputSpatialRef);
		OCTDestroyCoordinateTransformation(transformer);
		OGR_F_Destroy(inputFeature);
		GDALClose(outputDataset);
	}

	GDALClose(inputDataset);

	fprintf(stdout, "reproject complete successfully\n");
	return 0;
}
