# AGENTS.md — gotel

## What this is

`gotel` is a reusable OpenTelemetry initializer library. It provides a single
entry point to configure OTLP **tracing, metrics, and logs** in Go applications.

Goal: unify OTel setup across your apps via an **Options Pattern** and JSON
config parsing.

## Rules for agents

- **Commit after every meaningful change.** One logical change = one commit.
- **Write tests for every addition.** No new exported function or type ships without a test.
- **Tagging follows upstream OpenTelemetry versioning** (e.g. `v0.x.y` matching OTel SDK releases).
- Do not add comments unless asked.
- Run `golangci-lint run` before committing any Go change (requires golangci-lint v2).
- Run `go test ./...` before committing any Go change.
- Prefer table-driven tests (existing pattern in `observability_test.go`).
- Use `github.com/alecthomas/assert/v2` for test assertions (existing choice).

## Developer commands

```
golangci-lint run          # lint (uses .golangci.yml — v2 schema, requires golangci-lint v2)
go test ./...              # all tests
go test -run TestName      # single test
go mod tidy                # keep deps clean (required before commit)
```

There is no Makefile or task runner. Commands are run directly.

## Architecture

Single flat Go package — `package gotel`. No sub-packages.

Key files:
- `config.go` — `Config`, `TracingConfig`, `MetricsConfig`, `LogsConfig` types, `ParseConfig()` for JSON
- `options.go` — Functional options: `WithConfig`, `WithServiceName`, `WithVersion`, `WithTracing`, etc.
- `observability.go` — `Setup()` entry point, provider wiring
- `exporters.go` — HTTP/gRPC option builders, exporter constructors
- `resource.go` — OTel resource construction
- `observability_test.go` — table-driven tests for endpoint parsing
- `config_test.go` — tests for JSON config parsing

The `Setup()` function accepts variadic `Option` arguments and returns a
`SetupResult` containing a `ShutdownFunc` and optional `zapcore.Core` for the
OTLP log bridge.

## Public API

```go
// Primary entry point.
func Setup(ctx context.Context, opts ...Option) (SetupResult, error)

// Options.
WithConfig(cfg *Config) Option
WithLogger(logger *zap.Logger) Option
WithServiceName(name string) Option
WithVersion(version string) Option
WithAttributes(attrs map[string]string) Option
WithTracing(cfg *TracingConfig) Option
WithMetrics(cfg *MetricsConfig) Option
WithLogs(cfg *LogsConfig) Option
WithPropagator(p propagation.TextMapPropagator) Option

// JSON parsing.
func ParseConfig(data []byte) (*Config, error)

// Convenience helpers.
func Tracer(name string) trace.Tracer
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
```

## Config structure

```json
{
  "service_name": "my-app",
  "version": "1.0.0",
  "attributes": {"deployment.environment": "production"},
  "tracing": {"url": "http://localhost:4318", "insecure": true},
  "metrics": {"url": "http://localhost:4318", "interval": 60000000000},
  "logs": {"url": "http://localhost:4318"}
}
```

Parsed from JSON via `ParseConfig(data []byte)`.

## Struct tags

Fields carry `env` and `default` tags for use with env-parsing libraries:

- `Config.ServiceName` — `env:"GOTEL_SERVICE_NAME"`
- `Config.Version` — `env:"GOTEL_VERSION"`
- `TracingConfig.URL` — `env:"GOTEL_TRACING_URL"`
- `TracingConfig.Insecure` — `env:"GOTEL_TRACING_INSECURE"`, `default:"false"`
- `MetricsConfig.URL` — `env:"GOTEL_METRICS_URL"`
- `MetricsConfig.Insecure` — `env:"GOTEL_METRICS_INSECURE"`, `default:"false"`
- `MetricsConfig.Interval` — `env:"GOTEL_METRICS_INTERVAL"`, `default:"60s"`
- `LogsConfig.URL` — `env:"GOTEL_LOGS_URL"`
- `LogsConfig.Insecure` — `env:"GOTEL_LOGS_INSECURE"`, `default:"false"`

`ParseConfig` applies defaults after JSON unmarshalling.

## Linting quirks

`.golangci.yml` is tightly tuned for this repo:

- **godot** requires declaration comments end in a period, lowercase start.
- **gocyclo** threshold is 20; **funlen** is 120 lines / 80 statements.
- Exclusions: `_test.go` excludes gosec, goconst, funlen, gocyclo.
- Formatters: gofumpt and goimports run via golangci-lint with
  `module-path: gotel` and `local-prefixes: gotel`.

## External OTel dependencies

Pinned to OTel SDK semconv `v1.26.0` (see import in `resource.go`).
When upgrading OTel, update all `semconv` imports together.
