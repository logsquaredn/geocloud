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
		OGRGeometryH iGeom = OGR_F_GetGeometryRef(iFeat);
		if(!OGR_G_IsValid(iGeom)) {	
			if(OGR_L_DeleteFeature(iLay, OGR_F_GetFID(iFeat)) != OGRERR_NONE) {
				fatalError("failed to delete feature from input", __FILE__, __LINE__);			
			}	
		}

		OGR_G_DestroyGeometry(iGeom);
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

	info("remove bad geometry completed successfully");
	return 0;
}
