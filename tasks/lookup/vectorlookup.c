#include <stdio.h>

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

    const char *lonArg = getenv("GEOCLOUD_LONGITUDE");
	if(lonArg == NULL) {
		fatalErrorWithCode("env var: GEOCLOUD_LONGITUDE must be set", __FILE__, __LINE__, EX_CONFIG);
	}
    double lon = strtod(lonArg, NULL);
    if(lon == 0 || lon > 180 || lon < -180) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "longitude must be a double between -180 & 180. got: %s", lonArg);
        fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
    }
	sprintf(iMsg, "lon: %f", lon);
	info(iMsg);

    const char *latArg = getenv("GEOCLOUD_LATITUDE");
	if(latArg == NULL) {
		fatalErrorWithCode("env var: GEOCLOUD_LATITUDE must be set", __FILE__, __LINE__, EX_CONFIG);
	}
    double lat = strtod(latArg, NULL);
    if(lat == 0 || lat > 90 || lat < -90) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "latitude must be a double between -90 & 90. got: %s", latArg);
        fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
    }
	sprintf(iMsg, "lat: %f", lat);
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

	OGRGeometryH point = OGR_G_CreateGeometry(wkbPoint);
	if(point == NULL) {
		fatalErrorWithCode("failed to create point geometry", __FILE__, __LINE__, EX_CONFIG);
	}
	OGR_G_AddPoint_2D(point, lon, lat);

	OGRFeatureH iFeat;
	while((iFeat = OGR_L_GetNextFeature(iLay)) != NULL) {
		OGRGeometryH iGeom = OGR_F_GetGeometryRef(iFeat);
		if(!OGR_G_Intersects(iGeom, point)) {
			if(OGR_L_DeleteFeature(iLay, OGR_F_GetFID(iFeat)) != OGRERR_NONE) {
				fatalError("failed to delete feature from input", __FILE__, __LINE__);			
			}
		}

		OGR_G_DestroyGeometry(iGeom);
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

	info("vector lookup complete successfully");
	return 0;
}
