DOCKER-COMPOSE ?= docker compose
BREW ?= brew

.PHONY: tools
tools:
	@$(GO) install github.com/swaggo/swag/cmd/swag@v1.7.8
	@$(BREW) install golangci-lint
