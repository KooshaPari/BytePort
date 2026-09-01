# BytePort

Declarative deployment infrastructure from GitHub repository to running service.

[![AI slop inside](https://sladge.net/badge.svg)](https://sladge.net) [![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/KooshaPari/BytePort/total)](https://github.com/KooshaPari/BytePort/releases)

BytePort is a self-hosted deployment platform that uses a repository manifest to coordinate authenticated project ingestion, build/deployment configuration, AWS provisioning, endpoint registration, and observability behind a single developer workflow.

> **Development model**
>
> BytePort uses agent-assisted implementation and maintenance. Product direction,
> requirements, architecture, integration strategy, verification standards, and
> release governance are human-directed. Repository artifacts may include
> machine-generated code and documentation and are validated through the project's
> CI, security, and verification workflows.

## Status

- **Current repository:** Go/Gin/GORM/SQLite backend services, SvelteKit web frontend, Tauri shell, GitHub/OAuth and AWS integration surfaces, and CI/security workflows.
- **Planned:** manifest-driven delivery completion and isolated microVM execution. These are not represented as a currently available runtime.
## Zero-Config Start (Docker Compose)

> **Fastest path — no tools to install beyond Docker.**
> Addresses scorecard gaps **S01** (Install-to-Use Time) and **U08** (Zero-Config Start).

```sh
git clone https://github.com/KooshaPari/BytePort.git
cd BytePort
cp .env.example .env        # optional — defaults work for local dev
docker compose up --build
```

This builds two containers:

| Service | Port | What it runs |
|---|---|---|
| `backend` | **8080** | Go API (Gin + SQLite) |
| `web` | **3000** | SvelteKit frontend |

Open <http://localhost:3000> and follow the on-screen signup flow.
No tmux, no `spin`, no `air`, no manual Go/Node installs required.

> **Stop:** `Ctrl-C` or `docker compose down`.

---

## Quickstart (60 seconds)

### Prerequisites

| Tool | Version | Why |
|---|---|---|
| `go` | 1.25+ | Backend, NVMS |
| `node` | 20+ | SvelteKit |
| `npm` | 10+ | SvelteKit deps |
| `tmux` | any | `./start dev` orchestration |
| `spin` | 1.20+ | `nvms` runtime (https://developer.fermyon.com/spin) |
| `air` | latest | Go hot-reload (`go install github.com/cosmtrek/air@latest`) |
| `git` | any | obvious |

### One command to start everything

```sh
./start dev
```

This opens a tmux session with three panes:

- SvelteKit dev server (port 5173)
- Go backend with `air` hot-reload (port 8081)
- (Spin is started manually for `nvms` — see `backend/nvms/README.md`)

### Repository workflow

The GitHub linking, manifest, deployment, and instance-management surfaces are represented in the repository. The end-to-end manifest-driven deployment workflow remains under active implementation; see [PLAN.md](PLAN.md) for the current delivery status.

### Example `odin.nvms` manifest (planned workflow)


```yaml
NAME: my-app
DESCRIPTION: A task management web application

SERVICES:
  - NAME: "main"           # Required — public-facing service, exposed at "/"
    PATH: "./frontend"
    PORT: 8080
    RUNTIME: "nodejs"
    BUILD: ["npm install", "npm run build"]
    ENV:
      API_URL: "http://localhost:8081"

  - NAME: "backend"
    PATH: "./backend"
    PORT: 8081
    RUNTIME: "go"
    BUILD: ["go build -o server ./cmd/server"]
    ENV:
      DATABASE_URL: "postgres://localhost/myapp"

INFRASTRUCTURE:
  compute: ec2             # or ecs, lambda
  region: us-east-1
  instance_type: t3.micro

PORTFOLIO:
  generate_page: true
  description_source: llm  # or readme, manual
```

### Production

```sh
./start prod
```

Builds the SvelteKit frontend, runs `npm start`, then `go run main.go`.

---

## Project layout

```
BytePort/
├── backend/
│   ├── byteport/        # Core API: Go 1.25, Gin, GORM, SQLite
│   │   ├── main.go      # Entry: OTel init, auth init, Gin server
│   │   ├── lib/         # Auth, crypto, git, apilink (SSRF-safe)
│   │   ├── models/      # GORM data models
│   │   └── routes/      # Gin HTTP handlers
│   └── nvms/            # NVMS runtime: Go 1.25, Spin wasm, port 3000
│       ├── main.go      # Spin HTTP entry + router
│       ├── projectManager/  # deploy/terminate logic
│       ├── lib/         # LLM providers
│       ├── Provisioner/ # MicroVM lifecycle
│       └── Builder/     # Image building
├── frontend/
│   └── web/             # SvelteKit 2 admin UI
│       ├── src/         # Routes + components
│       └── src-tauri/   # Tauri 2 desktop shell
├── docs/                # Long-form documentation (auto-generated + hand-written)
├── .github/workflows/   # CI: go-ci, npm-ci, tauri-ci, nvms-ci, codeql, etc.
├── start                # tmux dev orchestration
├── start.bat            # Windows parity (Phase 9)
├── justfile             # just task runner
├── golangci.yml         # golangci-lint config
├── deny.toml            # cargo-deny config
├── AGENTS.md            # Forge agent instructions
├── CLAUDE.md            # Claude-specific orientation
├── CHARTER.md           # Mission, tenets, scope
├── PLAN.md              # v1.0 roadmap (Phase 0–11)
├── SPEC.md              # Canonical technical spec
├── SPECS_INDEX.md       # Auto-generated audit index
├── STATUS.md            # Current health + known gaps
├── PRD.md               # Epics + stories
├── FUNCTIONAL_REQUIREMENTS.md   # 20 FRs, traced to PRD epics
├── ARCHITECTURE.md      # Component boundaries
├── stub-inventory.md    # Open TODOs and stubs
└── worklog.md           # Development log
```

---

## Governance

- **`STATUS.md`** — current health, known gaps, what v1.0 means
- **`CHARTER.md`** — mission, tenets, scope, success criteria, authority levels
- **`PLAN.md`** — v1.0 roadmap (10 phases + governance pass)
- **`SPEC.md`** — canonical technical spec (stack, data models, API, security)
- **`PRD.md`** — epics + stories
- **`FUNCTIONAL_REQUIREMENTS.md`** — 20 FRs, grouped by capability, traced to PRD
- **`ARCHITECTURE.md`** — component boundaries
- **`SPECS_INDEX.md`** — auto-generated audit index
- **`stub-inventory.md`** — open TODOs and stubs
- **`worklog.md`** — development log

---

## Roadmap (v1.0)

See `PLAN.md` for the full 173-task DAG. Headline phases:

- **Phase 0** (PR #1) — Governance reset ← *we are here*
- **Phase 1** (PR #2) — Security & reliability floor (4 critical bugs)
- **Phase 2** (PR #3) — Manifest engine (NVMS YAML)
- **Phase 3** (PR #4) — Backend hardening (slog, OTLP, rate limit, healthz, shutdown)
- **Phase 4** (PR #5) — NVMS service completion (auth on, LLM providers, /metrics)
- **Phase 5** (PR #6) — SvelteKit frontend (11 routes, zod, superforms, runes, i18n, a11y)
- **Phase 6** (PR #7) — Tauri 2 desktop shell (signing, notarize, CSP, deep-link, updater)
- **Phase 7** (PR #8) — CI/CD (15+ workflows, dependabot, release-drafter, FR coverage)
- **Phase 8** (PR #9) — Dev orchestration & onboarding (parameterized `./start`, Dockerfiles)
- **Phase 9** (PR #10) — Long-form documentation
- **Phase 10** — Verification matrix (23 gates, fan-in at v1.0.0)

---

## API at a glance

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `POST` | `/signup` | public | Create account |
| `POST` | `/login` | public | Authenticate, set PASETO cookie |
| `GET` | `/authenticate` | cookie | Validate token, return user object |
| `GET` | `/link` | cookie | Initiate GitHub OAuth |
| `POST` | `/link` | cookie | Save AWS + LLM + Portfolio credentials |
| `GET` | `/instances` | cookie | List user's instances |
| `GET` | `/projects` | cookie | List user's projects |
| `GET` | `/api/github/callback` | OAuth | GitHub OAuth callback |
| `GET` | `/api/github/repositories` | cookie | List user's GitHub repos |
| `POST` | `/deploy` | cookie | Trigger deployment |
| `POST` | `/terminate` | cookie | Terminate an instance |
| `GET` | `/user/:id/creds` | cookie | Get decrypted credentials |
| `PUT` | `/user/:id/creds` | cookie | Update profile (name, email, password) |

Full request/response shapes in `SPEC.md` §4.

---

## Security model (summary)

- **At rest** — all credentials (AWS, GitHub, LLM, Portfolio) AES-256-CFB encrypted with auto-generated master key
- **Passwords** — Argon2id (memory=64MiB, iterations=3, parallelism=2, salt=16B, key=32B)
- **Session** — PASETO v2 tokens in httpOnly cookies
- **GitHub tokens** — auto-refresh every 7h45m via background goroutine
- **SSRF protection** — `lib/apilink.go` rejects loopback / private / link-local / multicast; allowlist via env
- **AWS validation** — STS session + `s3.ListBuckets` smoke test
- **OpenAI validation** — single `GET /v1/models` call
- **OTel traces** — every protected handler wrapped in otelgin middleware

Full security model in `SPEC.md` §5 and `CHARTER.md` §2 (Tenets 7, 8).

---

## Quality gates (CI)

| Gate | Command | Required |
|---|---|---|
| `go vet ./backend/...` | 0 warnings | yes |
| `go build ./backend/...` | 0 errors | yes |
| `go test ./backend/...` | all pass | yes |
| `golangci-lint run` | 0 errors | yes |
| `cargo test` (src-tauri) | all pass | yes |
| `cargo clippy -- -D warnings` (src-tauri) | 0 errors | yes |
| `npm run check` (frontend) | 0 errors | yes |
| `osv-scanner --recursive .` | clean | yes |
| `trufflehog filesystem .` | 0 secrets | yes |
| `codeql analyze` | 0 alerts | yes |

Full verification matrix in `PLAN.md` Phase 10.

---

## Example projects

- [Fixit-Go](https://github.com/kooshapari/fixit-go) — Go + SvelteKit todo list, ready for BytePort deploy
- [Chatta](https://github.com/kooshapari/chatta) — Real-time chat, multi-service manifest
- [Slickport](https://github.com/kooshapari/slickport) — Portfolio integration example

---

## API Documentation (Swagger UI)

Interactive API docs are published via **GitHub Pages**. The
[swagger-ui.yml](.github/workflows/swagger-ui.yml) workflow automatically
builds and deploys the Swagger UI whenever `docs/openapi.yaml` changes on
`main`.

**To enable it:**

1. Go to **Settings > Pages** in the GitHub repo.
2. Under **Build and deployment > Source**, select **GitHub Actions**.
3. Push any change to `docs/openapi.yaml` (or trigger the workflow manually via
   **Actions > Deploy Swagger UI > Run workflow**).

The live docs will be available at
`https://kooshapari.github.io/BytePort/`.

---

## Related work

- **Phenotype-org governance** — `phenotype-org-governance/` defines the org-wide
  rules this repo follows (FR IDs, ADRs, cargo-deny baseline, branch protection).
- **Authvault** (formerly `authkit`) — Rust auth/secrets crate used by sibling repos.
- **phenotype-auth-ts** — TypeScript auth SDK.

---

## Contributing

- Read `PLAN.md` Phase 0–1 to understand current state and what's queued next.
- Read `AGENTS.md` for worktree + integration rules.
- Open a draft PR early; the org quality gate (`fr-coverage.yml`) requires
  PR-to-FR traceability before merge.

---

## License

MIT. See `LICENSE`.