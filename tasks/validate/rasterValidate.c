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
	GDALClose(gd);

	info("raster lookup completed successfully");
	return 0;
}
