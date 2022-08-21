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

	const char *tpArg = getenv("ROTOTILLER_TARGET_PROJECTION");	
	long tp = strtol(tpArg, NULL, 10);
	if(tp <= 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "EPSG code must be a positive integer. got: %s", tpArg);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "target projection: %ld", tp);
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

	// Create coordinate transformer
	OGRSpatialReferenceH iSr = OGR_L_GetSpatialRef(iLay);
	if(iSr == NULL) {
		fatalErrorWithCode("failed to get input spatial reference", __FILE__, __LINE__, EX_DATAERR);
	}
	OGRSpatialReferenceH oSr = OSRNewSpatialReference("");
	if(OSRImportFromEPSG(oSr, tp) != OGRERR_NONE) {
		fatalError("failed to create output spatial reference", __FILE__, __LINE__);
	}
	OGRCoordinateTransformationH tf = OCTNewCoordinateTransformation(iSr, oSr);

	// Create output dataset
	OGRSFDriverH driver = OGRGetDriverByName("ESRI Shapefile");
	if(driver == NULL) {
		fatalError("failed to get shapefile driver", __FILE__, __LINE__);
	}

	char *dupIFp = strdup(iFp);
	char *iDir = dirname(dupIFp);
	char wDir[ONE_KB];
	sprintf(wDir, "%s/work", iDir);
	if(mkdir(wDir, 0755) != 0) {
		fatalError("failed to create work directory", __FILE__, __LINE__);
	}

	const char *lName = OGR_L_GetName(iLay);
	char oDsPath[ONE_KB];
	sprintf(oDsPath, "%s/%s.shp", wDir, lName);
	sprintf(iMsg, "work filepath: %s", oDsPath);
	info(iMsg); 

	OGRwkbGeometryType iGeomType = OGR_L_GetGeomType(iLay);
	GDALDatasetH oDs = GDALCreate(driver, oDsPath, 0, 0, 0, iGeomType, NULL);
	if(oDs == NULL) {
		fatalError("failed to create output dataset", __FILE__, __LINE__);
	}

	OGRLayerH oLay = GDALDatasetCreateLayer(oDs, lName, oSr, iGeomType, NULL);
	if(oLay == NULL) {
		fatalError("failed to create output layer", __FILE__, __LINE__);
	}

	OGRFeatureDefnH iLayDefn = OGR_L_GetLayerDefn(iLay);
	int iFieldCount = OGR_FD_GetFieldCount(iLayDefn);
	for(int i = 0; i < iFieldCount; ++i) {
		OGRFieldDefnH fDef = OGR_FD_GetFieldDefn(iLayDefn, i);
		if(fDef == NULL) {
			fatalError("failed to get input field definition", __FILE__, __LINE__);
		}

		if(OGR_L_CreateField(oLay, fDef, i) != OGRERR_NONE) {
			fatalError("failed to create output field definition", __FILE__, __LINE__);
		}
	}

	OGRFeatureH iFeat;
	while((iFeat = OGR_L_GetNextFeature(iLay)) != NULL) {
		OGRGeometryH iGeom = OGR_F_GetGeometryRef(iFeat);
		
		if(OGR_G_Transform(iGeom, tf) != OGRERR_NONE) {
			fatalError("failed to transform geometry", __FILE__, __LINE__);
		}

		
		if(OGR_F_SetGeometry(iFeat, iGeom) != OGRERR_NONE) {
			fatalError("failed set updated geometry on input feature", __FILE__, __LINE__);
		}	

		if(OGR_L_CreateFeature(oLay, iFeat) != OGRERR_NONE) {
			fatalError("failed create feature in output layer", __FILE__, __LINE__);			
		}
	}
	
	OSRDestroySpatialReference(oSr);
	OSRDestroySpatialReference(iSr);
	OCTDestroyCoordinateTransformation(tf);
	GDALClose(iDs);
	GDALClose(oDs);

	if(produceShpOutput(iFp, oDir, oDsPath) != 0) {
		fatalErrorWithCode("failed to produce output", __FILE__, __LINE__, EX_CANTCREAT);
	}
	
	free(vFp);
	free(dupIFp);

	info("reproject complete successfully");
	return 0;
}
