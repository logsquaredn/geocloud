#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	GDALAllRegister();

	char iMsg[ONE_KB];

	const char *ifp = getenv(ENV_VAR_INPUT_FILEPATH);
	if(ifp == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_INPUT_FILEPATH);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "input filepath: %s", ifp);
	info(iMsg);
	
	const char *od = getenv(ENV_VAR_OUTPUT_DIRECTORY);
	if(od == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "env var: %s must be set", ENV_VAR_OUTPUT_DIRECTORY);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
	}
	sprintf(iMsg, "output directory: %s", od);
	info(iMsg);

	char *bands = getenv("GEOCLOUD_BANDS");
	if(bands == NULL) {
		fatalErrorWithCode("env var: GEOCLOUD_BANDS must be set", __FILE__, __LINE__, EX_CONFIG);
	}
	const char delim[2] = ",";
	char *tok = strtok(bands, delim);
	int bNums[ONE_KB];
	int bCt = 0;
	while(tok != NULL) {
		int bNum = atoi(tok);
		if(bNum <= 0) {
			char eMsg[ONE_KB];
			sprintf(eMsg, "invalid band: %s. must be an interger greater than zero", tok);
			fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_CONFIG);
		}
		bNums[bCt] = bNum;

		sprintf(iMsg, "band: %d", bNum);
		info(iMsg);

		++bCt;
		tok = strtok(NULL, delim);
	}
	if(bCt < 1) {
		fatalErrorWithCode("at least one band required as input", __FILE__, __LINE__, EX_CONFIG);
	}

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

	if(!isZip(ifp)) {
		fatalErrorWithCode("input file must be a .zip", __FILE__, __LINE__, EX_NOINPUT);
	}

	char **fl = unzip(ifp);
	if(fl == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to unzip: %s", ifp);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_NOINPUT);
	}

	char *f;
	int fc = 0;
	while((f = fl[fc]) != NULL) {
		++fc;
	}
	if(fc != 1) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "input zip must contain exactly one raster file. found: %d", fc);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_NOINPUT);
	}

	char *rfp = fl[0];
	sprintf(iMsg, "raster filepath: %s", rfp);
	info(iMsg); 

	GDALDatasetH gd = GDALOpen(rfp, GA_ReadOnly);
	if(gd == NULL) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to open raster file: %s", rfp);
		fatalErrorWithCode(eMsg, __FILE__, __LINE__, EX_DATAERR);
	}
	free(fl);

	double *buff = (double*) calloc(6, sizeof(double));
	if(GDALGetGeoTransform(gd, buff) != OGRERR_NONE) {
		fatalErrorWithCode("failed to get geotransform", __FILE__, __LINE__, EX_DATAERR);
	}
	double xOrigin = buff[0];
	double yOrigin = buff[3];
	double pixWidth = buff[1];
	double pixHeight = buff[5];
	free(buff);
	if(pixHeight < 0) {
		pixHeight = pixHeight * -1;
	}
	int col = (lon - xOrigin) / pixWidth;
	int row = (yOrigin - lat) / pixHeight;

    char ofp[ONE_KB];
    sprintf(ofp, "%s%s", od, "/output.json");
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

	--bCt;
	while(bCt >= 0) {
		GDALRasterBandH band = GDALGetRasterBand(gd, bNums[bCt]);
		if(band == NULL) {
			char eMsg[ONE_KB];
			sprintf(eMsg, "failed to find band: %d", bNums[bCt]);
			fatalError(eMsg, __FILE__, __LINE__);
		}
		
		float* rasterIOBuff = calloc(1, sizeof(float));
		if(GDALRasterIO(band, GF_Read, col, row, 1, 1, rasterIOBuff, 1, 1, GDT_Float32, 0, 0) != OGRERR_NONE) {
			fatalError("failed to read raster value", __FILE__, __LINE__);
		}

		char result[ONE_KB];
		if(bCt - 1 < 0) {
			sprintf(result, "{\"band%d\": %f}", bNums[bCt], rasterIOBuff[0]);	
		} else {
			sprintf(result, "{\"band%d\": %f},", bNums[bCt], rasterIOBuff[0]);	
		}	

		if(fputs(result, fptr) == EOF) {
			fatalError("failed to write results to output file", __FILE__, __LINE__);
		}

		free(rasterIOBuff);
		--bCt;
	}

	if(fputs("]}", fptr) == EOF) {
		fatalError("failed to write end results to output file", __FILE__, __LINE__);
	}

	fclose(fptr);
	GDALClose(gd);

	info("raster lookup completed successfully");
	return 0;
}
