# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a minimal Go project (`github.com/herdkey/hello-go`) using Go 1.24.6. The project currently contains only the basic module setup with no source code files yet.

## Ignore Files

Do not read these at all, they will confuse you.

* `private/` — Uncommitted scratch files.

## Development Commands

Since this is a new Go project, here are the standard Go commands you'll likely need:

- `go run .` - Run the main package
- `go build` - Build the project
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up module dependencies
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Run static analysis

## Architecture

This is a fresh Go module with no existing architecture. When adding code:
- Follow standard Go project layout conventions
- Place main package in root directory or `cmd/` subdirectory for larger projects
- Use `internal/` for private packages
- Use `pkg/` for public packages if building a library
