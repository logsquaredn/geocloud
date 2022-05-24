#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	GDALAllRegister();

	char iMsg[ONE_KB];

	const char *ifp = getenv(ENV_VAR_INPUT_FILEPATH);
	if(ifp == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_INPUT_FILEPATH);
		fatalError(eMsg, __FILE__, __LINE__);
	}
	sprintf(iMsg, "input filepath: %s", ifp);
	info(iMsg);
	
	const char *od = getenv(ENV_VAR_OUTPUT_DIRECTORY);
	if(od == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_OUTPUT_DIRECTORY);
		fatalError(eMsg, __FILE__, __LINE__);
	}
	sprintf(iMsg, "output directory: %s", od);
	info(iMsg);

	const char *bdArg = getenv("GEOCLOUD_BUFFER_DISTANCE");
	if(bdArg == NULL) {
		fatalError("env var: GEOCLOUD_BUFFER_DISTANCE must be set", __FILE__, __LINE__);		
	}
	double bd = strtod(bdArg, NULL);
	if(bd == 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "buffer distance must be a postitive double. got: %s", bdArg);
		fatalError(eMsg, __FILE__, __LINE__);
	}
	sprintf(iMsg, "buffer distance: %f", bd);
	info(iMsg);

	const char *qscArg = getenv("GEOCLOUD_QUADRANT_SEGMENT_COUNT");
	if(qscArg == NULL) {
		fatalError("env var: GEOCLOUD_QUADRANT_SEGMENT_COUNT must be set", __FILE__, __LINE__);		
	}
	int qsc = atoi(qscArg);
	if(qsc <= 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "quadrant segment count must be a positive integer. got: %s", qscArg);
		fatalError(eMsg, __FILE__, __LINE__);
	}
	sprintf(iMsg, "quadrant segment count: %d", qsc);
	info(iMsg);

	char *vfp = NULL;
	if(isZip(ifp)) {
		char **fl = unzip(ifp);
		if(fl == NULL) {
			char eMsg[ONE_KB];
			sprintf(eMsg, "failed to unzip: %s", ifp);
			fatalError(eMsg, __FILE__, __LINE__);	
		}

		char *f;
		int fc = 0;
		while((f = fl[fc]) != NULL) {
			if(isShp(f)) {
				vfp = f;
			} else {
				free(fl[fc]);
			}
			++fc;
		}

		if(vfp == NULL) {
			fatalError("input zip must contain a shp file", __FILE__, __LINE__);
		}

		free(fl);
	} else if(isGeojson(ifp)) {
		vfp = strdup(ifp);
	} else {
		fatalError("input file must be a .zip or .geojson", __FILE__, __LINE__);
	}
	sprintf(iMsg, "vector filepath: %s", vfp);
	info(iMsg); 

	GDALDatasetH gd = GDALOpenEx(vfp, GDAL_OF_VECTOR, NULL, NULL, NULL);
	if(gd == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to open vector file: %s", vfp);
		fatalError(eMsg, __FILE__, __LINE__);	
	}
	free(vfp);



	// struct GDALHandles gdalHandles;
	// gdalHandles.inputLayer = NULL;
	// if(vectorInitialize(&gdalHandles, inputGeoFilePath, outputDir)) {
	// 	error("failed to initialize", __FILE__, __LINE__);
	// 	fatalError();
	// }
	// free(inputGeoFilePath);
	
	// if(gdalHandles.inputLayer != NULL) {
	// 	OGRFeatureH inputFeature;
	// 	while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
	// 		OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
	// 		if(inputGeometry == NULL) {
	// 			error("failed to get input geometry", __FILE__, __LINE__);	
	// 			fatalError();	
	// 		}

	// 		OGRGeometryH rebuiltBufferedGeometry = OGR_G_CreateGeometry(wkbMultiPolygon);
	// 		OGRGeometryH splitGeoms[4096];
	// 		int geomsCount = splitGeometries(splitGeoms, 0, inputGeometry);
	// 		for(int i = 0; i < geomsCount; ++i) {
	// 			OGRGeometryH bufferedGeometry = OGR_G_Buffer(splitGeoms[i], bufferDistanceDouble, quadSegCountInt);
	// 			if(bufferedGeometry == NULL) {
	// 				error("failed to buffer input geometry", __FILE__, __LINE__);
	// 				fatalError();
	// 			}

	// 			if(OGR_G_AddGeometry(rebuiltBufferedGeometry, bufferedGeometry) != OGRERR_NONE) {
	// 				error("failed to add buffered geometry to rebuilt geometry", __FILE__, __LINE__);
	// 				fatalError();
	// 			}
				
	// 			OGR_G_DestroyGeometry(splitGeoms[i]);
	// 			OGR_G_DestroyGeometry(bufferedGeometry);
	// 		}

	// 		if(buildOutputVectorFeature(&gdalHandles, &rebuiltBufferedGeometry, &inputFeature)) {
	// 			error("failed to build output vector feature", __FILE__, __LINE__);
	// 			fatalError();
	// 		}

	// 		OGR_G_DestroyGeometry(rebuiltBufferedGeometry);
	// 	}

	// 	OGR_F_Destroy(inputFeature);
	// 	OSRDestroySpatialReference(gdalHandles.inputSpatialRef);
	// 	GDALClose(gdalHandles.outputDataset);
	// } else {
	// 	fprintf(stdout, "no layers found in input file\n");
	// }


	// if(zipShp(outputDir)) {
	// 	error("failed to zip up shp", __FILE__, __LINE__);
	// 	fatalError();
	// }

	// if(dumpToGeojson(outputDir)) {
	// 	error("failed to convert shp to geojson", __FILE__, __LINE__);
	// 	fatalError();
	// }

	// if(cleanup(outputDir)) {
	// 	error("failed to cleanup output", __FILE__, __LINE__);
	// 	fatalError();
	// }

	GDALClose(gd);

	info("buffer complete successfully");
	return 0;
}
