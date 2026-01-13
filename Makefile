.PHONY: build clean test lint install

# Build variables
BINARY_NAME=octl
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build the binary
build:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/octl

# Build without cgo (no keychain support on macOS)
build-nocgo:
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/octl

# Install to GOPATH/bin
install:
	CGO_ENABLED=1 $(GOCMD) install $(LDFLAGS) ./cmd/octl

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	$(GOLINT) run

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Check for vulnerabilities
vuln:
	$(GOCMD) run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Build for multiple platforms
build-all: build-linux build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/octl
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/octl

build-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/octl
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/octl

# Development helpers
dev: fmt lint test build

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary (with cgo)"
	@echo "  build-nocgo   - Build without cgo (no keychain)"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  tidy          - Tidy go.mod"
	@echo "  fmt           - Format code"
	@echo "  build-all     - Build for all platforms"
	@echo "  dev           - Format, lint, test, and build"
