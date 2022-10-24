GO = go

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

DOCKER-COMPOSE = docker-compose

-include mk/$(GOOS).mk

PKGS = $(shell $(GO) list ./... | grep -v /cmd/| grep -v /docs)

DOCKER = docker
SWAG = swag
LINT = golangci-lint
BUF = buf
GCC ?= gcc
INSTALL ?= sudo install
GIT ?= git

BIN ?= /usr/local/bin

REGISTRY ?= ghcr.io
REPOSITORY ?= logsquaredn/rototiller
MODULE ?= github.com/$(REPOSITORY)
TAG ?= $(REGISTRY)/$(REPOSITORY):latest

VERSION ?= 0.0.0
PRERELEASE ?= alpha0

WHOAMI ?= $(shell whoami)

.DEFAULT_GOAL := fallthrough

.PHONY: fallthrough
fallthrough: generate fmt install infra detach

.PHONY: fmt
fmt:
	@$(GO) $@ ./...
	@$(SWAG) fmt -d ./ --generalInfo ./cmd/rototiller/main.go
	@$(BUF) format -w

.PHONY: vet generate
vet generate:
	@$(GO) $@ ./...

.PHONY: download tidy
	@$(GO) mody $@

.PHONY: lint
lint:
	@$(LINT) run

.PHONY: tests
tests:
	@for pkg in $(PKGS); do \
		$(GO) test -o $(CURDIR)/bin/$$(basename $$pkg).test -c $$pkg; \
	done

.PHONY: test
test: tests
	@for test in $(CURDIR)/bin/*.test; do \
		$$test; \
	done

.PHONY: rototiller rotoctl
rototiller rotoctl:
	@$(GO) build -ldflags "-s -w -X $(MODULE).Version=$(VERSION) -X $(MODULE).Prerelease=$(PRERELEASE)" -o $(CURDIR)/bin $(CURDIR)/cmd/$@

.PHONY: install-rototiller
install-rototiller: rototiller
	@$(INSTALL) $(CURDIR)/bin/rototiller $(BIN)

.PHONY: install-rotoctl
install-rotoctl: rotoctl
	@$(INSTALL) $(CURDIR)/bin/rotoctl $(BIN)

install: install-rototiller install-rotoctl

.PHONY: services
services:
	@$(DOCKER-COMPOSE) up -d minio postgres rabbitmq

.PHONY: build
build:
	@$(DOCKER-COMPOSE) build

.PHONY: secretary
secretary:
	@$(DOCKER-COMPOSE) up --build secretary

.PHONY: migrate
migrate:
	@$(DOCKER-COMPOSE) up --build migrate

.PHONY: infra infrastructure
infra infrastructure: services sleep migrate secretary

.PHONY: up
up:
	@$(DOCKER-COMPOSE) up --build worker api proxy

.PHONY: detach
detach:
	@$(DOCKER-COMPOSE) up -d --build worker api proxy
	
.PHONY: restart
restart:
	@$(DOCKER-COMPOSE) stop worker api proxy
	@$(DOCKER-COMPOSE) up --build worker api proxy

.PHONY: down
down:
	@$(DOCKER-COMPOSE) down --remove-orphans

CLEAN ?= bin/* hack/rototiller/* hack/rototiller/blobstore/* hack/minio/.minio.sys hack/minio/rototiller-archive/* hack/minio/rototiller/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/*

.PHONY: clean
clean: down
	@rm -rf $(CLEAN)

.PHONY: prune
prune: clean
	@$(DOCKER) system prune --volumes -a

.PHONY: sleep
sleep:
	@sleep 2

.PHONY: docs
docs:
	@$(SWAG) init -d ./cmd/rototiller --pd --parseDepth 4

MIGRATION = $(shell date -u +%Y%m%d%T | tr -cd [0-9])
TITLE ?= replace_me

.PHONY: migration
migration:
	@touch pkg/store/data/postgres/sql/migrations/$(MIGRATION)_$(TITLE).up.sql
	@echo "created pkg/store/data/postgres/sql/migrations/$(MIGRATION)_$(TITLE).up.sql; replace title and add SQL"

RELEASE ?= $(VERSION)
ifneq "$(strip $(PRERELEASE))" ""
	RELEASE = $(VERSION)-$(PRERELEASE)
endif

.PHONY: release
release:
	@$(GIT) tag -a $(RELEASE) -m $(RELEASE)
	@$(GIT) push --follow-tags
