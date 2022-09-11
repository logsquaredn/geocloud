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

	char *attributes = getenv("ROTOTILLER_ATTRIBUTES");
	if(attributes == NULL) {
		fatalErrorWithCode("env var: ROTOTILLER_ATTRIBUTES must be set", __FILE__, __LINE__, EX_CONFIG);
	}
	const char delim[2] = ",";
	char *tok = strtok(attributes, delim);
	const char *attNames[ONE_KB];
	int aCt = 0;
	while(tok != NULL) {
		attNames[aCt] = tok;

		sprintf(iMsg, "attribute: %s", tok);
		info(iMsg);

		++aCt;
		tok = strtok(NULL, delim);
	}
	if(aCt < 1) {
		fatalErrorWithCode("at least one attribute required as input", __FILE__, __LINE__, EX_CONFIG);
	}

    char *polygonArg = getenv("ROTOTILLER_POLYGON");
	if(polygonArg == NULL) {
		fatalErrorWithCode("env var: ROTOTILLER_POLYGON must be set", __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "polygon: %s", polygonArg);
	info(iMsg);

	char *vFp = NULL;
	if(isZip(iFp)) {
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

	OGRGeometryH polygon = OGR_G_CreateGeometry(wkbPolygon);
	if(polygon == NULL) {
		fatalErrorWithCode("failed to create polygon geometry", __FILE__, __LINE__, EX_CONFIG);	
	}
	if(OGR_G_ImportFromWkt(polygon, &polygonArg) != OGRERR_NONE) {
		fatalErrorWithCode("failed to import polygon from WKT", __FILE__, __LINE__, EX_CONFIG);
	}

	char ofp[ONE_KB];
	sprintf(ofp, "%s%s", oDir, "/output.json");
	sprintf(iMsg, "output filepath: %s", ofp);
	info(iMsg); 
	FILE *fptr = fopen(ofp, "w");
	if(fptr == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to open output file: %s", ofp);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CANTCREAT);
	}
	if(fputs("{\"results\":[", fptr) == EOF) {
		fatalError("failed to write beginning results to output file", __FILE__, __LINE__);
	}

	int hits = 0;
	OGRFeatureH iFeat;
	while((iFeat = OGR_L_GetNextFeature(iLay)) != NULL) {
		OGRGeometryH iGeom = OGR_F_GetGeometryRef(iFeat);

		if(OGR_G_Intersects(iGeom, polygon)) {
			++hits;
			if(hits > 1) {
				if(fputs(",", fptr) == EOF) {
					fatalError("failed to write results to output file", __FILE__, __LINE__);
				}
			}
			if(fputs("{", fptr) == EOF) {
				fatalError("failed to write results to output file", __FILE__, __LINE__);
			}

			for(int i = 0; i < aCt; ++i) {
				int fIdx = OGR_F_GetFieldIndex(iFeat, attNames[i]);
				if(fIdx < 0) {
					sprintf(iMsg, "couldn't find attribute: %s", attNames[i]);
					info(iMsg); 
				} else {
					const char *aVal = OGR_F_GetFieldAsString(iFeat, fIdx);

					char result[ONE_KB];
					if(i + 1 == aCt) {
						sprintf(result, "\"%s\":\"%s\"", attNames[i], aVal);
					} else {
						sprintf(result, "\"%s\":\"%s\",", attNames[i], aVal);
					}

					if(fputs(result, fptr) == EOF) {
						fatalError("failed to write results to output file", __FILE__, __LINE__);
					}
				}
			}

			if(fputs("}", fptr) == EOF) {
				fatalError("failed to write results to output file", __FILE__, __LINE__);
			}
		}

		OGR_G_DestroyGeometry(iGeom);
	}

	if(fputs("]}", fptr) == EOF) {
		fatalError("failed to write end results to output file", __FILE__, __LINE__);
	}

	fclose(fptr);
	GDALClose(iDs);
	OGR_G_DestroyGeometry(polygon);
	free(vFp);

	info("polygon vector lookup complete successfully");
	return 0;
}
