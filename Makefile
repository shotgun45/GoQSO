# GoQSO Makefile

.PHONY: build test lint clean coverage help

# Default target
all: test lint build

# Build the application
build:
	@echo "Building GoQSO..."
	go build -o goqso *.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting (requires golangci-lint)
lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not installed. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	}
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f goqso
	rm -f coverage.out
	rm -f coverage.html
	rm -f test_qso_log.json

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run all checks (format, vet, lint, test)
check: fmt vet lint test

# Display help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests with coverage report"
	@echo "  lint       - Run golangci-lint"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  check      - Run all checks (fmt, vet, lint, test)"
	@echo "  help       - Display this help message"