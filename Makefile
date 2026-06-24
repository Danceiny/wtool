# wtool Makefile

BINARY_NAME := wtool
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/Danceiny/wtool/pkg/utils.Version=$(VERSION) -X github.com/Danceiny/wtool/pkg/utils.BuildTime=$(BUILD_TIME)"

.PHONY: all build test clean install lint

all: build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/wtool

test:
	go test -v ./...

clean:
	rm -rf bin/

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/ || cp bin/$(BINARY_NAME) ~/go/bin/

lint:
	golangci-lint run ./...

# Cross-compilation
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/wtool
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/wtool
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/wtool
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/wtool
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/wtool

.DEFAULT_GOAL := build
