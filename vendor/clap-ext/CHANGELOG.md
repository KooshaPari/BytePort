# Changelog

All notable changes to `clap-ext` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] — 2026-06-12

### Added

- **Initial release** of `clap-ext` — shared Rust CLI extension library
  for the Phenotype org.
- `common_args` module:
  - `ConfigArg` — `-c, --config <path>` with `PHENOTYPE_CONFIG` env var
  - `Verbosity` — `-v, --verbose` (count) + `--quiet` (mutually exclusive),
    with `to_filter()` mapping to `tracing_subscriber::filter::LevelFilter`
  - `OutputFormat` — `human` / `json` / `yaml` clap `ValueEnum`
- `common_subcommands` module:
  - `InitCmd` — `init <path> [-f] [-t <template>]`
  - `ValidateCmd` — `validate <path> [--strict]`
  - `VersionCmd` — `version` (zero-arg)
  - `CommonCommands` — bundled enum of all 3
  - `add_common_subcommands(cmd)` — imperative builder helper
  - `CommonSubcommand` trait alias
- `error` module:
  - `CliError` — unified error enum (Io, Config, Parse, Validation,
    NotFound, PermissionDenied, Network, Other) with `From<...>` for
    `io::Error`, `anyhow::Error`, `&str`, `String`
  - `CliResult<T>` — `Result<T, CliError>` type alias
  - `exit_with(err)` — print to stderr and exit 1
- `logging` module:
  - `setup_tracing(filter)` — `tracing-subscriber` setup with RUST_LOG honor
  - `setup_tracing_from_count(verbose_count, quiet)` — convenience wrapper
- `prelude` module — one-line import for the common case
- `clap-ext-macros` crate:
  - `#[clap_ext_common_subcommands]` — proc-macro attribute marker
    (passthrough for now; future versions will inject variants)
- `examples/basic` — runnable demo CLI
- 13 tests across 4 test files:
  - `tests/common_args.rs` (4 tests)
  - `tests/common_subcommands.rs` (3 tests)
  - `tests/error.rs` (4 tests)
  - `tests/integration.rs` (2 tests, exercises the basic example)
- GitHub Actions CI workflow (`.github/workflows/ci.yml`)
- README, LICENSE-MIT, LICENSE-APACHE, .gitignore, CHANGELOG

### Notes

- This is the v0.1.0 cut for the V9-T3-5 rollout wave. Adoption PRs
  in 5 sample repos are filed under
  `feat/clap-ext-adopt-2026-06-11` branches.
- The `clap_ext_common_subcommands` proc macro is currently a
  passthrough + doc-comment tag. Future versions (0.2.0+) will
  inject the 3 common variants automatically.
- MSRV is **1.82** (matches the org-wide `clippy.toml` MSRV policy).
