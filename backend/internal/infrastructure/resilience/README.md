# resilience — Circuit Breakers

Sequential circuit breaker (L87) wrapping external HTTP/DB calls.

## Usage

```go
import "byteport/internal/infrastructure/resilience"

cb := resilience.New(resilience.Config{
    FailureThreshold: 5,             // consecutive fails → open
    SuccessThreshold: 2,             // consecutive successes in half-open → close
    CoolDown:         30 * time.Second,
})

err := cb.Do(ctx, func(ctx context.Context) error {
    return callExternalAPI(ctx)
})
if errors.Is(err, resilience.ErrBreakerOpen) {
    // fast-fail; do not block on upstream
}
```

## State Machine

```
        FailureThreshold reached
Closed ───────────────────────────► Open
   ▲                                  │
   │ SuccessThreshold reached         │ CoolDown elapsed
   │                                  ▼
   └──────────────── Half-Open ──────┘
                                          │
                                          │ Any failure
                                          ▼
                                        Open
```

- **Closed**: calls flow; failures accumulate.
- **Open**: calls fast-fail with `ErrBreakerOpen`.
- **Half-Open**: probe traffic; close on `SuccessThreshold` or re-open on any failure.

## Tests

```sh
go test ./internal/infrastructure/resilience/...
```

9 unit tests cover all transitions, counter reset, concurrency, and `Do` wrapping.
