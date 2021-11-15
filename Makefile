DOCKER ?= docker
DOCKER-COMPOSE ?= docker-compose
GCC ?= gcc
GO ?= go

BIN ?= bin
ASSETS ?= assets

REPOSITORY ?= logsquaredn/geocloud
TAG ?= $(REPOSITORY):latest

TASKS-TAR ?= runtime/tasks.tar
TASKS-REPOSITORY ?= $(REPOSITORY)
TASKS-TAGS ?= $(TASKS-REPOSITORY):task-removebadgeometry $(TASKS-REPOSITORY):task-buffer $(TASKS-REPOSITORY):task-filter $(TASKS-REPOSITORY):task-reproject
TASKS-DIR ?= tasks

GCC-ARGS ?= -Wall

.PHONY: build-tasks
build-tasks:
	$(DOCKER-COMPOSE) -f tasks/docker-compose.yml build

.PHONY: build-tasks-c
build-tasks-c:
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/buffer/buffer.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/buffer
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/filter/filter.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/filter
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/reproject/reproject.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/reproject
	$(GCC) $(GCC-ARGS) $(TASKS-DIR)/badGeometry/removeBadGeometry.c $(TASKS-DIR)/shared/shared.c -l gdal -o $(BIN)/removebadgeometry

.PHONY: push-tasks
push-tasks: build-tasks
	$(DOCKER-COMPOSE) -f tasks/docker-compose.yml push

.PHONY: save-tasks
save-tasks: build-tasks
	$(DOCKER) save -o $(TASKS-TAR) $(TASKS-TAGS)

.PHONY: datastore
datastore:
	$(DOCKER-COMPOSE) up -d datastore

.PHONY: objectstore
objectstore:
	$(DOCKER-COMPOSE) up -d objectstore

.PHONY: messagequeue
rabbitmq:
	$(DOCKER-COMPOSE) up -d messagequeue

.PHONY: services
services: datastore objectstore messagequeue

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
	$(DOCKER-COMPOSE) down --remove-orphans

CLEAN ?= hack/geocloud/* hack/minio/geocloud/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/* $(TASKS-TAR)

.PHONY: clean
clean: down
	rm -rf $(CLEAN)

.PHONY: prune
prune: clean
	$(DOCKER) system prune -a

.PHONY: test
test: save-tasks
	$(GO) test -test.v -race ./...

.PHONY: vet
vet: save-tasks
	$(GO) fmt ./...
	$(GO) vet ./...
