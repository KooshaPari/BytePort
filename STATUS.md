# BytePort — Status

> Last updated: 2026-07-08 — reflects current hexagonal architecture and resolved security gaps.

## What BytePort is

BytePort is a **self-hosted Infrastructure-as-Code deployment + portfolio UX generation
platform**. Developers define one manifest (`odin.nvms`) at their repo root; BytePort
provisions a MicroVM-backed deployment, registers the resulting endpoints with a portfolio
site, and uses an LLM to generate showcase metadata for each project.

## Architecture (current)

```
┌─ Gin HTTP ─────────────────────────────────────┐
│  middleware: CORS → OTel → Auth → CSRF → Rate  │
│  → RBAC → handler                              │
│                                                 │
│  hexagonal: POST /api/v1/deployments/...        │
│  legacy:    POST /api/v1/legacy/deployments/... │
│  webhook:   POST /api/v1/webhooks/github        │
│  UDS proxy: POST /api/v1/chat/completions       │
│           → UDS socket → (Rust engine daemon)   │
└─────────────────────────────────────────────────┘
                        │
            DI Container │ (5 use cases wired)
                ┌───────┴───────┐
           Application      Infrastructure
           (use cases)      (repositories, middleware, secrets)
                │
           Domain Layer
           (entities, state machines)
                │
       ┌────────┴────────┐
  ┌────┴────┐   ┌───┴────┐
  │ NVMS   │   │ Rust   │
  │ (Spin) │   │ Engine │
  │ port   │   │ daemon │
  │ 9700   │   │ :9703  │
  └─────────┘   └────────┘
```

| Component | Language | Status | Purpose |
|---|---|---|---|
| `backend/` | Go 1.25 | ✅ operational | Gin + GORM + SQLite + hexagonal (domain/application/infrastructure) |
| `backend/nvms/` | Go 1.25 (Spin WASM) | ✅ operational | Spin module — deploy/terminate HTTP API, auth enabled |
| `crates/byteport-engine/` | Rust | ✅ building | Engine trait + adapters (mock, docker, nvms) + daemon binary |
| `crates/byteport-cli/` | Rust | ☐ stub | DAG execution CLI |
| `frontend/web/` | SvelteKit 2 / Svelte 5 / Tailwind 4 | ✅ building | Admin UI (~17 routes) |
| `frontend/web/src-tauri/` | Rust + Tauri 2 | ☐ not started | Desktop shell |

## Current health

| Gate | Status | Notes |
|---|---|---|
| `go build ./...` (backend) | ✅ green | 0 errors, 11 packages |
| `go test ./...` (backend) | ✅ green | 49 test files, all pass |
| `gofmt -l .` (backend) | ✅ clean | 0 formatting issues |
| `cargo check` (byteport-engine) | ✅ green | 28 tests pass |
| `go vet ./...` | ✅ green | 0 issues |
| `npm run check` (frontend) | ⬜ TBD | SvelteKit 2 + svelte-check |

## Resolved gaps (since 2026-06-12)

All 7 items from the old STATUS.md are **resolved**:

| Old Gap | Resolution | Date |
|---------|-----------|------|
| `routes/instances.go:12` Find().Where() bug | File deleted — hexagonal rewrite | sibling session |
| `routes/projects.go:12` Find().Where() bug | File deleted — hexagonal rewrite | sibling session |
| `routes/deployment.go` missing returns | File deleted — hexagonal rewrite | sibling session |
| nvms auth middleware commented out | Re-enabled, PASETO v4 | sibling session |
| deploy.go:51 NVMS YAML TODO | `parseNVMSConfig()` at `nvms/projectManager/parser.go:18` | sibling session |
| `llm.go:22` ProviderGemini stub | Full Gemini provider at `nvms/lib/providers/gemini/gemini.go:90` | sibling session |
| `nvms.rs:280` todo!() | Removed 2026-06-30 | sibling session |

## Current in-flight gaps (2026-07-08)

| Gap | Priority | File(s) |
|-----|----------|---------|
| Legacy handlers still serve some routes (not migrated to hexagonal) | P2 | `backend/handlers.go` |
| Rust engine daemon not connected to Go backend UDS proxy | P2 | `crates/byteport-engine/src/bin/daemon.rs` |
| Parser unit tests missing | P2 | `nvms/projectManager/parser.go` |
| NVMS Spin SDK v2.2.0 incompatible with Go 1.26 | P3 | `nvms/vendor/` — blocks `go test ./nvms/...` |
| `byteport-transport` workspace dep conflict | P3 | `crates/byteport-transport/Cargo.toml` |
| Graceful shutdown added (2026-07-08) — needs acceptance test | P2 | `backend/main.go` |

## Key subsystems

| Subsystem | Lines | Tests | Status |
|-----------|-------|-------|--------|
| Domain entities (deployment, project, org, user) | ~500 | 8 files | ✅ stable |
| Application use cases (5) | ~350 | integral | ✅ stable |
| Infrastructure repos (postgres, memory) | ~600 | 6 files | ✅ stable |
| HTTP middleware (auth, CSRF, rate-limit, RBAC, OTel) | ~600 | 12 files | ✅ stable |
| Secrets manager (AWS+Vault+Env) | 719 | 5 files | ✅ stable |
| NVMS project manager + parser | ~500 | 0 | ⚠️ needs unit tests |
| NVMS lib (s3, ec2, network, auth, providers) | ~2,000 | 7 (sdk2) | ⚠️ Spin SDK version lock |
| byteport-engine (Rust) | 423 + lib | 28 | ✅ building |
| GitHub webhook handler | 263 | 8 tests | ✅ stable |
| Per-user credential scoping | 120 + test | 6 tests | ✅ stable |

## How to read this repo

- **`README.md`** — project pitch, quickstart, entry point for new contributors
- **`SPEC.md`** — canonical technical spec
- **`ARCHITECTURE.md`** — component boundaries and integration points
- **`PLAN.md`** — current roadmap (2026-07-08 refresh)
- **`CHARTER.md`** — project mission, tenets, scope, success criteria
- **`PRD.md`** — epics + stories
- **`stub-inventory.md`** — current open stubs and TODOs

## Out of scope (v1.0)

- Multi-cloud (Azure, GCP) — tracked as a non-goal in `PRD.md`
- Custom domain management
- Billing / cost management UI
- WebAuthn / passkey auth (post-v1.0)
- Rollback + redeploy + preview-env endpoints (post-v1.0)
