#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	GDALAllRegister();

	char iMsg[ONE_KB];

	char *iFp = getenv(ENV_VAR_INPUT_FILEPATH);
	if(iFp == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_INPUT_FILEPATH);
		fatalError(eMsg, __FILE__, __LINE__);
	}
	sprintf(iMsg, "input filepath: %s", iFp);
	info(iMsg);
	
	const char *oDir = getenv(ENV_VAR_OUTPUT_DIRECTORY);
	if(oDir == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_OUTPUT_DIRECTORY);
		fatalError(eMsg, __FILE__, __LINE__);
	}
	sprintf(iMsg, "output directory: %s", oDir);
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

	int isInputShp = 0;
	char *vFp = NULL;
	if(isZip(iFp)) {
		isInputShp = 1;
		char **fl = unzip(iFp);
		if(fl == NULL) {
			char eMsg[ONE_KB];
			sprintf(eMsg, "failed to unzip: %s", iFp);
			fatalError(eMsg, __FILE__, __LINE__);	
		}

		char *f;
		int fc = 0;
		while((f = fl[fc]) != NULL) {
			if(isShp(f)) {
				vFp = f;
			} else {
				free(fl[fc]);
			}
			++fc;
		}

		if(vFp == NULL) {
			fatalError("input zip must contain a shp file", __FILE__, __LINE__);
		}

		free(fl);
	} else if(isGeojson(iFp)) {
		vFp = strdup(iFp);
	} else {
		fatalError("input file must be a .zip or .geojson", __FILE__, __LINE__);
	}
	sprintf(iMsg, "vector filepath: %s", vFp);
	info(iMsg); 

	GDALDatasetH iDs = OGROpen(vFp, 1, NULL);
	if(iDs == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to open vector file: %s", vFp);
		fatalError(eMsg, __FILE__, __LINE__);	
	}

	int lCount = GDALDatasetGetLayerCount(iDs);
	if(lCount < 1) {
		fatalError("input dataset has no layers", __FILE__, __LINE__);
	}
	OGRLayerH iLay = OGR_DS_GetLayer(iDs, 0);
	if(iLay == NULL) {
		fatalError("failed to get layer from input dataset", __FILE__, __LINE__);
	}
	OGR_L_ResetReading(iLay);
	
	OGRFeatureH iFeat;
	while((iFeat = OGR_L_GetNextFeature(iLay)) != NULL) {
		OGRGeometryH iGeom = OGR_F_StealGeometry(iFeat);
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

		if(OGR_F_SetGeometry(iFeat, rebuiltBuffGeom) != OGRERR_NONE) {
			fatalError("failed set updated geometry on input feature", __FILE__, __LINE__);
		}
		
		if(OGR_L_SetFeature(iLay, iFeat) != OGRERR_NONE) {
			fatalError("failed to set updated feature on input layer", __FILE__, __LINE__);
		}

		OGR_G_DestroyGeometry(rebuiltBuffGeom);
		OGR_F_Destroy(iFeat);
	}

	GDALClose(iDs);

	if(isInputShp) {
	 	if(produceShpOutput(iFp, oDir, vFp) != 0) {
			 fatalError("failed to produce output", __FILE__, __LINE__);
		 }
	} else {
		if(produceJsonOutput(iFp, oDir) != 0) {
			fatalError("failed to produce output", __FILE__, __LINE__);
		}
	}

	free(vFp);

	info("buffer complete successfully");
	return 0;
}
