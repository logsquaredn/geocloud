package rototiller

//go:generate swag fmt -d ./ --generalInfo ./cmd/rototiller/main.go

//go:generate swag init -d ./cmd/rototiller --pd --parseDepth 4 -o ./docs/rototiller

//go:generate swag init -d ./cmd/rotoproxy --pd --parseDepth 4 -o ./docs/proxy

//go:generate go fmt ./...

//go:generate buf format -w

//go:generate buf generate .
