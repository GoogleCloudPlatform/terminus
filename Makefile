.PHONY: build test lint clean run-example

# Build the example application
build:
	go build -o bin/example ./cmd/example

# Run tests with coverage
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Run the example application
run-example: build
	./bin/example

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests and linter
check: test lint

# Development setup
setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download