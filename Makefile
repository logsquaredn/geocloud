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
PRE-COMMIT ?= pre-commit

BIN ?= /usr/local/bin

REGISTRY ?= ghcr.io
REPOSITORY ?= logsquaredn/rototiller
MODULE ?= github.com/$(REPOSITORY)
TAG ?= $(REGISTRY)/$(REPOSITORY):latest

SEMVER ?= 0.0.0

.DEFAULT_GOAL := fallthrough

fallthrough: install infra detach

fmt generate test vet:
	@$(GO) $@ ./...

download vendor verify:
	@$(GO) mod $@

protos:
	@$(BUF) format -w
	@$(BUF) generate .

lint:
	@$(GOLANGCI-LINT) run --fix

docs:
	@$(SWAG) init -d ./cmd/rototiller --pd --parseDepth 4 -o ./pkg/docs/rototiller
	@$(SWAG) init -d ./cmd/rotoproxy --pd --parseDepth 4 -o ./pkg/docs/proxy

rototiller rotoctl rotoproxy:
	@$(GO) build -ldflags "-s -w -X $(MODULE).Semver=$(SEMVER)" -o $(CURDIR)/bin $(CURDIR)/cmd/$@

install-rototiller: rototiller
	@$(INSTALL) $(CURDIR)/bin/rototiller $(BIN)

install-rotoctl: rotoctl
	@$(INSTALL) $(CURDIR)/bin/rotoctl $(BIN)

install-rotoproxy: rotoproxy
	@$(INSTALL) $(CURDIR)/bin/rotoproxy $(BIN)

install: install-rototiller install-rotoctl install-rotoproxy

services:
	@$(DOCKER-COMPOSE) up -d minio postgres rabbitmq

secretary:
	@$(DOCKER-COMPOSE) up --build secretary

migrate:
	@$(DOCKER-COMPOSE) up --build migrate

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

proto: protos
buf: proto
gen: generate
dl: download
ven: vendor
ver: verify
format: fmt
	@$(SWAG) fmt -d ./ -g ./cmd/rototiller/main.go
	@$(SWAG) fmt -d ./ -g ./cmd/rotoproxy/main.go
	@$(BUF) format -w
