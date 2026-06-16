# Security Policy

## Threat model

The current per-component STRIDE threat model lives at
[`docs/security/threat-model.md`](docs/security/threat-model.md). It enumerates
the production attack surface (SvelteKit frontend, Tauri shell, Go backend,
`nvms` orchestrator, LLM providers, AWS, CI/CD pipeline) and the mitigations
in place today. Review cadence: on every minor release, on any new external
dependency, and quarterly at minimum.

## Reporting Vulnerabilities

Please report security vulnerabilities via GitHub Security Advisories:

- Open a [private security advisory](../../security/advisories/new)
- For sensitive issues, contact the repository owner directly

## Supported Versions

Latest `main` branch. Older versions are not supported.

## Disclosure Policy

We follow coordinated disclosure with reporters. Once an issue is patched, an advisory will be published.

## Cargo-deny

Rust projects in this org enforce a zero-advisory floor via `cargo-deny.yml` workflow (Monday cron + on-demand).

## CodeQL

Static analysis runs Tuesday weekly via `codeql-rust.yml` workflow.
