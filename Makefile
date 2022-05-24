GO ?= go

ifneq "$(strip $(shell command -v $(GO) 2>/dev/null))" ""
	GOOS ?= $(shell $(GO) env GOOS)
else
	ifeq ($(GOOS),)
		UNAME_S := $(shell uname -s)
		ifeq ($(UNAME_S),Linux)
			GOOS = linux
		endif
		ifeq ($(UNAME_S),Darwin)
			GOOS = darwin
		endif
	else
		GOOS ?= $$GOOS
	endif
endif

-include Makefile.$(GOOS)

DOCKER ?= docker
DOCKER-COMPOSE ?= docker-compose
GCC ?= gcc
INSTALL ?= sudo install

BIN ?= /usr/local/bin

REGISTRY ?= ghcr.io
REPOSITORY ?= logsquaredn/geocloud
MODULE ?= github.com/$(REPOSITORY)
TAG ?= $(REGISTRY)/$(REPOSITORY):latest

VERSION ?= 0.0.0
PRERELEASE ?= alpha0

WHOAMI ?= $(shell whoami)

.DEFAULT_GOAL := fallthrough

.PHONY: fallthrough
fallthrough: fmt install infra detach

.PHONY: fmt
fmt:
	@$(GO) fmt ./...

geocloud geocloudctl:
	@$(GO) build -ldflags "-s -w -X $(MODULE).Version=$(VERSION) -X $(MODULE).Prerelease=$(PRERELEASE)" -o $(CURDIR)/bin $(CURDIR)/cmd/$@
	@$(INSTALL) $(CURDIR)/bin/$@ $(BIN)

install: geocloud geocloudctl

.PHONY: services
services:
	@$(DOCKER-COMPOSE) up -d datastore objectstore messagequeue

.PHONY: build
build:
	@$(DOCKER-COMPOSE) build

.PHONY: secretary
secretary:
	@$(DOCKER-COMPOSE) up --build secretary

.PHONY: migrate migrations
migrate migrations:
	@$(DOCKER-COMPOSE) up --build migrate
	@$(DOCKER-COMPOSE) exec -T datastore psql -U geocloud -c "INSERT INTO customer VALUES ('$(WHOAMI)') ON CONFLICT DO NOTHING;"

.PHONY: infra infrastructure
infra infrastructure: services sleep migrate secretary

.PHONY: up
up:
	@$(DOCKER-COMPOSE) up --build worker api

.PHONY: detach
detach:
	@$(DOCKER-COMPOSE) up -d --build worker api
	
.PHONY: restart
restart:
	@$(DOCKER-COMPOSE) stop worker api
	@$(DOCKER-COMPOSE) up --build worker api

.PHONY: down
down:
	@$(DOCKER-COMPOSE) down --remove-orphans

CLEAN ?= bin/* hack/geocloud/* hack/minio/.minio.sys hack/minio/geocloud/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/*

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
