#include "shared.h"

const char *ENV_VAR_INPUT_FILEPATH = "GEOCLOUD_INPUT_FILE";
const char *ENV_VAR_OUTPUT_DIRECTORY = "GEOCLOUD_OUTPUT_DIR";
int MAX_UNZIPPED_FILES = 16;
int ONE_KB = 1024;


void info(const char *msg) {
    fprintf(stdout, "INFO:%lu: %s\n", time(NULL), msg);
}

void error(const char *msg, const char *file, int line) {
	fprintf(stderr, "ERROR:%lu:%s:%d: %s\n", time(NULL), file, line, msg);
}

void fatalError(const char *msg, const char *file, int line) {
    error(msg, file, line);
	exit(1);
}

int isGeojson(const char *fp) {
    char *ext = strrchr(fp, '.');
    if(ext != NULL && !strcmp(ext, ".json")) {
        return 1;
    }

    return 0;
}

int isShp(const char *fp) {
    char *ext = strrchr(fp, '.');
    if(ext != NULL && !strcmp(ext, ".shp")) {
        return 1;
    }

    return 0;
}

int isZip(const char *fp) {
    char *ext = strrchr(fp, '.');
    if(ext != NULL && !strcmp(ext, ".zip")) {
        return 1;
    }

    return 0;
}

char **unzip(const char *fp) {
    char dfp[ONE_KB]; 
    strcpy(dfp, fp);

    char* d = dirname(dfp);
    char cmd[ONE_KB];
    snprintf(cmd, sizeof(cmd), "%s%s%s%s", "unzip -o ", fp, " -d ", d);    

    FILE *fptr = popen(cmd, "r");
    if(fptr == NULL) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to execute command: %s", cmd);
        error(eMsg, __FILE__, __LINE__);
        return NULL;
    }

    const char delim[3] = ": ";
    char **fl = calloc(MAX_UNZIPPED_FILES, sizeof(char*));
    char buff[ONE_KB];
    int fc = 0;
    while(fgets(buff, ONE_KB, fptr) != NULL) {
        if(strstr(buff, "inflating:") != NULL) {
            char *tok = strtok(buff, delim);
            tok = strtok(NULL, delim);

            if(fc <= MAX_UNZIPPED_FILES) {
                fl[fc] = calloc(ONE_KB, sizeof(char*));
                fl[fc] = strdup(tok);
                ++fc;
            }
        }
    }

    if(pclose(fptr) == -1) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to close output pipe from command: %s", cmd);
        error(eMsg, __FILE__, __LINE__);
    }

    return fl;
}

GDALDatasetH createVectorDataset(const char *fp) {
    GDALDriverH shpD = GDALGetDriverByName("ESRI Shapefile");
	if(shpD == NULL) {
		error("failed to get shapefile driver", __FILE__, __LINE__);
        return NULL;
	}

    GDALDatasetH ds = GDALCreate(shpD, 
                          fp, 
                          0, 0, 0, 
                          GDT_Unknown, 
                          NULL);
    if(ds == NULL) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to create vector dataset at: %s", fp);
        fatalError(eMsg, __FILE__, __LINE__);
        return NULL;
    }
    
    return ds;
}

OGRGeometryH createTopLeftPoly(OGREnvelope* envelope) {
    OGRGeometryH topLeftRing =  OGR_G_CreateGeometry(wkbLinearRing);
    if(topLeftRing == NULL) {
        error("failed to create top left ring", __FILE__, __LINE__);
        return NULL;
    }
    OGR_G_AddPoint_2D(topLeftRing, envelope->MinX, envelope->MaxY);
    OGR_G_AddPoint_2D(topLeftRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY);
    OGR_G_AddPoint_2D(topLeftRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(topLeftRing, envelope->MinX, envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(topLeftRing, envelope->MinX, envelope->MaxY);

    OGRGeometryH topLeftPoly = OGR_G_CreateGeometry(wkbPolygon);
    if(topLeftPoly == NULL) {
        error("failed to create top left poly", __FILE__, __LINE__);
        return NULL;
    }
    if(OGR_G_AddGeometry(topLeftPoly, topLeftRing) != OGRERR_NONE) {
        error("failed to add top left ring to top left poly", __FILE__, __LINE__);
        return NULL;
    }

    OGR_G_DestroyGeometry(topLeftRing);

    return topLeftPoly;
}

OGRGeometryH createTopRightPoly(OGREnvelope* envelope) {
    OGRGeometryH topRightRing =  OGR_G_CreateGeometry(wkbLinearRing);
    if(topRightRing == NULL) {
        error("failed to create top right ring", __FILE__, __LINE__);
        return NULL;
    }
    OGR_G_AddPoint_2D(topRightRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY);
    OGR_G_AddPoint_2D(topRightRing, envelope->MaxX, envelope->MaxY);
    OGR_G_AddPoint_2D(topRightRing, envelope->MaxX, envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(topRightRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(topRightRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY);

    OGRGeometryH topRightPoly = OGR_G_CreateGeometry(wkbPolygon);
    if(topRightPoly == NULL) {
        error("failed to create top right poly", __FILE__, __LINE__);
        return NULL;
    }
    if(OGR_G_AddGeometry(topRightPoly, topRightRing) != OGRERR_NONE) {
        error("failed to add top right ring to top right poly", __FILE__, __LINE__);
        return NULL;
    }

    OGR_G_DestroyGeometry(topRightRing);

    return topRightPoly;
}

OGRGeometryH createBottomRightPoly(OGREnvelope* envelope) {
    OGRGeometryH bottomRightRing =  OGR_G_CreateGeometry(wkbLinearRing);
    if(bottomRightRing == NULL) {
        error("failed to create bottom right ring", __FILE__, __LINE__);
        return NULL;
    }
    OGR_G_AddPoint_2D(bottomRightRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(bottomRightRing, envelope->MaxX, envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(bottomRightRing, envelope->MaxX, envelope->MinY);
    OGR_G_AddPoint_2D(bottomRightRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MinY);
    OGR_G_AddPoint_2D(bottomRightRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));

    OGRGeometryH buttomRightPoly = OGR_G_CreateGeometry(wkbPolygon);
    if(buttomRightPoly == NULL) {
        error("failed to create bottom right poly", __FILE__, __LINE__);
        return NULL;
    }
    if(OGR_G_AddGeometry(buttomRightPoly, bottomRightRing) != OGRERR_NONE) {
        error("failed to add bottom right ring to bottom right poly", __FILE__, __LINE__);
        return NULL;
    }

    OGR_G_DestroyGeometry(bottomRightRing);

    return buttomRightPoly;
}

OGRGeometryH createBottomLeftPoly(OGREnvelope* envelope) {
    OGRGeometryH bottomLeftRing =  OGR_G_CreateGeometry(wkbLinearRing);
    if(bottomLeftRing == NULL) {
        error("failed to create bottom left ring", __FILE__, __LINE__);
        return NULL;
    }
    OGR_G_AddPoint_2D(bottomLeftRing, envelope->MinX, envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(bottomLeftRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));
    OGR_G_AddPoint_2D(bottomLeftRing, envelope->MaxX - ((envelope->MaxX - envelope->MinX) / 2), envelope->MinY);
    OGR_G_AddPoint_2D(bottomLeftRing, envelope->MinX, envelope->MinY);
    OGR_G_AddPoint_2D(bottomLeftRing, envelope->MinX, envelope->MaxY - ((envelope->MaxY - envelope->MinY) / 2.0));

    OGRGeometryH buttomLeftPoly = OGR_G_CreateGeometry(wkbPolygon);
    if(buttomLeftPoly == NULL) {
        error("failed to create bottom left poly", __FILE__, __LINE__);
        return NULL;
    }
    if(OGR_G_AddGeometry(buttomLeftPoly, bottomLeftRing) != OGRERR_NONE) {
        error("failed to add bottom left ring to bottom left poly", __FILE__, __LINE__);
        return NULL;
    }

    OGR_G_DestroyGeometry(bottomLeftRing);

    return buttomLeftPoly;
}

int splitGeometries(OGRGeometryH splitGeoms[], OGRGeometryH inputGeometry, int seed) {
    if(OGR_G_GetGeometryCount(inputGeometry) < 50) {
        splitGeoms[seed] = inputGeometry; 
        return seed + 1;
    }

    OGREnvelope envelope;
    OGR_G_GetEnvelope(inputGeometry, &envelope);

    OGRGeometryH topLeft = createTopLeftPoly(&envelope);
    OGRGeometryH topRight = createTopRightPoly(&envelope);
    OGRGeometryH bottomRight = createBottomRightPoly(&envelope);
    OGRGeometryH bottomLeft = createBottomLeftPoly(&envelope);

    OGRGeometryH topLeftIntersection = OGR_G_Intersection(inputGeometry, topLeft);
    if(topLeftIntersection == NULL) {
        error("failed to intersect input geometry with top left geometry", __FILE__, __LINE__);
        return -1;
    }

    OGRGeometryH topRightIntersection = OGR_G_Intersection(inputGeometry, topRight);
    if(topRightIntersection == NULL) {
        error("failed to intersect input geometry with top right geometry", __FILE__, __LINE__);
        return -1;
    }

    OGRGeometryH bottomRightIntersection = OGR_G_Intersection(inputGeometry, bottomRight);
    if(bottomRightIntersection == NULL) {
        error("failed to intersect input geometry with bottom right geometry", __FILE__, __LINE__);
        return -1;
    }

    OGRGeometryH bottomLeftIntersection = OGR_G_Intersection(inputGeometry, bottomLeft);
    if(bottomLeftIntersection == NULL) {
        error("failed to intersect input geometry with bottom left geometry", __FILE__, __LINE__);
        return -1;
    }

    seed = splitGeometries(splitGeoms, topLeftIntersection, seed);
    seed = splitGeometries(splitGeoms, topRightIntersection, seed);
    seed = splitGeometries(splitGeoms, bottomRightIntersection, seed);
    seed = splitGeometries(splitGeoms, bottomLeftIntersection, seed);

    return seed;
}

int zipDir(const char *dst, const char *srcDir) {
    char cmd[ONE_KB];
    sprintf(cmd, "zip -jr %s %s", dst, srcDir);

    FILE *fptr = popen(cmd, "r");
    if(fptr == NULL) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to zip directory: %s", srcDir);
        error(eMsg, __FILE__, __LINE__);
        return 1;  
    }

    char buff[ONE_KB];
    while(fgets(buff, ONE_KB, fptr) != NULL) {}

    if(pclose(fptr) == -1) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to close output pipe from command: %s", cmd);
        error(eMsg, __FILE__, __LINE__);
    }

    return 0;
}

int dump2geojson(const char *dst, const char *src) {
    char cmd[ONE_KB];
    sprintf(cmd, "ogr2ogr -f GeoJSON %s %s", dst, src);

    FILE *fptr = popen(cmd, "r");
    if(fptr == NULL) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to convert src: %s to dst: %s", src, dst);
        error(eMsg, __FILE__, __LINE__);
        return 1;  
    }

    char buff[ONE_KB];
    while(fgets(buff, ONE_KB, fptr) != NULL) {}

    if(pclose(fptr) == -1) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to close output pipe from command: %s", cmd);
        error(eMsg, __FILE__, __LINE__);
    }

    return 0;
}

int dump2shp(const char *dst, const char *src) {
    char cmd[ONE_KB];
    sprintf(cmd, "ogr2ogr -f \"ESRI Shapefile\" %s %s", dst, src);

    FILE *fptr = popen(cmd, "r");
    if(fptr == NULL) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to convert src: %s to dst: %s", src, dst);
        error(eMsg, __FILE__, __LINE__);
        return 1;  
    }

    char buff[ONE_KB];
    while(fgets(buff, ONE_KB, fptr) != NULL) {}

    if(pclose(fptr) == -1) {
        char eMsg[ONE_KB];
        sprintf(eMsg, "failed to close output pipe from command: %s", cmd);
        error(eMsg, __FILE__, __LINE__);
    }

    return 0;
}

int produceShpOutput(char *iFp, const char *oDir, const char *vFp) {
    char iMsg[ONE_KB];

    if(remove(iFp) != 0) {
		char eMsg[ONE_KB];
        sprintf(eMsg, "failed to cleanup input file: %s", iFp);
        error(eMsg, __FILE__, __LINE__);
		return 1;			\
	}
	
    const char *iDir = dirname(iFp);

	char zPath[ONE_KB];
	sprintf(zPath, "%s/output.zip", oDir);
	if(zipDir(zPath, iDir) != 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to zip directory: %s", iDir);
        error(eMsg, __FILE__, __LINE__);
		return 1;	
	}
	sprintf(iMsg, "output zip: %s", zPath);
	info(iMsg); 

	char gPath[ONE_KB];
	sprintf(gPath, "%s/output.json", oDir);
	if(dump2geojson(gPath, vFp)) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to convert shp: %s to geojson: %s", vFp, gPath);
        error(eMsg, __FILE__, __LINE__);
		return 1;
	}
	sprintf(iMsg, "output json: %s", gPath);
	info(iMsg); 

    return 0;
}

int produceJsonOutput(char *iFp, const char *oDir) {
    char iMsg[ONE_KB];

    char iFpTmp[ONE_KB];
    strcpy(iFpTmp, iFp);
	const char *iDir = dirname(iFpTmp);

	char sPath[ONE_KB];
	sprintf(sPath, "%s/output.shp", iDir);
	if(dump2shp(sPath, iFp)) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to convert geojson: %s to shp: %s", iFp, sPath);
        error(eMsg, __FILE__, __LINE__);
		return 1;		
	}
	char gPath[ONE_KB];
	sprintf(gPath, "%s/output.json", oDir);
	if(rename(iFp, gPath) != 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to move geojson: %s to output: %s", iFp, gPath);
        error(eMsg, __FILE__, __LINE__);
		return 1;	
	}
	sprintf(iMsg, "output json: %s", gPath);
	info(iMsg); 

	char zPath[ONE_KB];
	sprintf(zPath, "%s/output.zip", oDir);
	if(zipDir(zPath, iDir) != 0) {
		char eMsg[ONE_KB];
		sprintf(eMsg, "failed to zip directory: %s", iDir);
		error(eMsg, __FILE__, __LINE__);
        return 1;	
	}
	sprintf(iMsg, "output zip: %s", zPath);
	info(iMsg); 

    return 0;
}
