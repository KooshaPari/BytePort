# BytePort

**BytePort is the only open-source desktop deployment application in the Phenotype ecosystem.** A multi-language cross-platform solution that turns your local machine into a deployment target for containerized workloads.

[![CI](https://github.com/KooshaPari/BytePort/actions/workflows/ci.yml/badge.svg)](https://github.com/KooshaPari/BytePort/actions)
[![License: MIT OR Apache-2.0](https://img.shields.io/badge/License-MIT%2FApache--2.0-blue.svg)](LICENSE)
[![Releases](https://img.shields.io/github/v/release/KooshaPari/BytePort?style=flat-square&label=release)](https://github.com/KooshaPari/BytePort/releases)

---

## What BytePort Is

A **Tauri 2** desktop application that lets you deploy, manage, and monitor containerized workloads from a native desktop interface. Unlike cloud-only solutions, BytePort runs locally, giving you direct control over deployment environments.

**Key distinction**: BytePort is the **only open-source desktop deployment app** in the Phenotype ecosystem. Closest competitors are Server Compass (proprietary, $29) and Coolify (web-based, requires separate server).

---

## Architecture

BytePort follows a **three-tier architecture**:

```
┌─────────────────────────────────────────────────────────────┐
│                      Tauri 2 Desktop Shell                    │
│  ┌─────────────────┐    ┌─────────────────────────────────┐   │
│  │ SvelteKit UI    │◄──►│ Tauri IPC Bridge                │   │
│  │ (Svelte 5)      │    │ Rust core + Go backend protocol │   │
│  └─────────────────┘    └─────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go / Gin Backend                         │
│  • HTTP gRPC API for Tauri shell                             │
│  • Container orchestration (Podman/Docker CLI)               │
│  • Persistence layer (SQLite)                                │
│  • Health check & readiness probes                           │
└─────────────────────────────────────────────────────────────┘
```

### Rust Workspace (`Cargo.toml`)

| Crate | Responsibility |
|-------|--------------|
| `crates/byteport-transport` | Wire transport primitives (serde models, errors) |
| `crates/byteport-cli` | CLI bindings for transport |
| `crates/byteport-dag` | DAG foundation: executor interfaces, topological sort |
| `crates/byteport-otel` | OpenTelemetry instrumentation |
| `frontend/web/src-tauri` | Desktop shell binary |

### Shared Crates (from [PhenoInfra](https://github.com/KooshaPari/PhenoInfra))

- `phenotype-crypto` – Cross-platform crypto primitives
- `phenotype-observability` – Structured logging, metrics
- `phenotype-health` – Health endpoint utilities

### Go Modules

| Module | Location |
|--------|----------|
| `byteport` (main) | `backend/byteport/` |
| `bytebridge` (legacy) | `backend/bytebridge/ByteBridge/` |

### Frontend

- **UI Framework**: SvelteKit v5 with Svelte 5 runes
- **Build Tools**: Vite, TypeScript, Tailwind CSS, ESLint/Prettier
- **App Shell**: Tauri 2 (Rust core + webview frontend)

---

## Installation

### Development Environment

```bash
# Prerequisites
- Rust stable (rust-toolchain.toml pins "stable" with rustfmt, clippy, rust-src)
- Go 1.25.0
- Node.js 22+ (bun or npm)
- Tauri 2 development dependencies (see docs/CONTRIBUTING.md)

# Clone
git clone https://github.com/KooshaPari/BytePort.git
cd BytePort

# Build the Rust backend/CLI crates.
# NOTE: the desktop GUI crate (frontend/web/src-tauri) is excluded on purpose.
# It must go through the Tauri CLI so frontend assets are embedded. A plain
# `cargo build` of that crate produces a binary that loads the dev-server URL
# (http://localhost:5173) and shows a white window when no dev server runs.
# See "Build Desktop App" below.
cargo build --workspace --exclude app

# Install frontend deps
cd frontend/web
npm install  # or bun install
```

### Build Desktop App

```bash
# Build Tauri app for current platform (also runs the frontend build)
cargo tauri build

# Outputs (verified on macOS/aarch64):
#   target/release/bundle/macos/Byteport.app
#   target/release/bundle/dmg/Byteport_<version>_<arch>.dmg
#
# `make build-app` runs the same thing. Do NOT use a bare
# `cargo build --release -p app`: it skips the frontend build and the
# `custom-protocol` feature, producing a white-window binary.
```

### Run Development Server

```bash
# Terminal 1: Build Rust core
cargo tauri dev

# Terminal 2: Start frontend dev
cd frontend/web
npm run dev
```

The app will open automatically in Tauri dev mode.

---

## Project Structure

```
BytePort/
├── backend/             # Go backend - TWO modules, see backend/README.md
│   ├── byteport/       # CANONICAL Go module: what the app talks to, port 8081
│   └── bytebridge/     # Legacy bridge components
├── crates/             # Rust workspace crates
│   ├── byteport-cli/   # CLI bindings
│   ├── byteport-dag/   # DAG executor/scheduler
│   ├── byteport-otel/  # OpenTelemetry instrumentation
│   ├── byteport-transport/  # Wire types
│   └── integration/    # Shared Phenotype crate wiring
├── frontend/           # SvelteKit frontend
│   └── web/            # Tauri-compatible SvelteKit app
├── src-tauri/          # Tauri desktop shell (inside frontend/web)
├── docs/               # Architecture, contributions, guides
├── Cargo.toml          # Rust workspace (resolver 3)
├── rust-toolchain.toml # Rust stable + components
└── go.mod              # Go 1.25.0 module
```

> **Backend has two Go modules.** `backend/byteport/` is canonical - it is what
> the desktop app talks to and what the Dockerfile builds. `backend/` is a
> separate, unreachable `/api/v1` module (WorkOS AuthKit rewrite) that nothing
> calls. See [`backend/README.md`](backend/README.md).


---

## Usage

After building/running:

1. **Launch the application** – Native desktop window opens
2. **Connect to local daemon** – Auto-connects to local Go backend
3. **Deploy containers** – Use UI or CLI tools to manage workloads
4. **Monitor** – View container health, logs, metrics in real-time

See `docs/` for detailed guides on features, CLI usage, and API reference.

---

## Documentation

| Document | Purpose |
|----------|---------|
| `docs/ARCHITECTURE.md` | Full architecture reference |
| `docs/CONTRIBUTING.md` | Development setup, testing, commits |
| `docs/ops/` | Operations guides |
| `docs/reviews/` | Design reviews, ADRs |
| `docs/research/` | Experiment protocols, findings |

---

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for:
- Development environment setup
- Code style (Rustfmt, Clippy, ESLint/Prettier)
- Testing (cargo test, cargo-bench, vitest)
- Commit conventions (ledger trailers: `tx-agent:`, `tx-validated:`)
- PR review process

---

## License

Dual-licensed under **Apache-2.0** OR **MIT**. See [LICENSE](LICENSE) for details.

---

## Related

- **PhenoInfra** – Shared phenotype crates ([KooshaPari/PhenoInfra](https://github.com/KooshaPari/PhenoInfra))
- **PhenoShared** – Shared workspace docs, CI, crates ([KooshaPari/PhenoShared](https://github.com/KooshaPari/PhenoShared))
<!-- AI-DD-META:END -->
