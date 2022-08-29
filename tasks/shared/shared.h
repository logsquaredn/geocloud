#ifndef SHARED_H
#define SHARED_H

#include "gdal.h"
#include <ogr_srs_api.h>
#include <stdlib.h>
#include <libgen.h>
#include <dirent.h>
#include <sysexits.h>

extern const char *ENV_VAR_INPUT_FILEPATH;
extern const char *ENV_VAR_OUTPUT_DIRECTORY;
extern int MAX_UNZIPPED_FILES;
extern int ONE_KB;

void info(const char*);
void error(const char*, const char*, int);
void fatalError(const char*, const char*, int);
void fatalErrorWithCode(const char *msg, const char *file, int line, int code);

int isGeojson(const char*);
int isZip(const char*);
int isShp(const char*);
// result of unzip() must be free()'d
char **unzip(const char*);

int splitGeometries(OGRGeometryH[], OGRGeometryH, int);

int zipDir(const char*, const char*);

int produceShpOutput(char *, const char *, const char *);
int produceJsonOutput(char *, const char *);

#endif
