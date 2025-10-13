# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**hello-go** (`github.com/savisec/hello-go`) is a minimal but opinionated Go HTTP service demonstrating modern Go development patterns. It exposes a single `POST /v1/echo` endpoint plus health/readiness probes, showcasing:

- OpenAPI-first development with `oapi-codegen` for type-safe API contracts
- Clean architecture with explicit dependency wiring (no DI framework)
- Configuration via `koanf` with layered precedence (files → env vars)
- Structured logging with `slog` (JSON for prod, text for dev)
- OpenTelemetry instrumentation via `otelhttp` middleware
- Two deployment targets: HTTP server (`cmd/api`) and Lambda handler (`cmd/lambda`)

**Go Version**: 1.25.2

## Ignore Files

Do not read these, they will confuse you:
- `private/` — Uncommitted scratch files
- `cdk/cdk.out/` — CDK synthesized CloudFormation templates

## Development Commands

Rely on `just` commands. Available commands are listed in `.claude/just-commands.txt`.

### Essential Commands

```bash
just              # Setup and validate (default task)
just run          # Run the API server locally
just test         # Run unit tests
just lint         # Lint and format code
just generate     # Regenerate API types from OpenAPI spec
just build        # Build all binaries (API + Lambda)
```

### Running Tests

```bash
just test                           # Unit tests only
just test coverage="true"           # Unit tests with coverage
just integration-test api           # Integration tests for API
just integration-test lambda        # Integration tests for Lambda
```

### Docker Workflows

```bash
just docker-up-api                  # Build and run API in Docker
just docker-up-lambda               # Build and run Lambda locally
```

## Architecture

### Entry Points

The project has **two entry points** sharing common business logic:

1. **`cmd/api/main.go`**: HTTP server using Cobra CLI
   - Subcommands: `serve` (start server), `health` (built-in health check)
   - Graceful shutdown with signal handling
   - Used for standalone deployments (ECS, EC2, local dev)

2. **`cmd/lambda/main.go`**: AWS Lambda handler
   - Uses `chiadapter` to adapt Chi router to Lambda events
   - Shares router/handlers with API server
   - Used for serverless deployments (API Gateway → Lambda)

### Application Initialization

The `internal/app` package provides centralized initialization:

```go
app.Initialize(ctx) -> *Application
```

This wires together:
- Configuration loading (`internal/config`)
- Structured logging setup (`internal/logging`)
- Telemetry provider (`internal/telemetry`)
- HTTP router with handlers (`internal/router`)
- HTTP server (`internal/httpserver`)

The Lambda handler reuses this pattern but skips the HTTP server setup.

### Code Structure

```
internal/
├── app/         # Application initialization and wiring
├── config/      # Koanf-based configuration loader
├── logging/     # slog setup with dev/prod modes
├── telemetry/   # OpenTelemetry provider setup
├── httpserver/  # HTTP server wrapper with Chi router
├── router/      # Route registration (assembles handlers)
├── handlers/    # HTTP handlers (generated + custom logic)
├── services/    # Business logic layer
├── middleware/  # HTTP middleware (error handling)
└── api/         # Generated code from openapi.yml
```

### Request Flow

```
HTTP Request
  ↓
Chi Router (internal/router)
  ↓
otelhttp Middleware (observability)
  ↓
Handler (internal/handlers)
  ↓
Service (internal/services)
  ↓
Response
```

Handlers are thin adapters between HTTP and service logic. Services contain business logic and are framework-agnostic.

## OpenAPI Workflow

The API contract is defined in `api/openapi.yml` and drives code generation.

### Regenerating API Code

After editing `api/openapi.yml`, run:

```bash
just generate
```

This regenerates `internal/api/*.gen.go`:
- `models.gen.go`: Request/response types
- `server.gen.go`: Chi server interface
- `client.gen.go`: Typed HTTP client for testing

### Implementing New Endpoints

1. Add endpoint to `api/openapi.yml`
2. Run `just generate` to create types/interfaces
3. Implement handler in `internal/handlers/`
4. Register route in `internal/router/router.go`
5. Add business logic in `internal/services/`

Generated server interfaces ensure compile-time safety—missing implementations cause build failures.

## Configuration Management

Configuration uses **layered precedence** via `koanf`:

```
configs/default.yml    (defaults, checked in)
  ↓ overridden by
configs/local.yml      (local overrides, .gitignored)
  ↓ overridden by
configs/private.yml    (secrets, .gitignored)
  ↓ overridden by
Environment variables  (APP_* prefix, e.g., APP_SERVER_PORT)
```

### Configuration Structure

See `internal/config/config.go` for full schema:

```go
type Config struct {
    Server    ServerConfig    // host, port, timeouts
    Logging   LoggingConfig   // level, format (text|json)
    Telemetry TelemetryConfig // service name, version, enabled
}
```

### Setting Config via Environment

```bash
export APP_SERVER_PORT=9090
export APP_LOGGING_LEVEL=debug
export APP_LOGGING_FORMAT=json
```

Underscores in env vars map to nested config keys (converted to lowercase with dots).

## Testing Strategy

### Unit Tests

- Colocated with implementation (e.g., `handlers/echo_test.go`)
- Use `testify/require` for assertions
- Mock services using interfaces
- Run with: `just test`

### Integration Tests

Located in `tests/integration/`:
- **`api/`**: Tests against running HTTP server
- **`lambda/`**: Tests against Lambda runtime
- **`config/`**: Test configuration loader (YAML + overrides)

Integration tests use the generated `api.Client` for type-safe requests.

Configuration:
- `tests/integration/config/default.yml`: Default test config
- `tests/integration/config/local.yml`: Local overrides (.gitignored)

Run with: `just integration-test api` or `just integration-test lambda`

## Deployment

### AWS CDK Infrastructure

The `cdk/` directory contains AWS CDK v2 (TypeScript) infrastructure for deploying as a containerized Lambda behind HTTP API Gateway.

#### Prerequisites

1. ECR image already pushed to ECR registry
2. CDK bootstrapped in target account/region
3. Node.js/pnpm installed (managed by Volta)

#### Deployment Modes

**Ephemeral Test Stacks** (for PRs/branches):
```bash
cd cdk
pnpx cdk deploy \
  -c stage=test \
  -c namespace=pr-123 \
  -c ecr_image_uri=<ECR_URI>
```

**Stable Live Stacks** (prod/stage):
```bash
cd cdk
pnpx cdk deploy \
  -c stage=prod \
  -c ecr_image_uri=<ECR_URI>
```

See `cdk/README.md` for full deployment documentation.

### Docker Images

- **API**: `docker/api/` - Standalone HTTP server image
- **Lambda**: `docker/lambda/` - Lambda-compatible image with Runtime Interface Client

Build with: `just docker-build-api` or `just docker-build-lambda`

## Common Patterns

### Adding a New Service

1. Create `internal/services/myservice.go` with interface and implementation
2. Inject logger via constructor: `NewMyService(logger *slog.Logger)`
3. Add tests in `internal/services/myservice_test.go`
4. Wire into handler in `internal/router/router.go`

### Adding Middleware

1. Create middleware in `internal/middleware/`
2. Middleware signature: `func(next http.Handler) http.Handler`
3. Register in `internal/httpserver/router.go` using `router.Use()`

### Logging Best Practices

Use structured logging with context:

```go
logger.Info("processing request",
    "user_id", userID,
    "request_id", requestID)
```

Log levels:
- `Debug`: Development troubleshooting
- `Info`: Normal operations
- `Warn`: Unexpected but handled
- `Error`: Failures requiring investigation

### Error Handling

Return errors up the call stack. HTTP handlers convert errors to responses via `internal/middleware/error_handler.go`.

Service layer should return domain errors, not HTTP status codes.

## Project Dependencies

- **Router**: `chi/v5` (lightweight, composable)
- **Config**: `koanf` (layered configuration)
- **OpenAPI**: `oapi-codegen` (code generation)
- **Logging**: `slog` (standard library)
- **Observability**: `opentelemetry-go`
- **CLI**: `cobra` (API server commands)
- **Lambda**: `aws-lambda-go`, `aws-lambda-go-api-proxy`
- **Testing**: `testify`

## Linting

Uses `golangci-lint` with configuration in `.golangci.yml`. Enabled linters:
- `govet`, `staticcheck`, `revive` (bug detection)
- `gofumpt`, `goimports` (formatting)
- `ineffassign` (unused assignments)

Run with: `just lint` (auto-fixes in dev mode)

## Prerequisites

Install via Homebrew:
- `golangci-lint`
- `just`
- `direnv`

Additionally, clone [just-common](https://github.com/savisec/just-common) and set `JUST_COMMON_ROOT`:
```bash
export JUST_COMMON_ROOT=/path/to/just-common
"$JUST_COMMON_ROOT/link.py"
```

This symlinks shared justfile recipes into `.justfiles/`.
