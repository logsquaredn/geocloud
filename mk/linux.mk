DOCKER-COMPOSE = docker-compose

.PHONY: tools
tools:
	@$(GO) install github.com/swaggo/swag/cmd/swag@v1.7.8
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
