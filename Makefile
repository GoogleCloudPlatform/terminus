.PHONY: help build build-cli test lint clean run-example run-cli deps check setup

# Default target
all: help

# Display help information
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the example application"
	@echo "  build-cli    Build the CLI client"
	@echo "  test         Run tests with coverage"
	@echo "  lint         Run linter"
	@echo "  clean        Clean build artifacts"
	@echo "  run-example  Run the example application (Web Server)"
	@echo "  run-cli      Run the CLI client (requires running server)"
	@echo "  deps         Install dependencies"
	@echo "  check        Run tests and linter"
	@echo "  setup        Setup development environment"

# Build the example application
build:
	@echo "Building example..."
	go build -o bin/example ./cmd/example

# Build the CLI client
build-cli:
	@echo "Building CLI client..."
	go build -o bin/terminus-cli ./cmd/terminus

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ coverage.out coverage.html terminus-cli

# Run the example application
run-example: build
	@echo "Running example..."
	./bin/example

# Run the CLI client
run-cli: build-cli
	@echo "Running CLI client..."
	./bin/terminus-cli

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run tests and linter
check: test lint

# Development setup
setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download
