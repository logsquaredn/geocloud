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

	const char *fc = getenv("ROTOTILLER_FILTER_COLUMN");
	if(fc == NULL) {
		fatalErrorWithCode("env var: ROTOTILLER_FILTER_COLUMN must be set", __FILE__, __LINE__, EX_CONFIG);		
	}
	sprintf(iMsg, "filter column: %s", fc);
	info(iMsg);

	const char *fv = getenv("ROTOTILLER_FILTER_VALUE");
	if(fv == NULL) {
		fatalErrorWithCode("env var: ROTOTILLER_FILTER_VALUE must be set", __FILE__, __LINE__, EX_CONFIG);		
	}
	sprintf(iMsg, "filter value: %s", fv);
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

	char afc[ONE_KB];
	sprintf(afc, "%s!='%s'", fc, fv);
	sprintf(iMsg, "attribute filter query: %s", afc);
	info(iMsg);
	if(OGR_L_SetAttributeFilter(iLay, afc) != OGRERR_NONE) {
		fatalError("failed to set attribute filter on input layer", __FILE__, __LINE__);
	}

	OGRFeatureH iFeat;
	while((iFeat = OGR_L_GetNextFeature(iLay)) != NULL) {		
		if(OGR_L_DeleteFeature(iLay, OGR_F_GetFID(iFeat)) != OGRERR_NONE) {
			fatalError("failed to delete feature from input", __FILE__, __LINE__);			
		}
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

	info("filter complete successfully");
	return 0;
}
