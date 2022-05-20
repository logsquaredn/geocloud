#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 6) {
		error("raster lookup requires four arguments. Input file, output directory, list of bands, longitude, and latitude", __FILE__, __LINE__);
		fatalError();
	}

	const char *inputFilePath = argv[1];
	fprintf(stdout, "input filepath: %s\n", inputFilePath);
	
	const char *outputDir = argv[2];
	fprintf(stdout, "output directory: %s\n", outputDir);

	const char delim[2] = ",";
	char *tok = strtok(argv[3], delim);
	int bNums[32];
	int bCt = 0;
	while(tok != NULL) {
		int bNum = atoi(tok);
		if(bNum <= 0) {
			error("band must be an integer greater than 0", __FILE__, __LINE__);
			fatalError();
		}
		fprintf(stdout, "band: %d\n", bNum);
		bNums[bCt] = bNum;

		++bCt;
		tok = strtok(NULL, delim);
	}

    const char *lonArg = argv[4];
    double lon = strtod(lonArg, NULL);
    if(lon == 0 || lon > 180 || lon < -180) {
        error("longitude must be a double between -180 & 180", __FILE__, __LINE__);
		fatalError();
    }
	fprintf(stdout, "lon: %f\n", lon);

    const char *latArg = argv[5];
    double lat = strtod(latArg, NULL);
    if(lat == 0 || lat > 90 || lat < -90) {
        error("latitude must be a double between -90 & 90", __FILE__, __LINE__);
		fatalError();
    }
	fprintf(stdout, "lat: %f\n", lat);

	char *inputGeoFilePath = getInputGeoFilePath(inputFilePath);
	if(inputGeoFilePath == NULL) {
		error("failed to find input geo filepath", __FILE__, __LINE__);
		fatalError();
	}
	fprintf(stdout, "input geo filepath: %s\n", inputGeoFilePath);

	struct GDALHandles gdalHandles;
	gdalHandles.inputLayer = NULL;
	if(rasterInitialize(&gdalHandles, inputGeoFilePath, outputDir)) {
		error("failed to initialize", __FILE__, __LINE__);
		fatalError();
	}
	free(inputGeoFilePath);


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

	char *opath = getOutputFilePath(outputDir, "/output.json");
	fprintf(stdout, "output filepath: %s\n", opath);
	FILE *fptr = fopen(opath, "w");
	if(fputs("{\"results\":[", fptr) == EOF) {
		error("failed to write beginning results to output file", __FILE__, __LINE__);
		fatalError();
	}

	--bCt;
	while(bCt >= 0) {
		GDALRasterBandH band = GDALGetRasterBand(gdalHandles.inputDataset, bNums[bCt]);
		if(band == NULL) {
			error("failed to find band", __FILE__, __LINE__);
			fatalError();
		}
		
		float* rasterIOBuff = malloc(1 * sizeof(float));
		if(GDALRasterIO(band, GF_Read, col, row, 1, 1, rasterIOBuff, 1, 1, GDT_Float32, 0, 0) != OGRERR_NONE) {
			error("failed to read raster value", __FILE__, __LINE__);
			fatalError();
		}

		char result[128];
		if(bCt - 1 < 0) {
			sprintf(result, "{\"band%d\": %f}", bNums[bCt], rasterIOBuff[0]);	
		} else {
			sprintf(result, "{\"band%d\": %f},", bNums[bCt], rasterIOBuff[0]);	
		}	

		if(fputs(result, fptr) == EOF) {
			error("failed to write results to output file", __FILE__, __LINE__);
			fatalError();
		}

		free(rasterIOBuff);
		--bCt;
	}

	if(fputs("]}", fptr) == EOF) {
		error("failed to write end results to output file", __FILE__, __LINE__);
		fatalError();
	}

	fclose(fptr);
	GDALClose(gdalHandles.inputDataset);

	fprintf(stdout, "raster lookup completed successfully\n");
	return 0;
}
