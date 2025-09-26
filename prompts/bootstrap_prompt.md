## Task
Scaffold a new Go project called `hello-go`.

## Requirements

- **Go version:** 1.24.6 (`toolchain go1.24.6` in `go.mod`)

### HTTP & API
- Use `net/http` + `chi`.
- Add `/healthz` and `/readyz` endpoints.
- API defined in `api/openapi.yml`, embed with `go:embed`.
- Use **oapi-codegen** to generate types, Chi server, and client.
- Serve the spec at `/api/openapi.yml`.
- Implement one endpoint: `POST /v1/echo` → echoes JSON body, which contains two fields `message` and `author`.

### Config
- Use **cobra** + **viper**.
- Layered sources: `config.default.yml` → envdir → env vars (prefix `APP_`) → flags.

### Logging
- Use **slog**.
- Configurable format: `"json"` for prod, `"text"` for dev.
- Level configurable (`debug`, `info`, `warn`, `error`).

### Observability
- Use **OpenTelemetry** (`otel`, `otelhttp`).
- Initialize tracer/meter provider, shutdown cleanly.

### Testing
- Use stdlib `testing` + **testify**.
- (Skip integration tests for now.)

### IoC / DI
- **No DI framework.**
- Explicit wiring via constructors + interfaces.
- Wiring/composition root lives in `internal/app/init.go`. Keep `main.go` very small. But init.go should also be small, and individual constructions should live in other files.

### Structure
- Prefer many small focused files.
- Suggested packages:
    - `internal/config` (structs, load)
    - `internal/logging` (slog setup)
    - `internal/telemetry` (otel setup)
    - `internal/httpserver` (router, health, server)
    - `internal/handlers` (echo handler)
    - `internal/services` (echo service)
    - `internal/embedfs` (embed openapi.yml)
    - `internal/api` (oapi-codegen output + generate.go)
    - `internal/app` (composition root init.go)

### Tooling
- `Makefile` + `justfile` with targets: build, test, lint, generate.
- `.golangci.yml` enabling: govet, staticcheck, revive, ineffassign, gofumpt, goimports.
- `tools.go` to pin `oapi-codegen` + `golangci-lint`.

### Docker

- A dockerfile under `docker/`.
- Dockerfile should use multi-stage build to minimize size.
- A `docker/compose.yml` to run it.

## Deliverables
- Complete directory scaffold.
- `go.mod` with required deps (Claude can resolve versions).
- A working server:
  ```sh
  make generate
  go run ./cmd/hello-go-api serve
- Working unit tests, `just` targets, observability, logs, configuration, etc.
- Working docker image & compose file.

## Style
- Trailing newlines at the end of files
