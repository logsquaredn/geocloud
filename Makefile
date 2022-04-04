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

.PHONY: infra
infra: save-tasks
	$(DOCKER-COMPOSE) up -d datastore objectstore messagequeue

.PHONY: build
build:
	$(DOCKER-COMPOSE) build

.PHONY: up
up: build
	$(DOCKER-COMPOSE) up migrate worker api

.PHONY: restart
restart: build
	$(DOCKER-COMPOSE) stop worker api
	$(DOCKER-COMPOSE) up worker api

.PHONY: scan
scan: build
	$(DOCKER) scan -f Dockerfile $(TAG)

.PHONY: push
push: build
	$(DOCKER-COMPOSE) push

.PHONY: down
down:
	$(DOCKER-COMPOSE) down --remove-orphans

CLEAN ?= hack/geocloud/* hack/minio/.minio.sys hack/minio/geocloud/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/* $(TASKS-TAR)

.PHONY: clean
clean: down
	rm -rf $(CLEAN)

.PHONY: prune
prune: clean
	$(DOCKER) system prune --volumes -a

.PHONY: test
test: save-tasks
	$(GO) test -test.v -race ./...

.PHONY: vet
vet: save-tasks
	$(GO) fmt ./...
	$(GO) vet ./...

LOCALHOST-DIR ?= hack/localhost
LOCALHOST-KEY ?= $(LOCALHOST-DIR)/localhost.key
LOCALHOST-PEM ?= $(LOCALHOST-DIR)/localhost.pem
LOCALHOST-CRT ?= $(LOCALHOST-DIR)/localhost.crt
SSL-CRTS ?= /etc/ssl/certs.pem

.PHONY: crt
crt:
	mkdir $(LOCALHOST-DIR) || true
	openssl req -x509 -nodes -new -sha256 -days 90 -keyout $(LOCALHOST-KEY) -out $(LOCALHOST-PEM) -subj "/C=US/CN=localhost"
	openssl x509 -outform pem -in $(LOCALHOST-PEM) -out $(LOCALHOST-CRT)

.PHONY: trust
trust:
	sudo cat $(LOCALHOST-CRT) >> $(SSL-CRTS)
# sudo security add-trusted-cert -d -r trustAsRoot -k /Library/Keychains/System.keychain $(LOCALHOST-CRT)
