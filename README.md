# hello-go

**hello-go** is a minimal but opinionated "Hello" service written in Go. It was bootstrapped by Claude Code, using [this prompt](./prompts/bootstrap_prompt.md).
It’s not meant to be feature-rich, but to **exemplify how to wire together common Go frameworks and tools** into a clean, idiomatic project.

The service exposes a single `POST /v1/echo` endpoint that echoes back JSON input, plus standard `/healthz` and `/readyz` endpoints. Alongside this, it demonstrates:

- OpenAPI-first development with code generation
- Layered configuration with files, envdir, environment variables, and flags
- Structured logging with a human-friendly dev mode
- OpenTelemetry instrumentation
- Explicit dependency wiring without a DI framework
- A modular project layout with many small focused files

---

## Tech Stack

- **Go 1.24.6** (with `toolchain go1.24.6`)
- **Router:** [`chi`](https://github.com/go-chi/chi)
- **Koanf:** [`koanf`](https://github.com/knadh/koanf) for configuration
- **OpenAPI:** [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) for types, Chi server, and client
- **Validation:** [`kin-openapi`](https://github.com/getkin/kin-openapi) (request/response validation in dev/CI)
- **Logging:** Go’s standard `slog`, with `"json"` (prod) and `"text"` (dev) modes
- **Observability:** [OpenTelemetry](https://opentelemetry.io/) with `otelhttp` middleware
- **Testing:** Go `testing` + [`testify`](https://github.com/stretchr/testify)
- **Task runner:** [`just`](https://github.com/casey/just) (mirrors Makefile)
- **Linting:** [`golangci-lint`](https://github.com/golangci/golangci-lint) with `govet`, `staticcheck`, `revive`, `ineffassign`, `gofumpt`, `goimports`
- **No database** yet (kept intentionally simple)

---

## Endpoints

- `POST /v1/echo` — Echo a JSON body `{"message": "...", "author": "..."}`
- `GET /healthz` — Liveness probe
- `GET /readyz` — Readiness probe
- `GET /api/openapi.yaml` — Serve the OpenAPI specification

---

## Requirements

Install the following packages with brew:
- `golangci-lint`
- `just`
- `direnv`

## Running

Clone [just-common](https://github.com/herdkey/just-common) and set `JUST_COMMON_ROOT` to the root of that repo. From the root of this project (hello-go), run this: `"$JUST_COMMON_ROOT/link.py"`

Run default task:
```shell
just
```

Then run:
```shell
just run
```
