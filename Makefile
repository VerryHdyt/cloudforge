.PHONY: help run build test test-race vet vuln fmt tidy check

BACKEND_DIR := backend

help:
	@echo "CloudForge Development Commands"
	@echo "  make run        - Run API server"
	@echo "  make build      - Build all Go packages"
	@echo "  make test       - Run unit tests"
	@echo "  make test-race  - Run tests with race detector"
	@echo "  make vet        - Run Go vet"
	@echo "  make vuln       - Run govulncheck"
	@echo "  make fmt        - Format Go source files"
	@echo "  make tidy       - Clean Go dependencies"
	@echo "  make check      - Run all quality checks"

run:
	cd $(BACKEND_DIR) && go run ./cmd/api

build:
	cd $(BACKEND_DIR) && go build ./...

test:
	cd $(BACKEND_DIR) && go test ./...

test-race:
	cd $(BACKEND_DIR) && go test -race ./...

vet:
	cd $(BACKEND_DIR) && go vet ./...

vuln:
	cd $(BACKEND_DIR) && govulncheck ./...

fmt:
	cd $(BACKEND_DIR) && gofmt -w $$(find . -name '*.go' -type f)

tidy:
	cd $(BACKEND_DIR) && go mod tidy

check: test test-race vet vuln
