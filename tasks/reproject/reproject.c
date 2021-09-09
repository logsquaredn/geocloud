#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 4) {
		error("reproject requires three arguments. Input file, output file, and target projection in EPSG code format", __FILE__, __LINE__);
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputFilePath = argv[2];
	fprintf(stdout, "output file path: %s\n", outputFilePath);

	const char *targetProjection = argv[3];	
	long targetProjectionLong = strtol(targetProjection, NULL, 10);
	if(targetProjectionLong == 0) {
		error("EPSG code must be an integer", __FILE__, __LINE__);
	}
	fprintf(stdout, "target projection: %ld\n", targetProjectionLong);

	struct GDALHandles gdalHandles;
	gdalHandles.inputLayer = NULL;
	if(vectorInitialize(&gdalHandles, inputFilePath, outputFilePath)) {
		error("failed to initialize", __FILE__, __LINE__);
		fatalError();
	}
	
	if(gdalHandles.inputLayer != NULL) {	
		OGRSpatialReferenceH outputSpatialRef = OSRNewSpatialReference("");
		if(OSRImportFromEPSG(outputSpatialRef, targetProjectionLong) != OGRERR_NONE) {
			error("failed to create output spatial ref", __FILE__, __LINE__);
			fatalError();
		}
		OGRCoordinateTransformationH transformer = OCTNewCoordinateTransformation(gdalHandles.inputSpatialRef, outputSpatialRef);

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
            OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
			
			if(OGR_G_Transform(inputGeometry, transformer) != OGRERR_NONE) {
				error("failed to transform geometry", __FILE__, __LINE__);
				fatalError();
			}

			if(buildOutputVectorFeature(&gdalHandles, &inputGeometry, &inputFeature)) {
				error(" failed to build output vector feature", __FILE__, __LINE__);
				fatalError();
			}

			OGR_G_DestroyGeometry(inputGeometry);
		}

		OSRDestroySpatialReference(outputSpatialRef);
		OCTDestroyCoordinateTransformation(transformer);
		OGR_F_Destroy(inputFeature);
		OSRDestroySpatialReference(gdalHandles.inputSpatialRef);
		GDALClose(gdalHandles.outputDataset);
	} else {
		fprintf(stdout, "no layers found in input file\n");
	}

	GDALClose(gdalHandles.inputDataset);

	fprintf(stdout, "reproject complete successfully\n");
	return 0;
}
