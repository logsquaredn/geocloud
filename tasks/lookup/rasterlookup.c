#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 5) {
		error("raster lookup requires four arguments. Input file, output directory, longitude, and latitude", __FILE__, __LINE__);
		fatalError();
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputDir = argv[2];
	fprintf(stdout, "output directory: %s\n", outputDir);

    const char* lonArg = argv[3];
    double lon = strtod(lonArg, NULL);
    if(lon == 0 || lon > 180 || lon < -180) {
        error("longitude must be a valid double between -180 & 180", __FILE__, __LINE__);
		fatalError();
    }

    const char* latArg = argv[4];
    double lat = strtod(latArg, NULL);
    if(lat == 0 || lat > 90 || lat < -90) {
        error("latitude must be a valid double between -90 & 90", __FILE__, __LINE__);
		fatalError();
    }

	char *inputGeoFilePath = getInputGeoFilePath(inputFilePath);
	if(inputGeoFilePath == NULL) {
		error("failed to find input geo file path", __FILE__, __LINE__);
		fatalError();
	}
	fprintf(stdout, "input geo file path: %s\n", inputGeoFilePath);

	struct GDALHandles gdalHandles;
	gdalHandles.inputLayer = NULL;
	if(rasterInitialize(&gdalHandles, inputGeoFilePath, outputDir)) {
		error("failed to initialize", __FILE__, __LINE__);
		fatalError();
	}
	free(inputGeoFilePath);

	GDALClose(gdalHandles.inputDataset);

	fprintf(stdout, "raster lookup completed successfully\n");
	return 0;
}
