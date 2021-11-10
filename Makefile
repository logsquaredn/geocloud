DOCKER ?= docker
DOCKER-COMPOSE ?= docker-compose
GCC ?= gcc
GO ?= go
TERRAFORM ?= terraform

BIN ?= bin
ASSETS ?= assets

REPOSITORY ?= logsquaredn/geocloud
TAG ?= $(REPOSITORY):latest

TASKS-TAR ?= runtime/tasks.tar
TASKS-REPOSITORY ?= $(REPOSITORY)
TASKS-TAGS ?= $(TASKS-REPOSITORY):task-remove-bad-geometry $(TASKS-REPOSITORY):task-buffer $(TASKS-REPOSITORY):task-filter $(TASKS-REPOSITORY):task-reproject
TASKS-DIR ?= tasks

GCC-ARGS ?= -Wall

.PHONY: build-tasks
build-tasks:
	$(DOCKER-COMPOSE) -f tasks/docker-compose.yml build

.PHONY: build-tasks-c
build-tasks-c:
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/buffer/buffer.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/buffer
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/filter/filter.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/filter
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/reproject/reproject.c $(TASKS-DIR)/hared/shared.c -l gdal -o $(BIN)/reproject
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/badGeometry/removeBadGeometry.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/removeBadGeometry

.PHONY: push-tasks
push-tasks: build-tasks
	$(DOCKER-COMPOSE) -f tasks/docker-compose.yml push

.PHONY: save-tasks
save-tasks: build-tasks
	$(DOCKER) save -o $(TASKS-TAR) $(TASKS-TAGS)

.PHONY: postgres
postgres:
	$(DOCKER-COMPOSE) up -d postgres

.PHONY: minio
minio:
	$(DOCKER-COMPOSE) up -d minio

.PHONY: services
services: postgres minio

.PHONY: build
build: save-tasks
	$(DOCKER-COMPOSE) build

.PHONY: up
up: build services
	$(DOCKER-COMPOSE) up migrate worker api

.PHONY: restart
restart:
	$(DOCKER-COMPOSE) restart

.PHONY: scan
scan: build
	$(DOCKER) scan -f Dockerfile $(TAG)

.PHONY: push
push: build
	$(DOCKER-COMPOSE) push

.PHONY: down
down:
	$(DOCKER-COMPOSE) down

CLEAN ?= hack/geocloud/* hack/minio/geocloud/* $(TASKS-TAR)

.PHONY: clean
clean: down
	rm -rf $(CLEAN)

.PHONY: prune
prune: clean
	$(DOCKER) system prune -a

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: terraform
terraform:
	$(TERRAFORM) -chdir=infrastructure/tf/ init

.PHONY: infrastructure
infrastructure: terraform
	$(TERRAFORM) -chdir=infrastructure/tf/ apply
