# Terraform Provider for BIND9 - Makefile

BINARY_NAME=terraform-provider-bind9
VERSION=1.0.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR=~/.terraform.d/plugins/example/bind9/$(VERSION)/$(OS_ARCH)

.PHONY: all build install clean test docs

all: build

# Build the provider
build:
	go build -o $(BINARY_NAME)

# Install locally for testing
install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY_NAME) $(PLUGIN_DIR)/

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Run tests
test:
	go test -v ./...

# Run acceptance tests (requires running BIND9 API)
testacc:
	TF_ACC=1 go test -v ./... -timeout 30m

# Generate documentation
docs:
	go generate ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)_linux_amd64
	GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)_linux_arm64
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)_darwin_arm64
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)_windows_amd64.exe

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run the example
example:
	cd examples && terraform init && terraform plan

