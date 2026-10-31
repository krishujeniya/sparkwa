# SparkWA Lite — Build Automation
SHELL := /bin/bash
VERSION ?= v1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "release")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.GitCommit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)

.PHONY: all build build-all test test-race vet doctor package clean

all: test build

build:
	@echo "Building SparkWA Lite ($(VERSION))..."
	@mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o bin/sparkwa ./cmd/sparkwa
	@echo "✓ Binary compiled: bin/sparkwa"

build-all:
	@echo "Cross-compiling linux/amd64 and linux/arm64..."
	@mkdir -p bin/linux-amd64 bin/linux-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o bin/linux-amd64/sparkwa ./cmd/sparkwa
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o bin/linux-arm64/sparkwa ./cmd/sparkwa
	@echo "✓ Multi-arch compilation complete."

test:
	@echo "Running unit and integration tests..."
	go test -v ./...

test-race:
	@echo "Running tests with race detection..."
	go test -race ./internal/config ./internal/identity ./internal/message ./internal/spark ./internal/store ./internal/worker ./internal/health ./tests

vet:
	@echo "Checking go vet..."
	go vet ./...

doctor: build
	./bin/sparkwa doctor

package:
	@chmod +x scripts/package.sh
	./scripts/package.sh

clean:
	rm -rf bin dist data *.db *.db-wal *.db-shm
