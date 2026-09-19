# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Frontend: the lint gate was failing on `prettier --check` (22 unformatted files,
  including `package.json` and the Tauri config, which used spaces while
  `.prettierrc` sets `useTabs`) and on `eslint` (`@eslint/js` is imported by
  `eslint.config.js` but was never declared, so a clean `npm ci` could not
  resolve it). Also fixes `no-useless-assignment` in `src/lib/api.ts`.
- Benchmarks: the baseline comparison step lost its closing `fi`, which made the
  shell script fail with "syntax error: unexpected end of file".
- CI: repair the Tier-0/Tier-1 gates. Several workflows failed before running at
  all because of unresolvable action pins; others broke on tool version drift
  (golangci-lint v2 CLI flags, govulncheck requiring Go 1.26, Node 20 vs
  chromatic 18), on markdown templates living under `.github/workflows`, and on
  a missing Tauri system-dependency step in the Rust jobs.
- Go: record HTTP request/error metrics in middleware and count deploys and
  sandbox stops, so the Prometheus counters are no longer dead code.
- Rust: patch RUSTSEC-2026-0194/0195 (quick-xml) and RUSTSEC-2026-0009 (time),
  and drop dependencies cargo-machete reports as unused.
- Scorecard: the ENV_VARS pillar now accepts `.env.example` templates.

### Changed

- Rename `justfile` to `Justfile` so the JUSTFILE pillar passes on
  case-sensitive filesystems.

### Added

- B34: Tier-1 enforcement on PR (cargo audit security scan, CycloneDX SBOM
  generation/validation, LICENSE presence check, CHANGELOG update check)
- BytePort→NanoVMS contract alignment (5 endpoints, Bearer auth, SandboxConfig)
- Contract integration tests (6 tests: deploy, stop, config shape, URL, token)
- E2E lifecycle tests (full lifecycle, multi-sandbox, auth propagation)
- Load testing benchmarks (deploy, list, stop throughput, 50-concurrent stress)
- /health endpoint (uptime, status, service name)
- /metrics endpoint (Prometheus text format: uptime, requests, deploys, memory)
- INSTALL.md (Docker Compose, source, Tauri, CLI install paths)
- DEPLOYMENT.md (production setup, nginx reverse proxy, env vars, troubleshooting)
- GOVERNANCE.md (decision process, release process, versioning)
- API_REFERENCE.md (complete API documentation with NanoVMS integration)
- Integration evidence package (cross-repo contract verification)

### Changed

- Fixed 6 API contract mismatches with NanoVMS (URL prefix, endpoint names, port, auth, request/response shapes)
- CI: every `uses:` pin is now a verified commit SHA. Four workflows were pinned to
  non-existent SHAs (`actions/setup-go@0a12ed9e…`, `golangci/golangci-lint-action@aa6339a8…`,
  `actions/setup-node@1a4442ca…`, `actions/upload-artifact@65c4c4a1…`,
  `ossf/scorecard-action@99c09fe9…`), which aborted the jobs before they ran.
- CI: `golangci-lint-action` moved to v9.3.0 (action v6 cannot run golangci-lint v2), the
  removed `--disable` flags were replaced by `linters.disable` in `golangci.yml`, and the
  config was migrated to the golangci-lint v2 schema.
- CI: Tier-0 Rust jobs install the GTK/WebKit development packages the Tauri build needs.
- CI: frontend installs use `--legacy-peer-deps` (as `release-tauri.yml` already did) and
  `@chromatic-com/storybook` is pinned to the Storybook 8 line (`^3.2.7`); `package-lock.json`
  and `yarn.lock` regenerated.
- CI: `cargo fmt --all` applied; unused dependencies removed from `crates/byteport-dag`,
  `crates/byteport-otel` and `frontend/web/src-tauri`.
- CI: govulncheck pinned to `v1.5.0` (newer releases require Go ≥ 1.26 while the module targets
  1.25), `cargo cyclonedx --format json` (the 0.5.x flag), LICENSE placeholder check fixed,
  lefthook installed onto `PATH`, pre-commit `language_version` unpinned from python3.11.
- CI: `ratchet.yml`, `verify-attestation.yml` and `release-attest.yml` were markdown templates
  that had been dropped into `.github/workflows/` (they produced workflow runs with zero jobs);
  they now live in `docs/ci-templates/`. The orphan duplicate `.github/workflows/ci.yaml` was
  removed.
- Renamed `justfile` to `Justfile` (the `JUSTFILE` pillar check is case-sensitive).

### Security

- `Cargo.lock`: `quick-xml` 0.38.4 → 0.41.0 (RUSTSEC-2026-0194, RUSTSEC-2026-0195) and
  `time` 0.3.41 → 0.3.55 (RUSTSEC-2026-0009), pulling `plist` 1.7.4 → 1.10.0. `cargo audit`
  no longer reports vulnerabilities.
- `rustsec/audit-check` jobs now hold `checks: write` so they can publish their check run.

### Fixed

- Align deploy endpoint to use NanoVMS SandboxConfig instead of models.Project
- Align terminate endpoint to use NanoVMS stop API
- Update monitor poll paths to /v1/sandboxes prefix

## [0.1.0] - 2026-06-14

### Added

- Initial release with version tracking.

[Unreleased]: https://github.com/KooshaPari/BytePort/compare/0.1.0...HEAD
[0.1.0]: https://github.com/KooshaPari/BytePort/releases/tag/0.1.0
