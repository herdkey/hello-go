# Default recipe
default: generate build test lint

# Build the application
build:
    go build -o bin/hello-go ./cmd/hello-go

# Run tests
test:
    go test -v ./...

# Run linter
lint:
    golangci-lint run

# Generate code from OpenAPI spec
generate:
    go generate ./internal/api

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f internal/api/*.gen.go

# Install development tools
install-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Run the server
run:
    go run ./cmd/hello-go serve

# Format code
fmt:
    go fmt ./...

# Tidy module dependencies
tidy:
    go mod tidy

# Run server in development mode with live reload
dev:
    go run ./cmd/hello-go serve

# Build and run
build-run: build
    ./bin/hello-go serve

# Test with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
