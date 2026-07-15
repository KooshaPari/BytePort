# BytePort Structured Logging

BytePort uses Go's standard [`log/slog`](https://pkg.go.dev/log/slog) package (Go 1.21+) for
structured, context-aware logging. This document covers the conventions and helpers in
`backend/internal/infrastructure/logging`.

## Goals

1. **Structured** — every record is a JSON object (or text in dev) with stable keys.
2. **Contextual** — request_id, trace_id, span_id, user_id, org_id propagate through `context.Context`.
3. **Conventional** — every record carries `service`, `version`, `env` base attributes.
4. **Fast** — JSON handler is allocation-aware; default handler writes to stderr.

## Quick start

```go
import "github.com/byteport/api/internal/infrastructure/logging"

logger := logging.New(logging.Config{
    Service: "byteport-api",
    Version: version.Version,
    Env:     "production",
    Level:   logging.LevelInfo,
    Format:  logging.FormatJSON, // default for production
})

logger.Info("deployment started",
    "deployment_id", depID,
    "user_id", userID,
    "duration_ms", elapsed.Milliseconds(),
)
```

## Output formats

| Environment | Format | Source info | Color |
|-------------|--------|-------------|-------|
| `production` | JSON  | disabled | none |
| `staging`    | JSON  | disabled | none |
| `development`| text  | enabled | none |
| `test`       | JSON  | disabled | none |

The format is auto-selected based on `Env` but can be overridden via `Config.Format`.

## Level filtering

```go
logger := logging.New(logging.Config{
    Level: logging.ParseLevel(os.Getenv("LOG_LEVEL")), // debug|info|warn|error
})
```

Levels mirror `slog.Level`:
- `LevelDebug` (-4) — verbose, only useful for diagnosis
- `LevelInfo` (0) — default
- `LevelWarn` (4) — recoverable issues
- `LevelError` (8) — failures

## Context propagation

```go
// At request entry:
ctx := logging.WithLogger(r.Context(), logger.With("request_id", reqID))
r = r.WithContext(ctx)

// Anywhere downstream:
logger := logging.FromContext(ctx)
logger.Info("doing work")
```

The package stores `*slog.Logger` in the context. `FromContext` falls back to
`slog.Default()` if none is set.

## Helper attributes

| Helper | Adds |
|--------|------|
| `WithRequestID(logger, id)` | `request_id` |
| `WithTrace(logger, traceID, spanID)` | `trace_id`, `span_id` |
| `WithUser(logger, userID, orgID)` | `user_id`, `org_id` |

When paired with OpenTelemetry, propagate `trace_id`/`span_id` from the active span.

## Environment variables

| Var | Purpose | Default |
|-----|---------|---------|
| `BYTEPORT_ENV` | Service environment | `development` |
| `ENVIRONMENT`  | Alternative env var | — |
| `NODE_ENV`     | Common convention | — |
| `GO_ENV`       | Go convention | — |
| `LOG_LEVEL`    | debug/info/warn/error | `info` |
| `LOG_FORMAT`   | json/text | json in prod, text otherwise |

## Examples

### Production JSON output

```json
{"time":"2026-07-10T12:00:00Z","level":"INFO","msg":"deployment started","service":"byteport-api","version":"1.2.0","env":"production","deployment_id":"d-123","user_id":"u-abc"}
```

### Development text output

```
time=2026-07-10T12:00:00.000-07:00 level=INFO msg="deployment started" service=byteport-api version=1.2.0 env=development deployment_id=d-123 user_id=u-abc
```

## Best practices

1. **Never log secrets.** Redact tokens, API keys, and passwords before adding to attributes.
2. **Use stable keys.** Use `deployment_id`, not `depId` or `id`. Enums (e.g., `level`) should use the same casing.
3. **Bound cardinality.** Avoid using unbounded user input as a label value.
4. **Prefer context propagation** over passing the logger explicitly.
5. **Log at boundaries.** Entry/exit of HTTP handlers, RPC methods, and queue workers.
6. **Sample noisy events.** At scale, drop verbose INFO records from `debug=trace` modes.

## Related packages

- [`backend/internal/infrastructure/otel`](../otel) — OpenTelemetry tracing/metrics that integrate with slog.
- [`backend/internal/infrastructure/observability`](../../../../../internal/infrastructure/observability) — Prometheus metrics.