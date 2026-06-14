# BytePort — Status

> Last updated: 2026-06-12 — current source of truth for what BytePort is and where it stands.
> Supersedes all prior "BytePort is a hardware ledger" or "BytePort is a data transport layer"
> narratives from earlier revisions of `CHARTER.md`, `STATUS.md`, and `PLAN.md`.

## What BytePort is

BytePort is a **self-hosted Infrastructure-as-Code deployment + portfolio UX generation
platform**. Developers define one manifest (`odin.nvms`) at their repo root; BytePort
provisions a MicroVM-backed deployment, registers the resulting endpoints with a portfolio
site, and uses an LLM to generate showcase metadata for each project.

| Component | Language | Status | Purpose |
|---|---|---|---|
| `backend/byteport/` | Go 1.25 | building | Gin + GORM + SQLite, PASETO auth, AWS SDK, NVMS proxy |
| `backend/nvms/` | Go 1.25 | building | Spin / Fermyon wasm module — deploy/terminate HTTP API on port 3000 |
| `frontend/web/` | SvelteKit 2 / Svelte 5 / Tailwind 4 | building | Admin UI (signup, login, link, projects, instances, deploy) |
| `frontend/web/src-tauri/` | Rust + Tauri 2 | building | Desktop/mobile shell wrapping the SvelteKit frontend |

## Current health

| Gate | Status | Notes |
|---|---|---|
| `go vet ./backend/byteport/...` | green | 0 errors |
| `go build ./backend/byteport/...` | green | 0 errors |
| `go test ./backend/byteport/...` | green | floor test (`smoke_test.go`) passes |
| `cargo check --all-targets` (src-tauri) | n/a | not yet wired into CI |
| `npm run check` (frontend) | not run here | SvelteKit 2 + svelte-check |

## Known in-flight gaps (2026-06-12)

The 4 governance files in this folder were rewritten in PR #1 to match reality.
The following gaps remain in code and are tracked under `stub-inventory.md`
and the BytePort DAG plan at `C:\Users\koosh\plans\2026-06-12-byteport-exhaustive-plan-v2.md`.

1. `backend/byteport/routes/instances.go:12` — `Find().Where(...)` drops the owner filter
2. `backend/byteport/routes/projects.go:12` — same `Find().Where(...)` bug as instances
3. `backend/byteport/routes/deployment.go` — 5 sites call `c.JSON(...)` after an error
   branch without `return`, so the handler keeps running after the response is sent
4. `backend/nvms/main.go:25-35` — `validateAction` has its auth middleware commented
   out, so anyone reachable on port 3000 can deploy/terminate any project
5. `backend/nvms/projectManager/deploy.go:51` — NVMS YAML unmarshalling is a TODO
6. `backend/nvms/lib/llm.go:22` — `ProviderGemini` is a stub
7. `backend/nvms.rs:280` — `todo!()` in `locateNVMS()` (unreachable in current build,
   tracked for cleanup)

Items 1–4 are security/reliability and are queued for **PR #2**.
Items 5–7 are feature gaps and are queued for the stub-remediation pass described
in `stub-inventory.md`.

## How to read this repo

- **`README.md`** — project pitch, quickstart, entry point for new contributors
- **`SPEC.md`** — canonical technical spec (Go/Gin/GORM/SQLite/PASETO/SvelteKit/Tauri 2)
- **`ARCHITECTURE.md`** — component boundaries and integration points
- **`FUNCTIONAL_REQUIREMENTS.md`** — the 20 FRs grouped by capability, traced to epics in `PRD.md`
- **`PLAN.md`** — current v1.0 roadmap (replaces the old NanoVMS-era phases)
- **`CHARTER.md`** — project mission, tenets, scope, success criteria
- **`PRD.md`** — epics + stories
- **`SPECS_INDEX.md`** — auto-generated audit index of CI workflows and FR coverage
- **`stub-inventory.md`** — current open stubs and TODOs

## Out of scope (v1.0)

- Multi-cloud (Azure, GCP) — tracked as a non-goal in `PRD.md`
- Custom domain management
- Billing / cost management UI
- WebAuthn / passkey auth (post-v1.0)
- Rollback + redeploy + preview-env endpoints (post-v1.0)
