#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 6) {
		error("raster lookup requires four arguments. Input file, output directory, band, longitude, and latitude", __FILE__, __LINE__);
		fatalError();
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input file path: %s\n", inputFilePath);
	
	const char *outputDir = argv[2];
	fprintf(stdout, "output directory: %s\n", outputDir);

	const char *bandArg = argv[3];
	int bandNumber = atoi(bandArg);
	if(bandNumber <= 0) {
		error("band must be a valid integer greater than 0", __FILE__, __LINE__);
		fatalError();
	}
	fprintf(stdout, "band: %d\n", bandNumber);

    const char *lonArg = argv[4];
    double lon = strtod(lonArg, NULL);
    if(lon == 0 || lon > 180 || lon < -180) {
        error("longitude must be a valid double between -180 & 180", __FILE__, __LINE__);
		fatalError();
    }
	fprintf(stdout, "lon: %f\n", lon);

    const char *latArg = argv[5];
    double lat = strtod(latArg, NULL);
    if(lat == 0 || lat > 90 || lat < -90) {
        error("latitude must be a valid double between -90 & 90", __FILE__, __LINE__);
		fatalError();
    }
	fprintf(stdout, "lat: %f\n", lat);

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

	GDALRasterBandH band = GDALGetRasterBand(gdalHandles.inputDataset, bandNumber);
	if(band == NULL) {
		error("failed to find band", __FILE__, __LINE__);
		fatalError();
	}

	double *buff = (double*) malloc(6 * sizeof(double));
	if(GDALGetGeoTransform(gdalHandles.inputDataset, buff) != OGRERR_NONE) {
		error("failed to get geotransform", __FILE__, __LINE__);
		fatalError();
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
	float* rasterIOBuff = malloc(1 * sizeof(float));
	if(GDALRasterIO(band, GF_Read, col, row, 1, 1, rasterIOBuff, 1, 1, GDT_Float32, 0, 0) != OGRERR_NONE) {
		error("failed to read raster value", __FILE__, __LINE__);
		fatalError();
	}

	char result[128];
	sprintf(result, "{\"band%d\": %f}", bandNumber, rasterIOBuff[0]);
	
	char *opath = getOutputFilePath(outputDir, "/output.json");
	fprintf(stdout, "output filepath: %s\n", opath);
	FILE *fptr = fopen(opath, "w");
	if(fptr == NULL) {
		error("failed to create output file", __FILE__, __LINE__);
		fatalError();
	}

	if(fputs(result, fptr) == EOF) {
		error("failed to write results to output file", __FILE__, __LINE__);
		fatalError();
	}

	fclose(fptr);
	free(rasterIOBuff);
	GDALClose(gdalHandles.inputDataset);

	fprintf(stdout, "raster lookup completed successfully\n");
	return 0;
}
