# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Fixed

- Align deploy endpoint to use NanoVMS SandboxConfig instead of models.Project
- Align terminate endpoint to use NanoVMS stop API
- Update monitor poll paths to /v1/sandboxes prefix

## [0.1.0] - 2026-06-14

### Added

- Initial release with version tracking.

[Unreleased]: https://github.com/KooshaPari/BytePort/compare/0.1.0...HEAD
[0.1.0]: https://github.com/KooshaPari/BytePort/releases/tag/0.1.0
