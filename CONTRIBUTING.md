# Contributing to BytePort

Thank you for your interest in contributing to BytePort — the cross-platform
high-throughput IPC + caching layer for the Phenotype ecosystem. We welcome
bug reports, documentation improvements, tests, refactors, and feature
contributions from everyone.

This document explains how to set up your development environment, run the
test suite, propose changes, and get them merged safely.

---

## 1. Code of Conduct

By participating, you agree to abide by the [Phenotype Code of Conduct](CODE_OF_CONDUCT.md)
(if present) and the GitHub Community Guidelines. Be respectful, assume good
faith, and prefer written communication that can be quoted later.

## 2. Project Overview

BytePort is a transport-agnostic, type-safe, zero-copy IPC substrate. The
core implementation is split between:

- A **Rust** core that owns the protocol state machine, codec, scheduler,
  and memory-mapped buffer pool.
- A **Go** sidecar/control plane that handles topology discovery, health
  probes, and admin RPCs.
- A **TypeScript / Svelte** desktop shell (`frontend/`) for operators
  and a **Tauri** (`src-tauri/`) shell for the macOS/Windows/Linux apps.

The repository root and the three sub-trees each have their own
`Cargo.toml`, `go.mod`, and `package.json`; the `Taskfile.yml` (and
`justfile`) at the root orchestrate them.

## 3. Development Environment

### 3.1 Required Toolchains

| Tool          | Version  | Why                                |
|---------------|----------|------------------------------------|
| Rust          | `stable` | Core IPC engine, FFI bindings      |
| `cargo`       | ≥ 1.78   | Build, test, fmt, clippy           |
| `rustfmt`     | stable   | Formatting                         |
| `clippy`      | stable   | Lints (fail-on-warn in CI)         |
| `cargo-deny`  | ≥ 0.14   | License + advisory gating          |
| `cargo-audit` | ≥ 0.20   | Vulnerability scan (pre-commit)    |
| Go            | ≥ 1.22   | Sidecar / control plane            |
| `golangci-lint` | ≥ 1.55 | Aggregated Go lints                |
| Node.js       | ≥ 20 LTS | Frontend build                     |
| `pnpm`        | ≥ 9      | Frontend package manager           |
| Tauri CLI     | ≥ 1.6    | Desktop shell                      |
| Task          | ≥ 3      | Cross-language task runner         |

### 3.2 Clone + Bootstrap

```bash
git clone https://github.com/KooshaPari/byteport.git
cd byteport
task bootstrap        # or: ./scripts/bootstrap.sh
```

`task bootstrap` will:

1. Install git hooks (`lefthook` or `.githooks/`).
2. Install `pre-commit` and `pre-push` checks.
3. Run `cargo fetch`, `go mod download`, `pnpm install`.
4. Run a smoke build of all three language sub-projects.

### 3.3 Editor Setup

- **VS Code**: open the workspace file at `.vscode/byteport.code-workspace`.
  Recommended extensions: `rust-analyzer`, `golang.go`, `svelte.svelte-vscode`,
  `tauri-apps.tauri-vscode`, `editorconfig.editorconfig`.
- **Neovim / Helix / Zed**: zero-config LSPs work out of the box; the
  `rust-analyzer` and `gopls` configuration is committed in `.config/`.
- **JetBrains**: RustRover for the core, GoLand for the sidecar, WebStorm
  for the frontend.

## 4. Building

```bash
# Everything (Rust + Go + Tauri + frontend)
task build

# Just the core
cargo build --workspace --all-targets

# Just the sidecar
(cd backend && go build ./...)

# Just the desktop app
(cd frontend && pnpm tauri build)
```

A successful full build produces:

- `target/release/libbyteport.rlib` (Rust core)
- `backend/bin/byteport-sidecar` (Go binary)
- `frontend/src-tauri/target/release/bundle/{dmg,msi,AppImage}/...` (Tauri bundles)

## 5. Testing

BytePort has a tiered test pyramid:

| Tier          | Command                                  | Owner       | Wall-clock |
|---------------|------------------------------------------|-------------|------------|
| Unit          | `cargo test --workspace`                 | Core team   | < 2 min    |
| Unit (Go)     | `go test ./...`                          | Sidecar     | < 1 min    |
| Unit (Web)    | `pnpm --filter ./frontend test`          | UI team     | < 1 min    |
| Integration   | `task test:integration`                  | Core team   | < 5 min    |
| Fuzz          | `cargo +nightly fuzz run codec -- -max_total_time=300` | Security | 5 min     |
| Soak (local)  | `task test:soak -- --duration 60`        | SRE         | 60 min     |

CI runs unit + integration + fuzz on every PR. The full soak matrix runs
nightly and on release tags.

## 6. Coding Standards

- **Rust**: `cargo fmt --all`, `cargo clippy --workspace --all-targets -- -D warnings`.
  No `unwrap()` in non-test code; use `anyhow::Context` or a typed error.
- **Go**: `gofmt -s`, `goimports -local github.com/KooshaPari/byteport`,
  `golangci-lint run`. Wrap errors with `%w`.
- **TypeScript**: `prettier --check .`, `eslint .`, `svelte-check`.
  Prefer discriminated unions over enums.
- **Svelte**: `<script lang="ts">` always; no untyped props on exported
  components.
- **Commits**: conventional commits — see §9.

## 7. Branching

- Default branch: `main`.
- Long-lived integration branches: `release/X.Y`.
- Feature/fix branches: `feat/<scope>-<short-desc>`, `fix/<scope>-<short-desc>`,
  `chore/<scope>-<short-desc>`. The `<scope>` matches the conventional
  commit scope (e.g. `feat/codec-varint`, `fix/sidecar-rpc`).
- Branch names are kebab-case, ≤ 60 chars.

## 8. Pull Request Process

1. **Open an issue first** for non-trivial changes. Bug fixes and
   documentation improvements may go straight to PR.
2. **Fork** the repo (or push to a feature branch if you have write
   access via the Phenotype org).
3. **Keep PRs focused**: < 400 lines diff where possible. Split larger
   refactors into a stack of dependent PRs.
4. **Fill the PR template** — it links to the design doc / spec / issue
   and the test plan.
5. **Pass CI**: lint, build, all tier-1 tests, license/advisory gating,
   CodeQL, and the rustsec / govulncheck scan.
6. **Request a review** from the CODEOWNERS — for BytePort, the default
   reviewer is `@KooshaPari`. Add a domain reviewer (e.g. security,
   frontend) for cross-cutting changes.
7. **Address review feedback** in additional commits; the maintainer
   will squash-merge once the conversation is resolved.
8. **After merge**, delete the source branch.

## 9. Commit Message Format (Conventional Commits)

BytePort uses [Conventional Commits 1.0.0](https://www.conventionalcommits.org/).

```
<type>(<scope>): <short summary>

<body — wrap at 72 cols; explain *what* and *why*>

<footer — e.g. "BREAKING CHANGE: ...", "Closes #123", "Refs: SPEC-42">
```

### Allowed types

| Type       | Semantics                                                    |
|------------|--------------------------------------------------------------|
| `feat`     | A new user-facing feature                                    |
| `fix`      | A bug fix                                                    |
| `docs`     | Documentation only                                           |
| `style`    | Whitespace/formatting, no code change                        |
| `refactor` | Code change that neither fixes a bug nor adds a feature      |
| `perf`     | Performance improvement                                      |
| `test`     | Add or correct tests                                         |
| `build`    | Build system, CI, or dependency change                       |
| `chore`    | Tooling, repo hygiene, governance (this PR)                  |
| `revert`   | Reverts a previous commit (include `Reverts: <sha>`)         |
| `security` | Security fix (also notify `security@phenotype.internal`)     |

### Scopes (non-exhaustive)

`core`, `codec`, `transport`, `sidecar`, `rpc`, `topology`, `frontend`,
`tauri`, `ci`, `docs`, `deps`, `governance`.

### Examples

```
feat(codec): add varint zigzag encoding for u32 fields

Reduces wire size for the 95-percentile small-int telemetry frame
by ~38% (from 8 bytes to 5 median) without losing the 32-bit
range needed for monotonic timestamps.

BREAKING CHANGE: the `FrameHeader` magic byte is bumped to 0xB7 to
signal v2 codec. Mixed-version nodes will be rejected by the
handshake.

Closes #842
Refs: SPEC-31 §4.2
```

```
fix(sidecar): drop admin RPC if grace window exceeded

The sidecar was forwarding admin RPCs to nodes that had already
left the mesh but whose heartbeat had not yet expired. We now
honour the configured `admin_rpc_grace_ms` window strictly.

Adds a regression test under `backend/services/admin/`.

Fixes #1289
```

## 10. Reviewer Expectations

- **First response** within 2 business days.
- Reviews cover: correctness, test coverage, security, performance,
  API stability, observability, and documentation.
- Maintainer privilege: squash-merge with the PR title as the squash
  subject and the PR body as the squash body. Override only when the
  history itself is meaningful.

## 11. Release Process

BytePort follows semver. Releases are cut from `main` via the
`release-please` GitHub App configured in `.github/release-please-config.json`.
The maintainer approves the release PR, which is auto-generated and
bumps versions, CHANGELOG, and tags.

## 12. Getting Help

- **Discord**: `#byteport` on the Phenotype Discord.
- **Discussions**: GitHub Discussions → *Q&A*.
- **Office hours**: Wednesdays 16:00 UTC, calendar link in the
  pinned issue.

Welcome aboard — we are glad you are here.
