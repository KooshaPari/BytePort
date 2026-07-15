# BytePort — PLAN.md

> Last updated: 2026-07-08. Reflects the current hexagonal architecture
> (`internal/domain`, `internal/application`, `internal/infrastructure`).
> Supersedes the 2026-06-12 plan (which described a dead codebase with
> GORM `Find().Where()` bugs, a stubbed `deploy.go:51`, and commented-out
> NVMS auth — all resolved by sibling sessions).

## Status legend

- ☐ not started
- ◧ in progress
- ☑ done (PR merged)
- ⊘ deferred / non-goal

---

## Architecture overview (current)

```
┌─────────────────────────────────────────────────────┐
│                    HTTP Router (server.go)           │
│  Gin + CORS + OTel + Auth + CSRF + RateLimit + RBAC │
├─────────────────────────────────────────────────────┤
│                  /api/v1/ endpoints                   │
│  ┌────────────────────┐  ┌────────────────────────┐ │
│  │  Hexagonal (NEW)   │  │   Legacy (deprecating) │ │
│  │  /deployments      │  │  /legacy/deployments   │ │
│  │  /webhook          │  │  /projects /instances  │ │
│  └────────┬───────────┘  └────────────────────────┘ │
│           │                                          │
├───────────┼──────────────────────────────────────────┤
│           ▼                                          │
│  ┌────────────────────┐                              │
│  │    Container (DI)  │                              │
│  │  ┌──────────────┐  │                              │
│  │  │ Application  │  │  Use cases (create, list,    │
│  │  │ (use cases)  │  │  get, terminate, update)     │
│  │  ├──────────────┤  │                              │
│  │  │ Domain       │  │  Deployment entity,          │
│  │  │ (entity +    │  │  Repository port, Service,   │
│  │  │  port)       │  │  Status state machine        │
│  │  ├──────────────┤  │                              │
│  │  │Infrastructure│  │  Postgres repo, WorkOS auth, │
│  │  │ (adapters)   │  │  Secrets Mgr (AWS+Vault+Env),│
│  │  │              │  │  GitHub webhook, UDS proxy,  │
│  │  │              │  │  CSRF, RateLimit, RBAC, OTel │
│  │  └──────────────┘  │                              │
│  └────────────────────┘                              │
│                                                       │
│  ┌────────────────────┐  ┌────────────────────────┐  │
│  │ NVMS (Spin/Wasm)  │  │ Rust Cargo Workspace   │  │
│  │ /deploy /terminate │  │ byteport-engine (trait) │  │
│  │ Auth, LLM, EC2,   │  │ byteport-dag, -otel,   │  │
│  │ S3, ALB, Route53  │  │ -cli, -transport,      │  │
│  │                    │  │ phenotype-types        │  │
│  └────────────────────┘  └────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Key subsystems and their status

| Subsystem | Location | Status |
|---|---|---|
| Hexagonal deployment use cases | `backend/internal/application/deployment/` | ☑ 5 use cases implemented |
| Domain entity + repository port | `backend/internal/domain/deployment/` | ☑ Entity, Repository interface, Service, Status enum |
| Postgres repository adapter | `backend/internal/infrastructure/persistence/postgres/` | ☑ Full CRUD + mapper |
| WorkOS AuthKit integration | `backend/internal/infrastructure/auth/` | ☑ 3 test files |
| Credential validation | `backend/internal/infrastructure/clients/` | ☑ |
| GitHub webhook (HMAC-SHA256) | `backend/internal/infrastructure/http/handlers/webhook_handler.go` | ☑ 263 lines |
| CSRF middleware | `backend/internal/infrastructure/http/middleware/csrf.go` | ☑ env-gated |
| Rate-limit middleware | `backend/internal/infrastructure/http/middleware/rate_limit.go` | ☑ env-gated token bucket |
| RBAC middleware | `backend/internal/infrastructure/http/middleware/rbac.go` | ☑ |
| UDS proxy middleware | `backend/internal/infrastructure/http/middleware/uds_proxy.go` | ☑ but **not wired to Rust server** |
| OTel middleware | `backend/internal/infrastructure/otel/middleware.go` | ☑ env-gated |
| Secrets Manager (AWS+Vault+Env) | `backend/internal/infrastructure/secrets/manager.go` | ☑ 719 lines |
| Per-user credential scoping | `backend/internal/infrastructure/secrets/scoped.go` | ☑ 295 lines |
| EC2 adapter (real SDK-v2) | `backend/ec2_adapter.go` | ☑ replaces old `simulateDeployment()` |
| NVMS auth middleware | `backend/nvms/main.go:25-32` | ☑ **re-enabled** (`validateAction` wrapper) |
| NVMS manifest parser | `backend/nvms/projectManager/parser.go` | ☑ replaces `deploy.go:51` TODO |
| Graceful shutdown (SIGTERM) | `backend/main.go` | ☑ |
| byteport-engine Rust trait | `crates/byteport-engine/src/engine.rs` | ☑ Engine trait + Docker/Mock/NVMS adapters |
| byteport-dag | `crates/byteport-dag/` | ☑ DAG engine + scheduler + topo |
| byteport-otel | `crates/byteport-otel/` | ☑ tracing + metrics + propagation |
| byteport-cli | `crates/byteport-cli/` | ☑ config + telemetry + update_check |
| Legacy `handlers.go` (500 lines) | `backend/handlers.go` | ◧ still active at `/legacy/deployments` |
| Legacy GORM models | `backend/models/` | ◧ under hexagonal migration |
| Tests | 58 files across Go + Rust | ☑ extensive (models, domain, infra, handlers) |
| SvelteKit frontend | `frontend/web/` | ☐ not started |
| Tauri 2 desktop shell | `frontend/web/src-tauri/` | ◧ scaffold exists, Rust not wired |

---

## Phase 1 — Legacy retirement & migration complete (PR #1)

Migrate all legacy paths to hexagonal, delete dead code.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-001 | Migrate legacy deployment handlers from `handlers.go` to hexagonal use cases | ☐ | #1 |
| BP-DAG-002 | Delete legacy `DeploymentStore` and `handlers.go` deployment routes | ☐ | #1 |
| BP-DAG-003 | Migrate project/instance/org routes from `models/` to domain entities | ☐ | #1 |
| BP-DAG-004 | Remove `/legacy/deployments` prefix; serve hexagonal at canonical paths | ☐ | #1 |
| BP-DAG-005 | Clean up orphaned Go files (types.go, ec2_adapter.go → nvms adapters) | ☐ | #1 |
| BP-DAG-006 | Update `server.go` to reflect final route structure | ☐ | #1 |
| BP-DAG-007 | Verify 0 breakages: existing 58 test files all pass | ☐ | #1 |

---

## Phase 2 — Rust engine production wiring (PR #2)

The Engine trait + adapters exist. Wire them into the actual deployment pipeline.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-020 | `byteport-engine` daemon binary: long-lived process with HTTP/Unix socket | ☐ | #2 |
| BP-DAG-021 | Wire Docker adapter to real Docker daemon (integration test) | ☐ | #2 |
| BP-DAG-022 | Wire NVMS adapter (`crates/byteport-engine/src/adapters/nvms/http.rs`) to Spin NVMS | ☐ | #2 |
| BP-DAG-023 | `EngineRegistry` dispatch: HTTP endpoint or UDS handler selects adapter by name | ☐ | #2 |
| BP-DAG-024 | Wire `UDSProxy` middleware to the Rust engine daemon (replace no-op forward) | ☐ | #2 |
| BP-DAG-025 | `byteport-dag` integration: deployment pipeline as a DAG workflow | ☐ | #2 |
| BP-DAG-026 | Rust ↔ Go health check: `/healthz` reports engine daemon status | ☐ | #2 |
| BP-DAG-027 | `justfile` recipes: `cargo build -p byteport-engine --release` | ☐ | #2 |

---

## Phase 3 — Security & production hardening (PR #3)

Hardening that is not yet done (CSRF/rate-limit exist but need production tuning).

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-040 | Argon2id params audit (memory=64MiB iters=3 parallelism=2 salt=16B key=32B) | ☐ | #3 |
| BP-DAG-041 | PASETO v2→v3 audit (consider v4 public if API surface grows) | ☐ | #3 |
| BP-DAG-042 | Encryption key auto-rotate hook (placeholder) | ☐ | #3 |
| BP-DAG-043 | Secrets manager production audit: IAM roles, Vault token lifecycle | ☐ | #3 |
| BP-DAG-044 | SSRF allowlist for outbound API calls | ☐ | #3 |
| BP-DAG-045 | CORS allowlist final audit (env-driven, done but check coverage) | ☐ | #3 |
| BP-DAG-046 | `golangci.yml` strict (errcheck, govet, staticcheck, ineffassign) | ☐ | #3 |
| BP-DAG-047 | `go mod verify` in CI | ☐ | #3 |
| BP-DAG-048 | Coverage threshold enforcement (≥70% in CI) | ☐ | #3 |
| BP-DAG-049 | `GET /healthz` with full dependency check (DB, NVMS, engine daemon) | ◧ | #3 |

---

## Phase 4 — Rust CLI complete & SDK publish (PR #4)

The CLI exists but is not feature-complete for end users.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-060 | `byteport deploy` — read manifest, invoke engine daemon, stream logs | ☐ | #4 |
| BP-DAG-061 | `byteport status <id>` — poll deployment status from daemon | ☐ | #4 |
| BP-DAG-062 | `byteport list` — list active deployments | ☐ | #4 |
| BP-DAG-063 | `byteport stop <id>` — terminate deployment | ☐ | #4 |
| BP-DAG-064 | `byteport config` — manage secrets scoped to user | ☐ | #4 |
| BP-DAG-065 | `byteport validate` — offline manifest linting | ☐ | #4 |
| BP-DAG-066 | Publish `byteport-sdk` crate (phenotype-types + engine client) | ☐ | #4 |

---

## Phase 5 — SvelteKit frontend (PR #5)

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-120 | `/` landing | ☐ | #5 |
| BP-DAG-121 | `/login`, `/signup` (WorkOS AuthKit) | ☐ | #5 |
| BP-DAG-122 | `/authenticate` (silent re-auth) | ☐ | #5 |
| BP-DAG-123 | `/link` (GitHub OAuth + AWS + LLM creds) | ☐ | #5 |
| BP-DAG-124 | `/dashboard` (deployments list) | ☐ | #5 |
| BP-DAG-125 | `/deploy` wizard | ☐ | #5 |
| BP-DAG-126 | `/projects` + `/projects/[uuid]` | ☐ | #5 |
| BP-DAG-127 | Zod schemas + superforms for all forms | ☐ | #5 |
| BP-DAG-128 | Svelte 5 runes for all stores | ☐ | #5 |
| BP-DAG-129 | i18n (en + es), a11y, dark mode | ☐ | #5 |
| BP-DAG-130 | Error boundaries, vitest, Playwright e2e | ☐ | #5 |

---

## Phase 6 — Tauri 2 desktop shell (PR #6)

Scaffold exists at `frontend/web/src-tauri/`. Needs production wiring.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-160 | Tauri capabilities file | ☐ | #6 |
| BP-DAG-161 | Tauri shell plugin (open external URLs) | ☐ | #6 |
| BP-DAG-162 | Tauri fs plugin (manifest picker) | ☐ | #6 |
| BP-DAG-163 | Tauri http plugin (frontend→backend) | ☐ | #6 |
| BP-DAG-164 | Tauri deep-link plugin (`byteport://`) | ☐ | #6 |
| BP-DAG-165 | Tauri updater plugin (signed updates) | ☐ | #6 |
| BP-DAG-166 | macOS notarization + Windows code signing | ☐ | #6 |
| BP-DAG-167 | CSP lockdown on webview | ☐ | #6 |
| BP-DAG-168 | Tauri commands: 3 (`pick_manifest`, `read_secret`, `write_secret`) | ☐ | #6 |

---

## Phase 7 — CI/CD & quality gates (PR #7)

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-190 | `dependabot.yml` (Go, npm, cargo, GitHub Actions, Docker) | ☐ | #7 |
| BP-DAG-191 | `go-ci.yml` (vet + build + test + race + cover) | ☐ | #7 |
| BP-DAG-192 | `rust-ci.yml` (clippy + test + build) | ☐ | #7 |
| BP-DAG-193 | `nvms-ci.yml` (spin build + test) | ☐ | #7 |
| BP-DAG-194 | `npm-ci.yml` (lint + typecheck + test + build) | ☐ | #7 |
| BP-DAG-195 | `codeql.yml` (Go + TS + Rust) | ☐ | #7 |
| BP-DAG-196 | `release.yml` (signed artifacts) | ☐ | #7 |
| BP-DAG-197 | `trufflehog.yml` (already exists, keep) | ☐ | #7 |
| BP-DAG-198 | `sbom.yml` (CycloneDX) | ☐ | #7 |
| BP-DAG-199 | Quality gate (single-source-of-truth, coverage, lint) | ☐ | #7 |

---

## Phase 8 — Dev orchestration & documentation (PR #8)

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-220 | `docker-compose.yaml` (full local stack: backend + postgres + engine) | ☐ | #8 |
| BP-DAG-221 | `Dockerfile` (backend Go binary) | ☐ | #8 |
| BP-DAG-222 | `setup-unix.sh` (one-command dev bootstrap) | ☐ | #8 |
| BP-DAG-240 | `docs/getting-started.md` | ☐ | #8 |
| BP-DAG-241 | `docs/architecture.md` (supersedes current ARCHITECTURE.md) | ☐ | #8 |
| BP-DAG-242 | `docs/api.md` (auto-generated from OpenAPI or oapi-codegen) | ☐ | #8 |
| BP-DAG-243 | `docs/security.md` | ☐ | #8 |

---

## Phase 9 — Verification matrix (fan-in, release gate)

| ID | Gate | Command | Required |
|---|---|---|---|
| BP-DAG-260 | Go race | `go test -race ./backend/...` | clean |
| BP-DAG-261 | Go vet | `go vet ./backend/...` | 0 warnings |
| BP-DAG-262 | golangci-lint | `golangci-lint run` | 0 errors |
| BP-DAG-263 | govulncheck | `govulncheck ./backend/...` | no known vulns |
| BP-DAG-264 | Rust test | `cargo test --workspace` | all pass |
| BP-DAG-265 | Cargo clippy | `cargo clippy -- -D warnings` | 0 errors |
| BP-DAG-266 | Cargo fmt | `cargo fmt --check` | clean |
| BP-DAG-267 | SvelteKit check | `npm run check` | 0 errors |
| BP-DAG-268 | Trufflehog | `trufflehog filesystem .` | 0 secrets |
| BP-DAG-269 | CodeQL | `codeql analyze` | 0 alerts |
| BP-DAG-270 | Coverage threshold | `go test -cover ./backend/...` | ≥70% |
| BP-DAG-271 | E2E smoke | `make e2e-smoke` | green |
| BP-DAG-272 | `go mod verify` | `go mod verify` | pass |

---

## Critical-path graph (to v1.0.0)

```
Phase 1 (PR #1) ───→ Phase 7 (CI/CD)
                        │
Phase 2 (PR #2) ───────┤
                        │
Phase 3 (PR #3) ───────┼──→ Phase 9 (Verification Matrix) → RELEASE v1.0.0
                        │
Phase 4 (PR #4) ───────┤
                        │
Phase 5 (PR #5) ───────┤
                        │
Phase 6 (PR #6) ───────┤
                        │
Phase 8 (PR #8) ───────┘
```

Phases 1–4 can run in parallel (different code owners). Phases 5–6 depend on
Phase 1 completing (frontend consumes hexagonal API, not legacy routes).
Phase 7 is cross-cutting and should start early. Phase 8 is documentation.

---

## Resources

| Role | Allocation |
|---|---|
| Backend Engineer (Go + Rust) | 1 FTE |
| Frontend Engineer (Svelte/TS) | 0.5 FTE |
| Desktop Engineer (Rust/Tauri) | 0.25 FTE |
| DevOps / CI | 0.25 FTE |

(FTE = full-time-equivalent at current solo/duo pace; many tasks run in parallel.)

---

## Success criteria (v1.0.0)

- [ ] Legacy `handlers.go` deployment routes deleted; all traffic through hexagonal API
- [ ] `byteport-engine` daemon running alongside Go backend with UDS communication
- [ ] A new user with `odin.nvms` + GitHub repo gets a live URL in <15 min
- [ ] All verification gates in Phase 9 are green
- [ ] No `TODO:` markers remain in production paths
- [ ] 58 existing tests + new coverage maintain ≥70% code coverage
- [ ] CSRF + rate-limit + RBAC all enforced in production configuration
