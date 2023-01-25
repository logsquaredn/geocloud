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
GOLANGCI-LINT = golangci-lint
BUF = buf
GCC ?= gcc
INSTALL ?= sudo install
GIT ?= git
NPM ?= npm

BIN ?= /usr/local/bin

REGISTRY ?= ghcr.io
REPOSITORY ?= logsquaredn/rototiller
MODULE ?= github.com/$(REPOSITORY)
TAG ?= $(REGISTRY)/$(REPOSITORY):latest

SEMVER ?= 0.0.0

.DEFAULT_GOAL := fallthrough

fallthrough: generate fmt install infra detach

fmt:
	@$(GO) $@ ./...
	@$(SWAG) fmt -d ./ --generalInfo ./cmd/rototiller/main.go
	@$(SWAG) fmt -d ./ --generalInfo ./cmd/rotoproxy/main.go
	@$(BUF) format -w

test vet generate:
	@$(GO) $@ ./...

download tidy:
	@$(GO) mody $@

lint:
	@$(GOLANGCI-LINT) run --fix

static:
	@cd ui/ && $(NPM) run build
	@cp -R ui/build/* static/

rototiller rotoctl:
	@$(GO) build -ldflags "-s -w -X $(MODULE).Semver=$(SEMVER)" -o $(CURDIR)/bin $(CURDIR)/cmd/$@

install-rototiller: rototiller
	@$(INSTALL) $(CURDIR)/bin/rototiller $(BIN)

install-rotoctl: rotoctl
	@$(INSTALL) $(CURDIR)/bin/rotoctl $(BIN)

install: install-rototiller install-rotoctl

services:
	@$(DOCKER-COMPOSE) up -d minio postgres rabbitmq

secretary migrate:
	@$(DOCKER-COMPOSE) up --build $@

infra infrastructure: services sleep migrate secretary

up:
	@$(DOCKER-COMPOSE) up --build worker api proxy

detach:
	@$(DOCKER-COMPOSE) up -d --build worker api proxy

restart:
	@$(DOCKER-COMPOSE) stop worker api proxy
	@$(DOCKER-COMPOSE) up --build worker api proxy

down:
	@$(DOCKER-COMPOSE) $@ --remove-orphans

clean: down
	@rm -rf bin/* hack/rototiller/* hack/rototiller/blobstore/* hack/minio/.minio.sys hack/minio/rototiller-archive/* hack/minio/rototiller/* hack/postgresql/* hack/rabbitmq/lib/* hack/rabbitmq/lib/.erlang.cookie hack/rabbitmq/log/*

prune: clean
	@$(DOCKER) system $@ --volumes -a

sleep:
	@$@ 2

MIGRATION = $(shell date -u +%Y%m%d%T | tr -cd [0-9])
TITLE ?= replace_me

migration:
	@touch pkg/store/data/postgres/sql/migrations/$(MIGRATION)_$(TITLE).up.sql
	@echo "created pkg/store/data/postgres/sql/migrations/$(MIGRATION)_$(TITLE).up.sql; replace title and add SQL"

release:
	@$(GIT) tag -a v$(SEMVER) -m v$(SEMVER)
	@$(GIT) push --follow-tags

gen: generate

.PHONY: clean retach down download fallthrough fmt gen generate infra infrastructure \
	install-rotoctl install-rototiller linnt migrate migration prune release restart \
	rotoctl rototiller secretary services sleep static tidy up vet
