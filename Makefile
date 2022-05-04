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

.PHONY: fallthrough
fallthrough: save-tasks infra up

.PHONY: build-tasks
build-tasks:
	@$(DOCKER-COMPOSE) -f $(TASKS-DIR)/docker-compose.yml build

.PHONY: push-tasks
push-tasks: build-tasks
	@$(DOCKER-COMPOSE) -f $(TASKS-DIR)/docker-compose.yml push

.PHONY: tasks save-tasks
tasks save-tasks: build-tasks
	@$(DOCKER) save -o $(TASKS-TAR) $(TASKS-TAGS)

.PHONY: services
services:
	@$(DOCKER-COMPOSE) up -d datastore objectstore messagequeue

.PHONY: migrate migrations
migrate migrations:
	@$(DOCKER-COMPOSE) up --build migrate

.PHONY: infra infrastructure
infra infrastructure: services sleep migrate

.PHONY: up
up:
	@$(DOCKER-COMPOSE) up --build worker api

.PHONY: restart
restart:
	@$(DOCKER-COMPOSE) stop worker api
	@$(DOCKER-COMPOSE) up --build worker api

.PHONY: build
build:
	@$(DOCKER-COMPOSE) build

.PHONY: scan
scan: build
	@$(DOCKER) scan -f Dockerfile $(TAG)

.PHONY: push
push: build
	@$(DOCKER-COMPOSE) push

.PHONY: down
down:
	@$(DOCKER-COMPOSE) down --remove-orphans

CLEAN ?= hack/geocloud/* hack/minio/.minio.sys hack/minio/geocloud/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/* $(TASKS-TAR)

.PHONY: clean
clean: down
	@rm -rf $(CLEAN)

.PHONY: prune
prune: clean
	@$(DOCKER) system prune --volumes -a

.PHONY: test
test: save-tasks
	@$(GO) test -test.v -race ./...

.PHONY: vet
vet: save-tasks
	@$(GO) fmt ./...
	@$(GO) vet ./...

.PHONY: sleep
sleep:
	@sleep 2
