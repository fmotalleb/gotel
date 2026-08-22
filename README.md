# gotel

Reusable OpenTelemetry initializer for Go applications. Configures OTLP **tracing, metrics, and logs** from a single entry point.

## Install

```bash
go get github.com/fmotalleb/gotel
```

## Quick start

```go
package main

import (
    "context"

    "github.com/fmotalleb/gotel"
    "go.uber.org/zap"
)

func main() {
    ctx := context.Background()
    logger, _ := zap.NewProduction()

    result, err := gotel.Setup(ctx,
        gotel.WithServiceName("my-app"),
        gotel.WithVersion("1.0.0"),
        gotel.WithLogger(logger),
        gotel.WithTracing(&gotel.TracingConfig{
            URL:      "http://localhost:4318",
            Insecure: true,
        }),
        gotel.WithMetrics(&gotel.MetricsConfig{
            URL:      "http://localhost:4318",
            Insecure: true,
        }),
        gotel.WithLogs(&gotel.LogsConfig{
            URL:      "http://localhost:4318",
            Insecure: true,
        }),
    )
    if err != nil {
        logger.Fatal("otel setup failed", zap.Error(err))
    }
    defer result.Shutdown(ctx)
}
```

## Parse from JSON

```go
cfg, err := gotel.ParseConfig([]byte(`{
    "service_name": "my-app",
    "version": "1.0.0",
    "tracing": {"url": "http://localhost:4318", "insecure": true}
}`))
if err != nil {
    log.Fatal(err)
}

result, err := gotel.Setup(ctx, gotel.WithConfig(cfg))
```

## API

### Setup

```go
func Setup(ctx context.Context, opts ...Option) (SetupResult, error)
```

Initializes OTel providers. Each signal (tracing, metrics, logs) is enabled only if its URL is set. Returns silently if no signals are configured.

### Options

| Option | Description |
|---|---|
| `WithConfig(cfg *Config)` | Use a pre-built config struct |
| `WithLogger(logger *zap.Logger)` | Logger for OTel setup messages |
| `WithServiceName(name string)` | Override `service.name` resource attribute |
| `WithVersion(version string)` | Set `service.version` resource attribute |
| `WithAttributes(attrs map[string]string)` | Add custom resource attributes |
| `WithTracing(cfg *TracingConfig)` | Enable tracing with endpoint config |
| `WithMetrics(cfg *MetricsConfig)` | Enable metrics with endpoint config |
| `WithLogs(cfg *LogsConfig)` | Enable log export with endpoint config |
| `WithPropagator(p propagation.TextMapPropagator)` | Override default propagator |

Options are applied in order. Later options override earlier ones for scalar fields; attributes are merged.

### ParseConfig

```go
func ParseConfig(data []byte) (*Config, error)
```

Parses a JSON byte slice into a `Config` struct. Applies defaults after parsing.

### Config types

```go
type Config struct {
    ServiceName string
    Version     string
    Attributes  map[string]string
    Tracing     *TracingConfig
    Metrics     *MetricsConfig
    Logs        *LogsConfig
}

type TracingConfig struct {
    URL      string
    Insecure bool
    Headers  map[string]string
}

type MetricsConfig struct {
    URL      string
    Insecure bool
    Headers  map[string]string
    Interval time.Duration  // default: 60s
}

type LogsConfig struct {
    URL      string
    Insecure bool
    Headers  map[string]string
}
```

### Struct tags

Each type carries `env` and `default` tags for use with env-parsing libraries:

| Type | Field | `env` tag | `default` |
|---|---|---|---|
| `Config` | `ServiceName` | `GOTEL_SERVICE_NAME` | — |
| `Config` | `Version` | `GOTEL_VERSION` | — |
| `TracingConfig` | `URL` | `GOTEL_TRACING_URL` | — |
| `TracingConfig` | `Insecure` | `GOTEL_TRACING_INSECURE` | `false` |
| `MetricsConfig` | `URL` | `GOTEL_METRICS_URL` | — |
| `MetricsConfig` | `Insecure` | `GOTEL_METRICS_INSECURE` | `false` |
| `MetricsConfig` | `Interval` | `GOTEL_METRICS_INTERVAL` | `60s` |
| `LogsConfig` | `URL` | `GOTEL_LOGS_URL` | — |
| `LogsConfig` | `Insecure` | `GOTEL_LOGS_INSECURE` | `false` |

### URL schemes

- `http://` — HTTP, plaintext
- `https://` — HTTP, TLS
- `grpc://` — gRPC, plaintext
- `grpcs://` — gRPC, TLS
- bare `host:port` — HTTP, plaintext

### Convenience helpers

```go
tracer := gotel.Tracer("my-component")
ctx, span := gotel.StartSpan(ctx, "operation-name")
defer span.End()
```

### SetupResult

```go
type SetupResult struct {
    Shutdown ShutdownFunc    // Call to flush pending telemetry
    LogCore  zapcore.Core   // OTLP log bridge core (nil if logs not configured)
}
```

## Versioning

Tags follow upstream OpenTelemetry SDK releases (e.g. `v1.x.y` matching `go.opentelemetry.io/otel/sdk`).
