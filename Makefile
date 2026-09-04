BINARY := bin/tgminer

.PHONY: build fmt test groups mine

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/tgminer

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

groups:
	go run ./cmd/tgminer groups

mine:
	go run ./cmd/tgminer mine
