# Use bash and fail immediately on errors
set shell := ["/usr/bin/env", "bash", "-euo", "pipefail", "-c"]

# ---- Tool binary location ----
# Put Go tools in a repo-local .bin
export GOBIN := justfile_directory() + "/.bin"
export PATH := GOBIN + ":" + env_var('PATH')

ci_mode := env_var_or_default("CI", "false")
dev_mode := if "{{ci_mode}}" == "true" { "false" } else { "true" }

# Default recipe
default: setup build-dev lint test

# CI setup task
setup: install install-tools generate
    @echo "Setup complete."

# Build the application for mac
build-dev:
    just build goos=darwin

# Build the application (for linux, by default)
build goos="linux":
    GOOS="{{goos}}" GOARCH=amd64 CGO_ENABLED=0 \
    go build -a \
    -ldflags '-s -w -extldflags=-static' \
    -o bin/hello-go-api \
    ./cmd/hello-go

# Run tests
test:
    go test $(go list ./... | grep -v 'tests/integration')

integration-test:
    go test ./tests/integration/...

# Run linter
lint args='':
    #!/usr/bin/env -S bash -eu -o pipefail
    args="{{args}}"
    fix_flag=""
    if [[ "{{dev_mode}}" == "true" ]]; then
        fix_flag="--fix"
    fi
    golangci-lint run $args $fix_flag

# Generate code from OpenAPI spec
generate:
    go generate ./api

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f internal/api/*.gen.go

install:
    go mod download

# Install development tools
install-tools:
    mkdir -p "$GOBIN"
    @if [ "{{ci_mode}}" = "true" ]; then \
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0; \
    fi
    go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

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
build-run-dev: build-dev
    ./bin/hello-go-api serve

# Test with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

docker-build: build
    ./scripts/docker_build.sh
