DOCKER ?= docker
DOCKER-COMPOSE ?= docker-compose
GCC ?= gcc
GO ?= go

BIN ?= bin
ASSETS ?= assets

REPOSITORY ?= logsquaredn/geocloud
TAG ?= $(REPOSITORY):latest

.PHONY: fallthrough
fallthrough: fmt infra up

.PHONY: fmt
fmt:
	@$(GO) fmt ./...

.PHONY: services
services:
	@$(DOCKER-COMPOSE) up -d datastore objectstore messagequeue

.PHONY: build
build:
	@$(DOCKER-COMPOSE) build

.PHONY: secretary
secretary:
	$(DOCKER-COMPOSE) up --build secretary

.PHONY: migrate migrations
migrate migrations:
	@$(DOCKER-COMPOSE) up --build migrate

.PHONY: infra infrastructure
infra infrastructure: services sleep migrate secretary

.PHONY: up
up:
	@$(DOCKER-COMPOSE) up --build worker api

.PHONY: restart
restart:
	@$(DOCKER-COMPOSE) stop worker api
	@$(DOCKER-COMPOSE) up --build worker api

.PHONY: down
down:
	@$(DOCKER-COMPOSE) down --remove-orphans

CLEAN ?= hack/geocloud/* hack/minio/.minio.sys hack/minio/geocloud/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/*

.PHONY: clean
clean: down
	@rm -rf $(CLEAN)

.PHONY: prune
prune: clean
	@$(DOCKER) system prune --volumes -a

.PHONY: test
test:
	@$(GO) test -test.v -race ./...

.PHONY: vet
vet: fmt
	@$(GO) vet ./...

.PHONY: sleep
sleep:
	@sleep 2
