DOCKER-COMPOSE = docker compose
BREW = brew
PRE-COMMIT = pre-commit

.PHONY: tools
tools:
	@$(GO) install github.com/swaggo/swag/cmd/swag@v1.7.8
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	@$(BREW) install golangci-lint pre-commit
	@$(PRE-COMMIT) install
