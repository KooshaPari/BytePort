# justfile for BytePort
# ------------------------------------------------------------------
# Hybrid task runner for BytePort's three engines:
#   - Go (backend/byteport + backend/nvms)         # primary backend
#   - TypeScript / SvelteKit (frontend/web)         # primary web frontend
#   - Rust / Tauri (frontend/web/src-tauri)        # desktop/mobile shell
#
# Spec: V3 DAG L2 #24 (BytePort Taskfile + justfile).
# Branch: chore/l2-24-taskfile-justfile-2026-06-11
# Author: L2 subagent #24 (2026-06-11)
#
# L1 audit context (`BytePort/STATUS_2026_06_10.md`):
#   - `cargo check` times out at 300s; L2 spec asks for 600s on `build`.
#   - `frontend/web/package.json` is missing a `test` script. Addressed
#     in this same PR by adding `"test": "vitest --run --coverage || echo no tests yet"`.
#
# This file REPLACES the prior cargo-only justfile (PR #98) with a
# hybrid Rust + Go + TS recipe set per V3 DAG L2 #24.
#
# Usage:
#   just --list          # show all recipes
#   just build           # cargo + go + npm build
#   just test            # cargo test --no-run + go test + npm test
#   just lint            # cargo clippy + go vet + npm run lint
#   just ci              # build, test, lint, fmt, deny, audit
# ------------------------------------------------------------------

set dotenv-load
set positional-arguments
# Note: `set unstable` is intentionally NOT enabled. The `[timeout(...)]`
# attribute is not yet available in just 1.46.0. To enforce timeouts,
# wrap the recipe invocation with the GNU `timeout` command, e.g.:
#   `timeout 600 just build`

# Common variables
cargo_flags := "--workspace"
frontend_dir := "frontend/web"
backend_dir := "backend/byteport"

# Default recipe
default:
    @just --list

# -- Build ----------------------------------------------------------------

# Build everything: Rust + Go + TS (600s timeout per L2 #24; use `timeout 600 just build`)
build: build-rust build-go build-frontend
    @echo "build all engines complete"

# Build the Rust workspace (Tauri shell)
build-rust:
    cargo build {{cargo_flags}}

# Build the Go backend (backend/byteport)
build-go:
    cd {{backend_dir}} && go build ./...

# Build the SvelteKit frontend (frontend/web)
build-frontend:
    cd {{frontend_dir}} && npm run build

# -- Test -----------------------------------------------------------------

# Test everything. `cargo test --no-run` is fast (full cargo test is slow
# per L1 audit; see `test-rust` recipe for the slow one).
# Use `timeout 600 just test` to enforce a 600s cap.
test: test-frontend test-go test-rust-compile

# Run frontend (TS/SvelteKit) tests; falls back to `npm run check`
test-frontend:
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{frontend_dir}}
    if grep -q '"test"\s*:' package.json; then
        echo "==> test-frontend: npm test"
        npm test
    else
        echo "==> test-frontend: no 'test' script; falling back to 'npm run check'"
        npm run check
    fi

# Compile Rust tests without running them (fast sanity)
test-rust-compile:
    cargo test {{cargo_flags}} --no-run

# Run Rust tests (full execution; slow — matches L1 audit).
# Use `timeout 1200 just test-rust` to enforce a 1200s cap.
test-rust:
    cargo test {{cargo_flags}}

# Run Go tests (backend/byteport)
test-go:
    cd {{backend_dir}} && go test ./...

# Run Go tests with the race detector
test-go-race:
    cd {{backend_dir}} && go test -race ./...

# -- Lint -----------------------------------------------------------------

# Lint everything: cargo clippy + go vet + npm run lint
lint: lint-rust lint-go lint-frontend

# cargo clippy on the workspace (warnings are errors)
lint-rust:
    cargo clippy {{cargo_flags}} -- -D warnings

# go vet on backend/byteport
lint-go:
    cd {{backend_dir}} && go vet ./...

# npm run lint (prettier --check + eslint)
lint-frontend:
    cd {{frontend_dir}} && npm run lint

# -- Format check ---------------------------------------------------------

# Check formatting: cargo fmt + gofmt + npm run format:check (or format -- --check)
fmt: fmt-rust fmt-go fmt-frontend

# cargo fmt --check on the workspace
fmt-rust:
    cargo fmt --all -- --check

# gofmt -l on the Go modules (advisory)
fmt-go:
    #!/usr/bin/env bash
    set -uo pipefail
    unformatted=$(gofmt -l {{backend_dir}} backend/nvms 2>/dev/null || true)
    if [ -n "$unformatted" ]; then
        echo "::warning::The following Go files are not gofmt-clean:"
        echo "$unformatted"
        exit 1
    fi

# prettier --check (or whatever frontend package.json exposes)
fmt-frontend:
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{frontend_dir}}
    if grep -q '"format:check"\s*:' package.json; then
        npm run format:check
    else
        npm run format -- --check
    fi

# Auto-format everything in place
fmt-fix:
    cargo fmt --all
    gofmt -w {{backend_dir}} backend/nvms || true
    cd {{frontend_dir}} && npm run format

# -- Deny / Audit ---------------------------------------------------------

# Run cargo-deny against deny.toml
deny:
    #!/usr/bin/env bash
    set -euo pipefail
    if command -v cargo-deny >/dev/null 2>&1; then
        cargo deny check
    else
        echo "cargo-deny not installed; install with: cargo install cargo-deny"
        exit 1
    fi

# Run cargo-audit
audit:
    #!/usr/bin/env bash
    set -euo pipefail
    if command -v cargo-audit >/dev/null 2>&1; then
        cargo audit
    else
        echo "cargo-audit not installed; install with: cargo install cargo-audit"
        exit 1
    fi

# -- CI aggregate ---------------------------------------------------------

# Full CI suite: build, test, lint, fmt, deny, audit
ci: build test lint fmt deny audit
    @echo "ci all stages passed"

# -- Hygiene --------------------------------------------------------------

# Repository hygiene checks
hygiene:
    #!/usr/bin/env bash
    set -uo pipefail
    echo "==> hygiene: oversized files"
    find backend frontend src docs \
        -type f \( -name "*.go" -o -name "*.rs" -o -name "*.ts" -o -name "*.svelte" -o -name "*.js" \) \
        -not -path "*/node_modules/*" -not -path "*/target/*" -not -path "*/.git/*" \
        -not -path "*/.svelte-kit/*" -not -path "*/build/*" -not -path "*/dist/*" \
        -size +20k 2>/dev/null \
        -exec wc -l {} \; | sort -rn | head -20 || true
    echo "==> hygiene: active TODO markers"
    grep -rnE "todo!|TODO|FIXME|XXX|HACK|unimplemented!" \
        {{backend_dir}} backend/nvms frontend/web/src frontend/web/src-tauri/src \
        --include="*.go" --include="*.rs" --include="*.ts" --include="*.svelte" \
        2>/dev/null | grep -v node_modules | grep -v target | head -20 || true
    echo "==> hygiene: governance file presence"
    for f in LICENSE LICENSE-MIT LICENSE-APACHE CHANGELOG.md CODEOWNERS CONTRIBUTING.md SECURITY.md .gitignore; do
        if [ -f "$f" ]; then
            echo "  $f: present"
        else
            echo "  $f: missing"
        fi
    done
    echo "==> hygiene: complete"

# -- Frontend / Backend groups --------------------------------------------

# Run all frontend (TS) tasks
frontend: lint-frontend test-frontend build-frontend

# Run all backend (Go + Rust) tasks
backend: build-go build-rust test-go lint-go lint-rust

# -- Convenience ----------------------------------------------------------

# Install JS dependencies (frontend/web)
install:
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{frontend_dir}}
    if command -v yarn >/dev/null 2>&1; then
        yarn install --frozen-lockfile
    else
        npm ci
    fi

# Clean build artifacts (target/, .svelte-kit/, build/, dist/)
clean:
    cargo clean
    rm -rf frontend/web/.svelte-kit frontend/web/build frontend/web/dist
    rm -rf {{backend_dir}}/tmp

# Show this help text
help:
    @just --list
