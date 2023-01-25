package rototiller

//go:generate make fmt static

//go:generate swag init -d ./cmd/rototiller --pd --parseDepth 4 -o ./docs/rototiller

//go:generate swag init -d ./cmd/rotoproxy --pd --parseDepth 4 -o ./docs/proxy

//go:generate buf generate .
