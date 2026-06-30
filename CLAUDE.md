# BytePort

## Build & Test

```bash
# Rust workspace (parser, transport, CLI, OTel crates)
cargo check --workspace
cargo test --workspace
cargo clippy --workspace -- -D warnings

# Go backend (gin API, NVMS provisioner, cloud layer)
cd backend && go build ./... && go test ./...

# Frontend (Astro + Svelte)
cd frontend/web && npm ci && npm run build

# Full stack via justfile
just build
just test
just lint
```

## Project Layout

```
BytePort/
├── backend/          # Go API server (gin), NVMS provisioner, cloud lib
│   ├── main.go       # Entry point; loads .env, inits DB, starts gin
│   ├── server.go     # Route registration
│   ├── handlers.go   # Top-level request handlers
│   ├── lib/          # Encryption, auth, AWS credentials
│   ├── models/       # DB models; ConnectDatabase() (SQLite via GORM)
│   ├── nvms/         # NVMS manifest parsing + deploy orchestration
│   │   ├── projectManager/deploy.go  # Core deploy pipeline
│   │   └── lib/      # LLM providers (OpenAI, Anthropic, Gemini, Local)
│   ├── bytebridge/   # AWS resource bridge
│   └── internal/     # Container utilities
├── backend/nvms.rs   # Rust YAML parser (standalone, not part of Go build)
├── crates/           # Rust workspace members
│   ├── byteport-cli/         # CLI binary
│   ├── byteport-dag/         # DAG execution engine
│   ├── byteport-otel/        # OTel instrumentation
│   └── byteport-transport/   # Upload transport abstraction
├── frontend/
│   └── web/          # Astro + Svelte UI (npm)
├── src/              # Shared TypeScript utilities
├── ports/            # Port/platform integration layer
├── docs/             # Architecture, operations, security, journey docs
├── scripts/          # Coverage scripts, build helpers
├── tests/            # Integration and E2E tests
├── .github/workflows/# CI: ci.yml (main gate), audit.yml, coverage gates
└── justfile          # Task runner (build, test, lint, release targets)
```

## Key Services

- **API server** (`backend/`): Go/gin, port 8080 default, SQLite via GORM
- **NVMS provisioner** (`backend/nvms/`): Parses `nvms.yaml`, orchestrates AWS deploy
- **Rust parser** (`backend/nvms.rs`, `crates/`): YAML config validation, CLI tooling
- **Frontend** (`frontend/web/`): Astro SSR + Svelte components

## Conventions

- Follow existing code style; do not bypass linters/formatters/type checkers.
- Add or update tests for any new behavior (see CONTRIBUTING.md).
- Go: `gofmt`, `golangci-lint` (config: `golangci.yml`).
- Rust: `rustfmt`, `clippy` with `clippy.toml` deny list.
- TypeScript/Svelte: ESLint + Prettier (config: `jest.config.js`).
- Secrets: never commit; use `.env` (gitignored); see SECURITY.md.
- Reference: `~/.claude/CLAUDE.md` and `../../CLAUDE.md` (global Phenotype governance).

## CI Gates

- `ci.yml` — main PR gate (build, lint, test across Go/Rust/TS)
- `tier2-coverage-gate.yml` — coverage thresholds (Rust >=71%, Go >=70%, frontend >=60%)
- `fr-coverage.yml` — FR coverage (Rust + Go, >=50% lines)
- `audit.yml` — weekly security scans (CodeQL, Gitleaks, TruffleHog, cargo-audit)
- `lint.yml` — dedicated lint job
- Full inventory: `docs/architecture/` and `.github/workflows/`
