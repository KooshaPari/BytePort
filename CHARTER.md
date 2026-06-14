# BytePort — Charter

> Last updated: 2026-06-12. Replaces the prior "data transport / byte-stream"
> charter that described a product which was never built. BytePort today is an
> Infrastructure-as-Code deployment + portfolio UX generation platform.

---

## 1. Mission Statement

**BytePort** turns a single `odin.nvms` manifest into a deployed, portfolio-worthy
project. The developer writes one declarative file describing their application's
services and infrastructure; BytePort provisions the deployment via the `nvms`
MicroVM runtime, registers the resulting endpoints with the developer's portfolio
site, and uses an LLM to generate showcase metadata for each project.

The project exists to remove the operational friction between "code committed to
GitHub" and "live URL visible to the world on a portfolio page" — with one
declarative file as the source of truth, and one command (`./start dev` for local,
`POST /deploy` for production) to drive the whole pipeline.

---

## 2. Tenets (Unless You Know Better Ones)

### Tenet 1: One Manifest, One Source of Truth

The `odin.nvms` file is the contract. Multi-service apps, runtime selection, port
mapping, environment injection, and (optionally) portfolio/AI configuration all
live there. CLI flags and dashboard inputs are conveniences that populate the
manifest — they are not parallel sources of truth.

### Tenet 2: Deploys Are Reproducible

A `odin.nvms` plus a Git ref should produce the same deployment every time.
NVMS records what it deployed; re-running with the same inputs converges to
the same state. Implies: no hidden state in the dashboard, no implicit env
vars, no "remember me" deploy options that aren't in the manifest.

### Tenet 3: Portfolio Is First-Class, Not An Afterthought

Every deployed project gets a portfolio entry — endpoint URL, description, tags,
LLM-enhanced metadata. The portfolio endpoint is configured alongside the deploy
manifest, not bolted on after the fact. If you don't want a portfolio entry,
opt out explicitly in the manifest.

### Tenet 4: Local Dev == Production

`./start dev` (tmux session: SvelteKit dev server on 5173, Go backend on 8081
with `air` hot-reload, `nvms` on 3000) should be functionally identical to
the production stack modulo AWS resources. If a feature works in dev and
breaks in prod, prod is the bug.

### Tenet 5: Pluggable LLM, No Lock-In

The LLM is an implementation detail. `OpenAI` and `local` are first-class;
`gemini` is stubbed for v1.0. Provider config lives in the manifest. Swapping
providers is a config change, not a refactor.

### Tenet 6: Self-Hosted, Not Cloud-Locked

BytePort runs on the developer's machine or a single VM. AWS is the deploy
target, not the runtime host. There is no BytePort-hosted dashboard. There
is no implicit outbound telemetry beyond OTel traces the user wires up.

### Tenet 7: SSRF-Safe by Default

Any HTTP client BytePort uses to validate a user-supplied URL (portfolio
endpoint, AI endpoint, webhooks) must use the SSRF-safe client in
`lib/apilink.go`. Direct `http.Get` against user input is a bug, not a
shortcut. Loopback, private, link-local, and multicast addresses are
rejected by default; allowlist overrides are explicit and logged.

### Tenet 8: Encrypted at Rest, Encrypted in Transit

All credentials — AWS keys, GitHub tokens, LLM keys, portfolio keys — are
AES-256-CFB encrypted with a key derived from an auto-generated master.
Passwords are Argon2id. Sessions are PASETO tokens in httpOnly cookies.
There is no plaintext credential path.

---

## 3. Scope & Boundaries

### In Scope (v1.0)

- **Core Deploy Engine** (Go 1.25, Gin, GORM, SQLite):
  - Manifest parsing and validation (`backend/byteport/lib/manifest.go` — TBD)
  - Project + Instance + Repository models with GORM automigrate
  - AWS SDK integration for EC2 / S3 / IAM provisioning
  - LLM-credentialed portfolio description generation
- **NVMS Runtime** (Go 1.25, Spin / Fermyon wasm):
  - HTTP API on port 3000: `POST /`, `POST /deploy`, `POST /terminate`
  - Auth middleware (currently disabled — tracked for PR #2)
  - NVMS YAML manifest unmarshalling (currently TODO — tracked for PR #2)
- **Frontend** (SvelteKit 2, Svelte 5, Tailwind 4):
  - Auth (signup, login, link, authenticate)
  - Project list, instance list, deploy wizard
  - GitHub repo picker
- **Desktop Shell** (Rust + Tauri 2):
  - Bundles the SvelteKit frontend
  - Code signing for Windows / macOS
  - CSP lockdown on the webview

### Out of Scope (v1.0)

- Multi-cloud (Azure, GCP) — see `PRD.md` Non-Goals
- Custom domain management
- Billing / cost management UI
- WebAuthn / passkey auth
- Rollback / redeploy / preview-env endpoints
- `process-compose` integration (the `nvms` runtime is Spin-only in v1.0)

### Boundaries

- BytePort deploys projects; it does not interpret their traffic
- The portfolio site is configured but not hosted by BytePort
- LLM providers are called but not trained on BytePort state
- The `nvms` runtime is a separate Spin wasm module with its own go.mod

---

## 4. Target Users & Personas

### Primary Persona: Indie Developer Indra

**Role:** Solo or small-team developer shipping portfolio projects  
**Goals:** Push to GitHub → get a live URL → add to portfolio with zero ops  
**Pain Points:** Docker, Kubernetes, Terraform, AWS console, all for a side project  
**Needs:** One-file deploy, LLM-enhanced portfolio text, a portfolio site URL  
**Tech Comfort:** High (can read YAML, run a Go binary, set env vars)

### Secondary Persona: Portfolio Maintainer Pia

**Role:** Developer with a `kooshapari.com`-style portfolio site  
**Goals:** Every new project on the portfolio without manual editing  
**Pain Points:** Portfolio site upkeep, copy-pasting descriptions, manual screenshots  
**Needs:** Auto-generated portfolio entries, descriptions, endpoint widgets  
**Tech Comfort:** High (runs a portfolio backend, e.g. `Slickport`)

### Tertiary Persona: Cloud Skeptic Sam

**Role:** Engineer who wants self-hosted IaC, not Vercel/Railway  
**Goals:** Single-VM deploy of their IaC engine  
**Pain Points:** Vendor lock-in, opaque cloud billing, surprise bills  
**Needs:** Local-first, deterministic, OpenTelemetry-visible  
**Tech Comfort:** Very high (reads Go, runs containers, sets up OTLP exporters)

---

## 5. Success Criteria (Measurable)

### Functionality

- **One-command deploy**: `POST /deploy` with a `odin.nvms` returns a live URL in <5 min
- **Multi-service manifest**: A manifest with N services produces N instances tracked in DB
- **Portfolio entry on success**: Each successful deploy emits a portfolio endpoint call
- **Self-hostable end-to-end**: `./start dev` brings up the whole stack on one machine

### Security

- **Zero plaintext credentials on disk**: AES-256-CFB everywhere, Argon2id for passwords
- **SSRF-safe credential validation**: All user-supplied URLs routed through `apilink.go`
- **Cookie hardening**: `httpOnly`, `Secure`, `SameSite=Lax` on auth cookies
- **Auth on `nvms`**: Auth middleware enabled on port 3000 (currently off — PR #2)

### Performance

- **Local dev startup**: <10s for `./start dev` (Go + SvelteKit + Spin)
- **Deploy time (hello-world)**: <90s end-to-end against a local `nvms` instance
- **API p95**: <50ms for `/authenticate`, `/projects`, `/instances`

### Developer Experience

- **First-deploy walkthrough**: README quickstart gets a new user to a live URL in <15 min
- **CLI errors are loud**: stderr + non-zero exit, no silent failures
- **Quality gates in CI**: `go vet`, `go build`, `go test`, `golangci-lint`, `npm run check`

---

## 6. Governance Model

### Component Organization

```
BytePort/
├── backend/
│   ├── byteport/        # Core API: Go 1.25, Gin, GORM, SQLite
│   └── nvms/            # NVMS runtime: Go 1.25, Spin wasm, port 3000
├── frontend/
│   └── web/             # SvelteKit 2 + Tauri 2 shell
├── docs/                # Auto-generated and hand-written long-form docs
├── .github/workflows/   # CI: go-ci, npm-ci, tauri-ci, nvms-ci, codeql, etc.
├── start                # tmux dev orchestration
├── justfile             # just task runner
└── golangci.yml         # golangci-lint config
```

### Development Process

**Security changes** (auth, encryption, cookie handling, SSRF):
- Two-reviewer requirement
- Manual test against the eval-time attack tree
- No regression in `apilink.go` SSRF checks

**Manifest format changes** (NVMS YAML schema):
- RFC in `docs/CHANGELOG.md`
- Migration guide in `docs/migration/`
- Version bump on the schema

**Breaking API changes** (route shapes, response bodies):
- Spec delta in `SPEC.md`
- `fr-coverage.yml` updated
- Deprecation notice in `CHANGELOG.md` one release ahead

---

## 7. Charter Compliance Checklist

### For Deploy Features

- [ ] Manifest validates against the schema
- [ ] Credentials are encrypted at rest
- [ ] SSRF-safe HTTP client is used for all user-supplied URLs
- [ ] Auth cookies have `httpOnly`, `Secure`, `SameSite`
- [ ] `nvms` auth middleware is enabled
- [ ] OTel spans cover the deploy path

### For LLM Features

- [ ] Provider is pluggable
- [ ] Provider config is in the manifest, not the dashboard
- [ ] LLM key is encrypted at rest
- [ ] LLM responses are validated before being stored

### For Breaking Changes

- [ ] `SPEC.md` is updated
- [ ] `FUNCTIONAL_REQUIREMENTS.md` FR IDs are stable
- [ ] Migration guide exists
- [ ] Version is bumped (SemVer)

---

## 8. Decision Authority Levels

### Level 1: Maintainer Authority

**Scope:** Bug fixes, documentation, dependency bumps within Dependabot auto-merge rules  
**Process:** Maintainer approval, CI green

### Level 2: Core Team Authority

**Scope:** New routes, new LLM providers, new CI workflows  
**Process:** PR review by 1 core team member + CI green

### Level 3: Technical Steering Authority

**Scope:** Manifest schema changes, auth model changes, breaking API changes  
**Process:** Written proposal in `docs/proposals/`, 2 core team approvals

### Level 4: Executive Authority

**Scope:** Mission changes, charter rewrites, tenet changes  
**Process:** Charter diff, Org governance review

---

*This charter governs BytePort, the self-hosted IaC + portfolio platform.*  
*Last Updated: 2026-06-12*  
*Next Review: 2026-09-12 (quarterly)*
