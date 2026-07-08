//! Error types for the BytePort CLI.
//!
//! The CLI uses a layered error strategy:
//!
//! 1. **Internal errors** ([`CliError`]) — `thiserror`-typed failures that
//!    are safe to display to the user and that map cleanly to exit codes.
//! 2. **Library errors** propagated from upstream crates (DAG, transport,
//!    telemetry) — wrapped in [`CliError::Internal`] with file/line context
//!    via `anyhow`.
//!
//! Use [`Result<T>`] throughout command handlers.

use std::process::ExitCode;

use thiserror::Error;

/// Top-level CLI error type. All command handlers return [`Result<T>`].
#[derive(Debug, Error)]
pub enum CliError {
    /// Configuration file could not be loaded or parsed.
    #[error("config error: {0}")]
    Config(String),

    /// I/O failure (file read/write, network, etc.).
    #[error("io error: {0}")]
    Io(#[from] std::io::Error),

    /// YAML / TOML / JSON deserialization failure.
    #[error("parse error: {0}")]
    Parse(String),

    /// DAG validation failed (cycle, missing node, etc.).
    #[error("dag error: {0}")]
    Dag(String),

    /// Network / API failure (login, deploy trigger, update check, …).
    #[error("network error: {0}")]
    Network(String),

    /// Authentication / authorization failure.
    #[error("auth error: {0}")]
    Auth(String),

    /// The requested operation is not yet implemented (placeholder stub).
    #[error("not implemented: {0}")]
    NotImplemented(String),

    /// User-supplied input is invalid.
    #[error("invalid input: {0}")]
    InvalidInput(String),

    /// A clap parser failure propagated upward.
    #[error("argument error: {0}")]
    Args(String),

    /// Anything else — preserves the underlying cause via `anyhow`.
    #[error("internal error: {0}")]
    Internal(#[from] anyhow::Error),
}

/// Convenience alias used throughout the CLI.
pub type Result<T> = std::result::Result<T, CliError>;

impl From<serde_json::Error> for CliError {
    fn from(e: serde_json::Error) -> Self {
        CliError::Parse(format!("json: {e}"))
    }
}

impl From<serde_yaml::Error> for CliError {
    fn from(e: serde_yaml::Error) -> Self {
        CliError::Parse(format!("yaml: {e}"))
    }
}

impl From<toml::de::Error> for CliError {
    fn from(e: toml::de::Error) -> Self {
        CliError::Parse(format!("toml (de): {e}"))
    }
}

impl From<toml::ser::Error> for CliError {
    fn from(e: toml::ser::Error) -> Self {
        CliError::Parse(format!("toml (ser): {e}"))
    }
}

/// Map a [`CliError`] to a stable exit code.
///
/// Convention follows the project's CLI exit-code table:
/// - `0` — success
/// - `1` — generic failure
/// - `2` — invalid usage (bad args, parse errors)
/// - `3` — authentication failure
/// - `4` — network error
/// - `5` — config error
/// - `6` — DAG validation error
/// - `7` — not-implemented stub (still surfaces a non-zero exit for CI)
pub fn exit_code(err: &CliError) -> u8 {
    match err {
        CliError::Config(_) => 5,
        CliError::Io(_) => 1,
        CliError::Parse(_) => 2,
        CliError::Dag(_) => 6,
        CliError::Network(_) => 4,
        CliError::Auth(_) => 3,
        CliError::NotImplemented(_) => 7,
        CliError::InvalidInput(_) => 2,
        CliError::Args(_) => 2,
        CliError::Internal(_) => 1,
    }
}

/// Run a [`Result<T>`] and convert it into an [`ExitCode`].
///
/// Use this in [`main`](crate::main) to keep the binary entry-point tiny.
pub fn run<T>(result: Result<T>) -> ExitCode {
    match result {
        Ok(_) => ExitCode::from(0),
        Err(e) => {
            eprintln!("error: {e:#}");
            ExitCode::from(exit_code(&e))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn exit_code_mapping_is_stable() {
        assert_eq!(exit_code(&CliError::NotImplemented("x".into())), 7);
        assert_eq!(exit_code(&CliError::Dag("x".into())), 6);
        assert_eq!(exit_code(&CliError::Config("x".into())), 5);
        assert_eq!(exit_code(&CliError::Network("x".into())), 4);
        assert_eq!(exit_code(&CliError::Auth("x".into())), 3);
        assert_eq!(exit_code(&CliError::Parse("x".into())), 2);
    }

    #[test]
    fn run_returns_zero_on_ok() {
        assert_eq!(run(Ok(())).to_string(), "0");
    }
}