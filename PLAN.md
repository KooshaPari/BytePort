# BytePort — PLAN.md

> Last updated: 2026-06-12. Replaces the prior NanoVMS-era phases (SpinCLI,
> MicroVM images, Puppeteer, S3+CloudFront portfolio hosting) that were never
> implemented. This is the v1.0 roadmap for the current shipping stack
> (Go/Gin/GORM/SQLite/PASETO/SvelteKit/Tauri 2 + Spin nvms).

## Status legend

- ☐ not started
- ◧ in progress
- ☑ done (PR merged)
- ⊘ deferred / non-goal

---

## Phase 0 — Governance Reset (PR #1, in progress)

The 4 contradicting identity docs at the repo root are rewritten against
current reality. This is the foundation for every later phase.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-001 | Rewrite `STATUS.md` against current stack | ◧ | #1 |
| BP-DAG-002 | Rewrite `CHARTER.md` against current product | ◧ | #1 |
| BP-DAG-003 | Rewrite `PLAN.md` (this file) | ◧ | #1 |
| BP-DAG-004 | Trim `README.md` to current reality (keep quickstart, retire the Loco.rs manifesto) | ◧ | #1 |

---

## Phase 1 — Security & Reliability Floor (PR #2, queued)

The 4 critical bugs in the eval. Each is a 1–3 line fix.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-040 | Re-audit `backend/byteport/routes/*.go` for `Find().Where(...)` patterns | ☐ | #2 |
| BP-DAG-041 | Add transaction + idempotency to `routes/pm.go::addNewProject` | ☐ | #2 |
| BP-DAG-042 | Fix owner scoping in `routes/instances.go:12` (`Where().Find()`, not `Find().Where()`) | ☐ | #2 |
| BP-DAG-043 | Fix owner scoping in `routes/projects.go:12` (same pattern) | ☐ | #2 |
| BP-DAG-044 | Add `return` after every error-branch `c.JSON` in `routes/deployment.go` (5 sites) | ☐ | #2 |
| BP-DAG-322 | Parameterize `http://localhost:3000/deploy` via env var in `deployment.go:30` | ☐ | #2 |
| BP-DAG-323 | Add request timeout + retry to the `nvms` call | ☐ | #2 |
| BP-DAG-324 | Re-enable auth middleware in `backend/nvms/main.go:25-35` (uncomment + fix import) | ☐ | #2 |
| BP-DAG-040b | Don't exit on empty `git_secrets` in `backend/byteport/main.go:104-113` (cold-start path) | ☐ | #2 |
| BP-DAG-040c | Set `SameSite=Lax` and pin cookie domain in `routes/auth.go::setAuthCookie` | ☐ | #2 |

---

## Phase 2 — Manifest Engine (PR #3, planned)

The core product feature — actually parsing `odin.nvms`. Without this, the
`/deploy` endpoint has nothing to do.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-020 | Define canonical `odin.nvms` schema (odin's idl-ish format) | ☐ | #3 |
| BP-DAG-021 | JSON-Schema for `odin.nvms` | ☐ | #3 |
| BP-DAG-022 | Go parser in `backend/byteport/lib/manifest.go` (yaml.v3) | ☐ | #3 |
| BP-DAG-023 | Unit tests for happy path | ☐ | #3 |
| BP-DAG-024 | Unit tests for schema violations (3 cases) | ☐ | #3 |
| BP-DAG-025 | Replace `deploy.go:51` TODO with real `UnmarshalYAML` (BP-DAG-325) | ☐ | #3 |
| BP-DAG-026 | CLI: `byteport validate` for offline manifest linting | ☐ | #3 |
| BP-DAG-027 | `docs/manifest.md` with full schema reference | ☐ | #3 |

---

## Phase 3 — Backend Hardening (PR #4, planned)

Beyond the bug fixes in Phase 1.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-045 | Wire `lib.AuthMiddleware` into every protected route (sanity check) | ☐ | #4 |
| BP-DAG-046 | Switch auth cookie to `httpOnly`, `Secure`, `SameSite=Lax` | ☐ | #4 |
| BP-DAG-047 | Argon2id params audit (memory=64MiB iters=3 parallelism=2 salt=16B key=32B) | ☐ | #4 |
| BP-DAG-048 | PASETO v2 → v3 audit (consider v4 public if API surface grows) | ☐ | #4 |
| BP-DAG-049 | Encryption key auto-rotate hook (placeholder for v1.0) | ☐ | #4 |
| BP-DAG-050 | `lib/apilink.go` SSRF allowlist allowlist-of-allowlists | ☐ | #4 |
| BP-DAG-051 | OpenAI key validation: tighten to single `GET /v1/models` call (no retries) | ☐ | #4 |
| BP-DAG-052 | slog JSON output | ☐ | #4 |
| BP-DAG-053 | OTel OTLP exporter (replace ConsoleSpanExporter) | ☐ | #4 |
| BP-DAG-054 | CORS allowlist (env-driven) | ☐ | #4 |
| BP-DAG-055 | Rate limiter (per-IP, per-route) | ☐ | #4 |
| BP-DAG-056 | `GET /healthz` (DB ping + nvms ping) | ☐ | #4 |
| BP-DAG-057 | Graceful shutdown (SIGTERM drains in-flight requests) | ☐ | #4 |
| BP-DAG-058 | `golangci.yml` strict (errcheck, govet, staticcheck, ineffassign) | ☐ | #4 |
| BP-DAG-059 | `justfile` (dev, build, test, lint, smoke) | ☐ | #4 |
| BP-DAG-060 | `go mod verify` in CI | ☐ | #4 |

---

## Phase 4 — NVMS Service Completion (PR #5, planned)

Get the Spin module to v1.0 with auth, manifest, LLM, and observability.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-080 | `spin.toml` audit (manifest, config) | ☐ | #5 |
| BP-DAG-081 | Spin CLI install instructions in `README.md` | ☐ | #5 |
| BP-DAG-082 | MicroVM lifecycle (create / start / stop / delete) | ☐ | #5 |
| BP-DAG-083 | NVMS auth middleware (port 3000) — replaces commented code | ☐ | #5 |
| BP-DAG-084 | `POST /deploy` end-to-end (manifest → MicroVM → live URL) | ☐ | #5 |
| BP-DAG-085 | `POST /terminate` end-to-end | ☐ | #5 |
| BP-DAG-086 | NVMS YAML unmarshalling (BP-DAG-325) | ☐ | #5 |
| BP-DAG-087 | NVMS manifest validation against schema (BP-DAG-021) | ☐ | #5 |
| BP-DAG-088 | NVMS error responses with stable codes | ☐ | #5 |
| BP-DAG-089 | LLM provider: OpenAI (production) | ☐ | #5 |
| BP-DAG-090 | LLM provider: local (Ollama) | ☐ | #5 |
| BP-DAG-091 | LLM provider: Gemini (BP-DAG-326) | ☐ | #5 |
| BP-DAG-092 | NVMS OTel spans | ☐ | #5 |
| BP-DAG-093 | NVMS `/metrics` (Prometheus) | ☐ | #5 |
| BP-DAG-094 | NVMS integration tests (real Spin runner) | ☐ | #5 |
| BP-DAG-095 | NVMS smoke test (`go test ./backend/nvms/...`) | ☐ | #5 |
| BP-DAG-096 | Clean up `backend/nvms.rs:280 todo!()` (BP-DAG-327) | ☐ | #5 |
| BP-DAG-097 | NVMS ARCHITECTURE.md in `backend/nvms/README.md` | ☐ | #5 |

---

## Phase 5 — SvelteKit Frontend (PR #6, planned)

11 routes, zod schemas, superforms, runes, i18n, a11y, dark mode, error boundaries.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-120 | `/` (landing) | ☐ | #6 |
| BP-DAG-121 | `/signup` | ☐ | #6 |
| BP-DAG-122 | `/login` | ☐ | #6 |
| BP-DAG-123 | `/authenticate` (silent re-auth) | ☐ | #6 |
| BP-DAG-124 | `/link` (GitHub OAuth) | ☐ | #6 |
| BP-DAG-125 | `/link` (POST: AWS + LLM + Portfolio creds) | ☐ | #6 |
| BP-DAG-126 | `/home` (dashboard) | ☐ | #6 |
| BP-DAG-127 | `/projects` | ☐ | #6 |
| BP-DAG-128 | `/projects/[uuid]` | ☐ | #6 |
| BP-DAG-129 | `/deploy` (wizard) | ☐ | #6 |
| BP-DAG-130 | `/instances` (BP-DAG-130 superset — 11 sub-tasks) | ☐ | #6 |
| BP-DAG-131 | zod schemas for all forms | ☐ | #6 |
| BP-DAG-132 | sveltekit-superforms wiring | ☐ | #6 |
| BP-DAG-133 | Svelte 5 runes for all stores | ☐ | #6 |
| BP-DAG-134 | i18n (en + es) | ☐ | #6 |
| BP-DAG-135 | a11y audit (axe-core) | ☐ | #6 |
| BP-DAG-136 | Dark mode toggle | ☐ | #6 |
| BP-DAG-137 | Error boundaries (route-level + root) | ☐ | #6 |
| BP-DAG-138 | Storybook 10 stories for all components | ☐ | #6 |
| BP-DAG-139 | vitest unit tests | ☐ | #6 |
| BP-DAG-140 | Playwright e2e tests | ☐ | #6 |
| BP-DAG-141 | prettier + eslint | ☐ | #6 |

---

## Phase 6 — Tauri 2 Desktop Shell (PR #7, planned)

Bundles the SvelteKit frontend as a desktop/mobile app.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-160 | Tauri capabilities file | ☐ | #7 |
| BP-DAG-161 | Tauri shell plugin (open external URLs) | ☐ | #7 |
| BP-DAG-162 | Tauri fs plugin (sandboxed read for manifest picker) | ☐ | #7 |
| BP-DAG-163 | Tauri http plugin (frontend → backend) | ☐ | #7 |
| BP-DAG-164 | Tauri deep-link plugin (e.g. `byteport://`) | ☐ | #7 |
| BP-DAG-165 | Tauri updater plugin (signed updates) | ☐ | #7 |
| BP-DAG-166 | Windows code signing | ☐ | #7 |
| BP-DAG-167 | macOS notarization (BP-DAG-167 superset — 4 sub-tasks) | ☐ | #7 |
| BP-DAG-168 | CSP lockdown on the webview | ☐ | #7 |
| BP-DAG-169 | Tauri commands: 3 (e.g. `pick_manifest`, `read_secret`, `write_secret`) | ☐ | #7 |
| BP-DAG-170 | Splash screen | ☐ | #7 |
| BP-DAG-171 | Tray icon | ☐ | #7 |
| BP-DAG-172 | `cargo test` for src-tauri | ☐ | #7 |

---

## Phase 7 — CI/CD (PR #8, planned)

Replace the 20-workflow sprawl with a clean set.

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-190 | `dependabot.yml` (Go, npm, cargo, GitHub Actions, Docker) | ☐ | #8 |
| BP-DAG-191 | `release-drafter.yml` | ☐ | #8 |
| BP-DAG-192 | `trufflehog.yml` (already fixed, keep) | ☐ | #8 |
| BP-DAG-193 | Rewrite `go-ci.yml` (vet + build + test + race + cover) | ☐ | #8 |
| BP-DAG-194 | `npm-ci.yml` (lint + typecheck + test + build) | ☐ | #8 |
| BP-DAG-195 | `tauri-ci.yml` (cargo test + signed build) | ☐ | #8 |
| BP-DAG-196 | `nvms-ci.yml` (spin build + test) | ☐ | #8 |
| BP-DAG-197 | `release.yml` (signed artifacts) | ☐ | #8 |
| BP-DAG-198 | `codeql.yml` (Go + TS + Rust) | ☐ | #8 |
| BP-DAG-199 | `fr-coverage.yml` (FR → test traceability) | ☐ | #8 |
| BP-DAG-200 | `quality-gate.yml` (single-source-of-truth) | ☐ | #8 |
| BP-DAG-201 | `cve-monitor.yml` (osv-scanner weekly) | ☐ | #8 |
| BP-DAG-202 | `sbom.yml` (CycloneDX) | ☐ | #8 |
| BP-DAG-203 | Reusable cache workflow | ☐ | #8 |
| BP-DAG-204 | CODEOWNERS enforcement | ☐ | #8 |

---

## Phase 8 — Dev Orchestration & Onboarding (PR #9, planned)

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-220 | Parameterize `./start` (no hardcoded `~/temp-PRODVERCEL/...` path) | ☐ | #9 |
| BP-DAG-221 | `./start.bat` (Windows parity) | ☐ | #9 |
| BP-DAG-222 | `setup-windows.ps1` (existing, audit) | ☐ | #9 |
| BP-DAG-223 | `setup-unix.sh` (new) | ☐ | #9 |
| BP-DAG-224 | `docker-compose.yaml` (full local stack) | ☐ | #9 |
| BP-DAG-225 | `Dockerfile` (backend) | ☐ | #9 |
| BP-DAG-226 | `Dockerfile.frontend` | ☐ | #9 |

---

## Phase 9 — Documentation (PR #10, planned)

| ID | Task | Status | PR |
|---|---|---|---|
| BP-DAG-240 | `docs/getting-started.md` | ☐ | #10 |
| BP-DAG-241 | `docs/api.md` (auto-generated from `swag` or `oapi-codegen`) | ☐ | #10 |
| BP-DAG-242 | `docs/architecture.md` (long-form, supersedes current ARCHITECTURE.md) | ☐ | #10 |
| BP-DAG-243 | `docs/security.md` | ☐ | #10 |
| BP-DAG-244 | `docs/release-process.md` | ☐ | #10 |
| BP-DAG-245 | `docs/troubleshooting.md` | ☐ | #10 |

---

## Phase 10 — Verification Matrix (fan-in, end of v1.0)

| ID | Gate | Command | Required |
|---|---|---|---|
| BP-DAG-260 | Go race | `go test -race ./backend/...` | clean |
| BP-DAG-261 | Go vet | `go vet ./backend/...` | 0 warnings |
| BP-DAG-262 | golangci-lint | `golangci-lint run` | 0 errors |
| BP-DAG-263 | govulncheck | `govulncheck ./backend/...` | no known vulns |
| BP-DAG-264 | SvelteKit check | `npm run check` (web) | 0 errors |
| BP-DAG-265 | npm audit | `npm audit` | 0 high/critical |
| BP-DAG-266 | Cargo test | `cargo test` (src-tauri) | all pass |
| BP-DAG-267 | Tauri signed build | `cargo tauri build` | signed artifact |
| BP-DAG-268 | Cargo clippy | `cargo clippy -- -D warnings` | 0 errors |
| BP-DAG-269 | Cargo fmt | `cargo fmt --check` | clean |
| BP-DAG-270 | osv-scanner | `osv-scanner --recursive .` | clean |
| BP-DAG-271 | trufflehog | `trufflehog filesystem .` | 0 secrets |
| BP-DAG-272 | codeql | `codeql analyze` | 0 alerts |
| BP-DAG-273 | SBOM | `cyclonedx-bom` | generated |
| BP-DAG-274 | E2E smoke | `make e2e-smoke` | green |
| BP-DAG-275 | FR coverage | `fr-coverage.yml` | 100% |
| BP-DAG-276 | Coverage threshold | `go test -cover` | ≥70% |
| BP-DAG-277 | A11y audit | `axe` | 0 criticals |
| BP-DAG-278 | Storybook build | `npm run build-storybook` | clean |
| BP-DAG-279 | Signed artifacts | `cosign verify-blob` | pass |
| BP-DAG-280 | odin.nvms round-trip | fixture | green |
| BP-DAG-281 | go mod verify | `go mod verify` | pass |

---

## Cross-cutting governance (Phase 11)

| ID | Task | Status |
|---|---|---|
| BP-DAG-300 | Final pass: `STATUS.md` reflects shipped v1.0 | ☐ |
| BP-DAG-301 | Final pass: `CHARTER.md` reflects shipped v1.0 | ☐ |
| BP-DAG-302 | Final pass: `PLAN.md` is empty (all phases done) | ☐ |
| BP-DAG-303 | Final pass: `SPEC.md` reflects shipped v1.0 | ☐ |
| BP-DAG-304 | Final pass: `PRD.md` epics all done | ☐ |
| BP-DAG-305 | Final pass: `README.md` 60-second quickstart | ☐ |

---

## Critical-path ASCII graph (to v1.0.0)

```
Phase 0 (PR #1) → Phase 1 (PR #2) → Phase 2 (PR #3) ─┐
                                  → Phase 3 (PR #4) ─┤
                                  → Phase 4 (PR #5) ─┤
                                  → Phase 5 (PR #6) ─┼→ Phase 7 → Phase 8 → Phase 9 → Phase 10 → RELEASE v1.0.0
                                  → Phase 6 (PR #7) ─┘
```

---

## Resources

| Role | Allocation |
|------|------------|
| Backend Engineer (Go) | 1 FTE |
| Frontend Engineer (Svelte/TS) | 0.5 FTE |
| Desktop Engineer (Rust/Tauri) | 0.25 FTE |
| DevOps / CI | 0.25 FTE |

(FTE = full-time-equivalent at the current solo/duo pace; many tasks can run in parallel.)

---

## Success Criteria (v1.0.0)

- [ ] `./start dev` brings up the full stack in <10s
- [ ] A new user with `odin.nvms` + GitHub repo gets a live URL in <15 min
- [ ] All 23 verification gates in Phase 10 are green
- [ ] No `todo!()` or `TODO:` markers remain in production paths
- [ ] `fr-coverage.yml` shows 100% FR-to-test mapping
- [ ] All 4 critical security bugs from the 2026-06-12 eval are fixed
- [ ] All 3 TODO stubs from `stub-inventory.md` are resolved
