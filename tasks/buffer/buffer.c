#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	GDALAllRegister();

	char iMsg[ONE_KB];

	char *iFp = getenv(ENV_VAR_INPUT_FILEPATH);
	if(iFp == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_INPUT_FILEPATH);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "input filepath: %s", iFp);
	info(iMsg);
	
	const char *oDir = getenv(ENV_VAR_OUTPUT_DIRECTORY);
	if(oDir == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_OUTPUT_DIRECTORY);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "output directory: %s", oDir);
	info(iMsg);

	const char *bdArg = getenv("ROTOTILLER_BUFFER_DISTANCE");
	if(bdArg == NULL) {
		fatalErrorWithCode("env var: ROTOTILLER_BUFFER_DISTANCE must be set", __FILE__, __LINE__, EX_CONFIG);		
	}
	double bd = strtod(bdArg, NULL);
	if(bd == 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "buffer distance must be a postitive double. got: %s", bdArg);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "buffer distance: %f", bd);
	info(iMsg);

	const char *qscArg = getenv("ROTOTILLER_QUADRANT_SEGMENT_COUNT");
	if(qscArg == NULL) {
		fatalErrorWithCode("env var: ROTOTILLER_QUADRANT_SEGMENT_COUNT must be set", __FILE__, __LINE__, EX_CONFIG);		
	}
	int qsc = atoi(qscArg);
	if(qsc <= 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "quadrant segment count must be a positive integer. got: %s", qscArg);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
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
			fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_NOINPUT);	
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
			fatalErrorWithCode("input zip must contain a shp file", __FILE__, __LINE__, EX_NOINPUT);
		}

		free(fl);
	} else if(isGeojson(iFp)) {
		vFp = strdup(iFp);
	} else {
		fatalErrorWithCode("input file must be a .zip or .json", __FILE__, __LINE__, EX_NOINPUT);
	}
	sprintf(iMsg, "vector filepath: %s", vFp);
	info(iMsg); 

	GDALDatasetH iDs = OGROpen(vFp, 1, NULL);
	if(iDs == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to open vector file: %s", vFp);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_DATAERR);	
	}

	int lCount = GDALDatasetGetLayerCount(iDs);
	if(lCount < 1) {
		fatalErrorWithCode("input dataset has no layers", __FILE__, __LINE__, EX_DATAERR);
	}
	OGRLayerH iLay = OGR_DS_GetLayer(iDs, 0);
	if(iLay == NULL) {
		fatalErrorWithCode("failed to get layer from input dataset", __FILE__, __LINE__, EX_DATAERR);
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
			 fatalErrorWithCode("failed to produce output", __FILE__, __LINE__, EX_CANTCREAT);
		 }
	} else {
		if(produceJsonOutput(iFp, oDir) != 0) {
			fatalErrorWithCode("failed to produce output", __FILE__, __LINE__, EX_CANTCREAT);
		}
	}

	free(vFp);

	info("buffer complete successfully");
	return 0;
}
