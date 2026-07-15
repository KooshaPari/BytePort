# BytePort Makefile
# Canonical build / test / development recipes for the BytePort monorepo.
# Usage:  make <target>  (e.g.  make help,  make dev,  make test)
#
# This Makefile wraps:
#   - Go backend (./backend)
#   - Rust workspace (./crates, ./ffi)
#   - Tauri frontend (./frontend/desktop)
#   - Next.js web (./frontend/web)
#   - Fuzz harnesses (./fuzz)
#
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

# ──────────────────────────────────────────────────────────────────────────────
# Configuration
# ──────────────────────────────────────────────────────────────────────────────
BACKEND_DIR    ?= backend
WORKSPACE_DIR  ?= .
FRONTEND_DIR   ?= frontend
FUZZ_DIR       ?= fuzz
BIN_DIR        ?= bin
REGISTRY       ?= ghcr.io/byteport
IMAGE_TAG      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

GO              ?= go
CARGO           ?= cargo
PNPM            ?= pnpm
DOCKER          ?= docker

GO_LDFLAGS      := -s -w -X main.version=$(IMAGE_TAG) -X main.commit=$(IMAGE_TAG)

# ──────────────────────────────────────────────────────────────────────────────
# Help / Info
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: version
version: ## Print version info
	@echo "BytePort $(IMAGE_TAG)"
	@echo "Go:      $$($(GO) version)"
	@echo "Cargo:   $$($(CARGO) --version)"
	@echo "Docker:  $$($(DOCKER) --version 2>/dev/null || echo 'not installed')"

# ──────────────────────────────────────────────────────────────────────────────
# Environment Setup
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: install
install: ## Install all toolchains (go, cargo, pnpm, rustup)
	@echo "→ Installing Go toolchain deps"
	cd $(BACKEND_DIR) && $(GO) mod download
	@echo "→ Installing Rust crates"
	$(CARGO) fetch --manifest-path $(WORKSPACE_DIR)/Cargo.toml
	@echo "→ Installing pnpm deps"
	cd $(FRONTEND_DIR)/web && $(PNPM) install --frozen-lockfile

.PHONY: tools
tools: ## Install development tools (clippy, mutants, sqlx, etc.)
	$(CARGO) install --locked --bin cargo-mutants --version ^0.5 || true
	$(CARGO) install --locked --bin cargo-clippy --version ^0.1 || true
	$(CARGO) install --locked --bin cargo-fuzz || true

# ──────────────────────────────────────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: build
build: build-backend build-cli ## Build all production artifacts
	@echo "✓ Build complete → $(BIN_DIR)/"

.PHONY: build-backend
build-backend: ## Build Go backend binary
	@mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(GO_LDFLAGS)" -o ../$(BIN_DIR)/byteport-api ./cmd/api
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(GO_LDFLAGS)" -o ../$(BIN_DIR)/byteport-mcp ./cmd/mcp-server
	@echo "✓ Backend built"

.PHONY: build-cli
build-cli: ## Build Rust CLI
	@mkdir -p $(BIN_DIR)
	$(CARGO) build --release --manifest-path $(WORKSPACE_DIR)/Cargo.toml -p byteport-cli
	cp target/release/byteport $(BIN_DIR)/byteport
	@echo "✓ CLI built"

.PHONY: build-all
build-all: build-backend build-cli build-frontend ## Build all artifacts (backend, cli, frontend)
	@echo "✓ All artifacts built"

.PHONY: build-frontend
build-frontend: ## Build Tauri desktop app
	cd $(FRONTEND_DIR)/desktop && $(PNPM) tauri build

# ──────────────────────────────────────────────────────────────────────────────
# Test
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: test
test: test-backend test-workspace ## Run all test suites

.PHONY: test-backend
test-backend: ## Run Go backend tests
	cd $(BACKEND_DIR) && $(GO) test -race -timeout 5m ./...

.PHONY: test-workspace
test-workspace: ## Run Rust workspace tests
	$(CARGO) test --workspace --quiet

.PHONY: test-property
test-property: ## Run property-based tests with verbose output
	cd $(BACKEND_DIR) && $(GO) test -v -run "Properties|Properties$" ./...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	cd $(BACKEND_DIR) && $(GO) test -tags=e2e -timeout 10m ./...

.PHONY: test-fuzz
test-fuzz: ## Run fuzz harnesses for 60 seconds each
	cargo fuzz run --manifest-path $(FUZZ_DIR)/Cargo.toml fuzz_dag_parse -- -max_total_time=60
	cargo fuzz run --manifest-path $(FUZZ_DIR)/Cargo.toml fuzz_transport_upload -- -max_total_time=60

.PHONY: coverage
coverage: ## Generate code coverage report
	cd $(BACKEND_DIR) && $(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report → coverage.html"

# ──────────────────────────────────────────────────────────────────────────────
# Lint / Format / Quality
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: lint
lint: lint-go lint-rust lint-frontend ## Run all linters

.PHONY: lint-go
lint-go: ## Lint Go code (vet + gofmt check)
	cd $(BACKEND_DIR) && $(GO) vet ./...
	cd $(BACKEND_DIR) && gofmt -l .

.PHONY: lint-rust
lint-rust: ## Lint Rust workspace with clippy
	$(CARGO) clippy --workspace -- -D warnings

.PHONY: lint-frontend
lint-frontend: ## Lint frontend code
	cd $(FRONTEND_DIR) && $(PNPM) lint

.PHONY: format
format: ## Format all code
	cd $(BACKEND_DIR) && gofmt -w .
	$(CARGO) fmt --all
	cd $(FRONTEND_DIR) && $(PNPM) format

.PHONY: mutation
mutation: ## Run cargo-mutants mutation testing
	$(CARGO) mutants --workspace --no-timescale -j 4

# ──────────────────────────────────────────────────────────────────────────────
# Run
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: dev
dev: ## Run backend + frontend in development mode
	@echo "Starting backend on :8080 and web on :3000..."
	@trap 'kill 0' EXIT; \
		(cd $(BACKEND_DIR) && $(GO) run ./cmd/api) & \
		(cd $(FRONTEND_DIR)/web && $(PNPM) dev) & \
		wait

.PHONY: dev-backend
dev-backend: ## Run only backend
	cd $(BACKEND_DIR) && $(GO) run ./cmd/api

.PHONY: dev-frontend
dev-frontend: ## Run only web frontend
	cd $(FRONTEND_DIR)/web && $(PNPM) dev

.PHONY: dev-desktop
dev-desktop: ## Run desktop app in dev mode
	cd $(FRONTEND_DIR)/desktop && $(PNPM) tauri dev

# ──────────────────────────────────────────────────────────────────────────────
# Docker
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: docker-build
docker-build: ## Build Docker images
	$(DOCKER) build -t $(REGISTRY)/backend:$(IMAGE_TAG) -f $(BACKEND_DIR)/Dockerfile --target backend .
	$(DOCKER) build -t $(REGISTRY)/cli:$(IMAGE_TAG) -f Dockerfile .

.PHONY: docker-compose-up
docker-compose-up: ## Start docker-compose stack
	$(DOCKER) compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop docker-compose stack
	$(DOCKER) compose down -v

# ──────────────────────────────────────────────────────────────────────────────
# Database
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: db-migrate
db-migrate: ## Run database migrations
	cd $(BACKEND_DIR) && $(GO) run ./cmd/migrate up

.PHONY: db-rollback
db-rollback: ## Rollback last database migration
	cd $(BACKEND_DIR) && $(GO) run ./cmd/migrate down 1

.PHONY: db-seed
db-seed: ## Seed database with development data
	cd $(BACKEND_DIR) && $(GO) run ./cmd/seed

# ──────────────────────────────────────────────────────────────────────────────
# Release
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: tag
tag: ## Create git tag for release (e.g. make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION required (e.g. make tag VERSION=v1.0.0)"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "✓ Tagged $(VERSION)"

.PHONY: release-changelog
release-changelog: ## Generate CHANGELOG.md via git-cliff
	git cliff --tag $$(git describe --tags --abbrev=0) > CHANGELOG.md

# ──────────────────────────────────────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)
	$(CARGO) clean
	cd $(BACKEND_DIR) && $(GO) clean -cache
	cd $(BACKEND_DIR) && rm -f coverage.out coverage.html
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name "node_modules" -prune -exec rm -rf {} + 2>/dev/null || true

.PHONY: clean-all
clean-all: clean ## Remove all build artifacts AND dependencies
	$(CARGO) clean --manifest-path $(WORKSPACE_DIR)/Cargo.toml
	cd $(BACKEND_DIR) && $(GO) clean -modcache
