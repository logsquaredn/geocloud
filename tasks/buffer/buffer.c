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

	GDALDatasetH iDs = GDALOpenEx(vfp, GDAL_OF_VECTOR, NULL, NULL, NULL);
	if(iDs == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to open vector file: %s", vfp);
		fatalError(eMsg, __FILE__, __LINE__);	
	}
	free(vfp);

	int lCount = GDALDatasetGetLayerCount(iDs);
	if(lCount < 1) {
		fatalError("input dataset has no layers", __FILE__, __LINE__);
	}
	OGRLayerH iLay = GDALDatasetGetLayer(iDs, 0);
	if(iLay == NULL) {
		fatalError("failed to get layer from input dataset", __FILE__, __LINE__);
	}
	OGR_L_ResetReading(iLay);
	
	GDALDriverH shpD = GDALGetDriverByName("ESRI Shapefile");
	if(shpD == NULL) {
		fatalError("failed to get shapefile driver", __FILE__, __LINE__);
	}
	// TODO create output file



	OGRFeatureH iFeat;
	while((iFeat = OGR_L_GetNextFeature(iLay)) != NULL) {
		OGRGeometryH iGeom = OGR_F_GetGeometryRef(iFeat);
		if(iGeom == NULL) {
			fatalError("failed to get input geometry", __FILE__, __LINE__);			
		}

		OGRGeometryH rebuiltBuffGeom = OGR_G_CreateGeometry(wkbMultiPolygon);
		OGRGeometryH splitGeoms[4 * ONE_KB];
		int geomCount = splitGeometries(splitGeoms, iGeom, 0);
		if(geomCount < 0) {
			fatalError("failed to split geometries", __FILE__, __LINE__);
		}

		for(int i = 0; i < geomCount; ++i) {
			OGRGeometryH buffGeom = OGR_G_Buffer(splitGeoms[i], bd, qsc);
			if(buffGeom == NULL) {
				fatalError("failed to buffer input geometry", __FILE__, __LINE__);
			}

			if(OGR_G_AddGeometry(rebuiltBuffGeom, buffGeom) != OGRERR_NONE) {
	 			fatalError("failed to add buffered geometry to rebuilt geometry", __FILE__, __LINE__);
	 		}

			OGR_G_DestroyGeometry(splitGeoms[i]);
			OGR_G_DestroyGeometry(buffGeom);
		}

		// TODO
	}

	OGR_F_Destroy(iFeat);
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

	GDALClose(iDs);

	info("buffer complete successfully");
	return 0;
}
