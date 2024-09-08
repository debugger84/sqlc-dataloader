.PHONY: build test

build:
	go build ./...

test: bin/sqlc-dataloader.wasm
	go test ./...

all: bin/sqlc-dataloader bin/sqlc-dataloader.wasm

bin/sqlc-dataloader: bin go.mod go.sum $(wildcard **/*.go)
	cd plugin && go build -o ../bin/sqlc-dataloader ./main.go

bin/sqlc-dataloader.wasm: bin/sqlc-dataloader
	cd plugin && GOOS=wasip1 GOARCH=wasm go build -o ../bin/sqlc-dataloader.wasm main.go

bin:
	mkdir -p bin
