#include <stdio.h>

#include "../shared/shared.h"

int main(int argc, char *argv[]) {
	if(argc != 5) {
		error("vector lookup requires four arguments. Input file, output directory, longitude, and latitude", __FILE__, __LINE__);
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
	if(vectorInitialize(&gdalHandles, inputGeoFilePath, outputDir)) {
		error("failed to initialize", __FILE__, __LINE__);
		fatalError();
	}
	free(inputGeoFilePath);

	if(gdalHandles.inputLayer != NULL) {
        OGRGeometryH point = OGR_G_CreateGeometry(wkbPoint);
        if(point == NULL) {
            error("failed to create point geometry", __FILE__, __LINE__);
            fatalError();
        }
        OGR_G_AddPoint_2D(point, lon, lat);

		OGRFeatureH inputFeature;
		while((inputFeature = OGR_L_GetNextFeature(gdalHandles.inputLayer)) != NULL) {
            OGRGeometryH inputGeometry = OGR_F_GetGeometryRef(inputFeature);
            if(OGR_G_Intersects(inputGeometry, point)) {
                OGRGeometryH intersection = OGR_G_Intersection(inputGeometry, point);
                if(intersection != NULL) {
                    if(buildOutputVectorFeature(&gdalHandles, &intersection, &inputFeature)) {
                        error("failed to build output vector feature", __FILE__, __LINE__);
                        fatalError();     
                    }
                }

                OGR_G_DestroyGeometry(intersection);
            }

			OGR_G_DestroyGeometry(inputGeometry);
		}

		OGR_F_Destroy(inputFeature);
		OSRDestroySpatialReference(gdalHandles.inputSpatialRef);
		GDALClose(gdalHandles.outputDataset);
	} else {
		fprintf(stdout, "no layers found in input file\n");
	}

	// TODO this seg faults on some geojson input
	// GDALClose(gdalHandles.inputDataset);

	if(zipShp(outputDir)) {
		error("failed to zip up shp", __FILE__, __LINE__);
		fatalError();
	}

	if(dumpToGeojson(outputDir)) {
		error("failed to convert shp to geojson", __FILE__, __LINE__);
		fatalError();
	}

	if(cleanup(outputDir)) {
		error("failed to cleanup output", __FILE__, __LINE__);
		fatalError();
	}

	fprintf(stdout, "vector lookup completed successfully\n");
	return 0;
}
