#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 5) {
		error("buffer requires four arguments. input file, output directory, buffer distance, and quadrant segment count", __FILE__, __LINE__);
		fatalError();
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input filepath: %s\n", inputFilePath);
	
	const char *outputDir = argv[2];
	fprintf(stdout, "output directory: %s\n", outputDir);

	const char *bufferDistance = argv[3];
	double bufferDistanceDouble = strtod(bufferDistance, NULL);
	if(bufferDistanceDouble == 0) {
		error("buffer distance must be a positive double", __FILE__, __LINE__);
		fatalError();
	}
	fprintf(stdout, "buffer distance: %f\n", bufferDistanceDouble);

	const char *quadSegCount = argv[4];
	int quadSegCountInt = atoi(quadSegCount);
	if(quadSegCountInt <= 0) {
		error("quadrant segment count must be a positive integer", __FILE__, __LINE__);
		fatalError();
	}
	fprintf(stdout, "quadrant segment count: %d\n", quadSegCountInt);
	
	char *inputGeoFilePath = getInputGeoFilePathVector(inputFilePath);
	if(inputGeoFilePath == NULL) {
		error("failed to find input geo filepath", __FILE__, __LINE__);
		fatalError();
	}
	fprintf(stdout, "input geo filepath: %s\n", inputGeoFilePath);

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
			if(inputGeometry == NULL) {
				error("failed to get input geometry", __FILE__, __LINE__);	
				fatalError();	
			}

			OGRGeometryH rebuiltBufferedGeometry = OGR_G_CreateGeometry(wkbMultiPolygon);
			OGRGeometryH splitGeoms[4096];
			int geomsCount = splitGeometries(splitGeoms, 0, inputGeometry);
			for(int i = 0; i < geomsCount; ++i) {
				OGRGeometryH bufferedGeometry = OGR_G_Buffer(splitGeoms[i], bufferDistanceDouble, quadSegCountInt);
				if(bufferedGeometry == NULL) {
					error("failed to buffer input geometry", __FILE__, __LINE__);
					fatalError();
				}

				if(OGR_G_AddGeometry(rebuiltBufferedGeometry, bufferedGeometry) != OGRERR_NONE) {
					error("failed to add buffered geometry to rebuilt geometry", __FILE__, __LINE__);
					fatalError();
				}
				
				OGR_G_DestroyGeometry(splitGeoms[i]);
				OGR_G_DestroyGeometry(bufferedGeometry);
			}

			if(buildOutputVectorFeature(&gdalHandles, &rebuiltBufferedGeometry, &inputFeature)) {
				error("failed to build output vector feature", __FILE__, __LINE__);
				fatalError();
			}

			OGR_G_DestroyGeometry(rebuiltBufferedGeometry);
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

	fprintf(stdout, "buffer complete successfully\n");
	return 0;
}
