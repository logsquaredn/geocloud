package rototiller

//go:generate swag fmt -d ./cmd/rototiller

//go:generate swag init -d ./cmd/rototiller --pd --parseDepth 4

//go:generate go fmt ./...

//go:generate buf format -w

//go:generate buf generate .
