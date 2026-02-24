# Copilot Instructions for clinia/x

Golang shared utility libraries for the Clinia ecosystem. Forked from [ory/x](https://github.com/ory/x) and extended with gRPC tracing, custom assertions, core error types, and more.

## Commands

```bash
# Run all tests (excludes Kafka)
go test `go list ./... | grep -v pubsubx/kgox`

# Run a single package's tests
go test ./errorx/...

# Run a single test
go test ./errorx/... -run TestMyTestName

# Run Kafka tests (requires Docker)
make test_kafka
make test_kafka_race   # with race detector

# Lint (excludes test files)
golangci-lint run --tests=false

# Format
gofumpt -l -w .
```

The CI test command skips `TestProviderReload` and uses env vars `ARANGO_URL` and `ELASTICSEARCH_URL` for integration tests.

## Architecture

This is a **multi-package utility monorepo** — one Go module (`github.com/clinia/x`) with many independent packages. Each top-level directory is its own package (e.g., `errorx`, `httpx`, `configx`, `pubsubx`).

**Package organization pattern:**

- Interfaces and core types are defined at the package root
- Concrete implementations live in sub-packages (e.g., `pubsubx/kgox` for Kafka, `pubsubx/inmemory` for testing)
- Generated mocks live in `mocks/` sub-directories (e.g., `pubsubx/mocks/`)

**Key packages:**

- `pubsubx` — PubSub abstraction with Kafka (`kgox`), in-memory, and message types (`messagex`) sub-packages
- `configx` — Configuration loading via [koanf](https://github.com/knadh/koanf) with file watching and JSON schema validation
- `errorx` — Typed errors with stack traces (uses `internal/tracex`)
- `snapshotx` — JSON snapshot testing wrapping [cupaloy](https://github.com/bradleyjkemp/cupaloy)
- `testx` — Generic HTTP integration test helpers (`GetJson[T]`, `PostJson[T]`, etc.)
- `otelx` / `tracex` — OpenTelemetry tracing setup and utilities

## Conventions

**Functional options** are the standard configuration pattern. Option types are named `<Modifier/Option>` (e.g., `OptionModifier func(*Provider)` in `configx`, `SubscriberOption` in `pubsubx`).

**Snapshot testing** with `snapshotx.SnapshotT(t, value, ...opts)` is used widely. Use `ExceptPaths(...)` and `ExceptNestedKeys(...)` to redact dynamic fields (IDs, timestamps). Snapshots are stored in `.snapshots/` directories alongside test files.

**Kafka tests** require a running Kafka cluster. The broker addresses are read from the `KAFKA` env var (e.g., `KAFKA=localhost:19092,localhost:29092,localhost:39092`). Use `docker-compose.ci.kgox.yml` to spin up the cluster locally.

**Formatting** uses `gofumpt` (stricter than `gofmt`). The CI checks formatting before running tests. Run `make install` to install it, then `make format` to apply.

**Linting** runs only on non-test files (`--tests=false`). Enabled linters: `gosec`, `govet`, `staticcheck`.

**PR titles** must follow Conventional Commits (enforced by `semantic.yml`).

**Releases** are automated on merge to `main` — a patch version tag is created and pushed automatically.
