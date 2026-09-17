//! `byteport` CLI entry point — Tauri desktop shell plus headless subcommands.
//!
//! Invoked with no arguments, the binary launches the Tauri desktop shell
//! (see [`app_lib::run`]). The headless subcommands (`transport check`,
//! `transport status`, `deploy`, `version`) emit machine-readable JSON on
//! stdout so CI and scripts can assert on transport configuration or review a
//! deployment plan without a display server. Errors are reported on stderr
//! with a non-zero exit code.
//!
//! This file owns the CLI surface ([`Cli`], [`Command`], [`TransportCommand`],
//! [`CliError`]) and the stdout/exit-code plumbing. The subcommand
//! implementations sit next to it in [`deploy`] (planning) and [`headless`]
//! (build identity and transport probes).

mod deploy;
mod headless;

use std::io::{self, Write};
use std::process::ExitCode;

use clap::{Parser, Subcommand};

#[derive(Debug, Parser)]
#[command(
    name = "byteport",
    version,
    about = "BytePort core engine CLI",
    long_about = None
)]
struct Cli {
    #[command(subcommand)]
    command: Option<Command>,
}

#[derive(Debug, Subcommand)]
enum Command {
    /// Launch the Tauri desktop shell
    Desktop,

    /// Print the deployment plan for a target as JSON (plans only, never executes)
    Deploy {
        /// Deployment target name, for example `docker` or `local`
        #[arg(long, value_name = "NAME")]
        target: String,
    },

    /// Upload transport operations
    Transport {
        #[command(subcommand)]
        command: TransportCommand,
    },

    /// Print name, crate version, and git revision as JSON
    Version,
}

#[derive(Debug, Subcommand)]
enum TransportCommand {
    /// Verify the S3 upload transport can mint an upload instruction
    Check,

    /// Print the resolved upload endpoint and bucket as JSON
    Status,
}

/// Errors surfaced by the headless subcommands.
#[derive(Debug, thiserror::Error)]
enum CliError {
    /// The resolved transport configuration cannot produce a usable upload.
    #[error("transport configuration error: {0}")]
    Config(String),

    /// The underlying transport rejected the request.
    #[error("upload transport error: {0}")]
    Transport(#[from] byteport_transport::UploadTransportError),

    /// JSON rendering failed.
    #[error("json serialization error: {0}")]
    Json(#[from] serde_json::Error),

    /// Writing to stdout failed.
    #[error("i/o error: {0}")]
    Io(#[from] std::io::Error),

    /// The deployment plan could not be assembled.
    #[error("deployment plan error: {0}")]
    Plan(#[from] deploy::PlanError),
}

fn main() -> ExitCode {
    let cli = Cli::parse();

    match cli.command.unwrap_or(Command::Desktop) {
        Command::Desktop => {
            app_lib::run();
            ExitCode::SUCCESS
        }
        Command::Version => with_stdout(headless::cmd_version),
        Command::Deploy { target } => with_stdout(|out| deploy::cmd_deploy(&target, out)),
        Command::Transport { command } => {
            let lookup = headless::default_env_lookup();
            let config = headless::resolve_upload_config(&lookup);
            match command {
                TransportCommand::Check => with_stdout(|out| headless::cmd_transport_check(&config, out)),
                TransportCommand::Status => with_stdout(|out| headless::cmd_transport_status(&config, out)),
            }
        }
    }
}

/// Run a JSON-emitting subcommand against stdout, mapping errors to a
/// non-zero exit code with a stderr diagnostic.
fn with_stdout<F>(run: F) -> ExitCode
where
    F: FnOnce(&mut dyn Write) -> Result<(), CliError>,
{
    let stdout = io::stdout();
    let mut out = stdout.lock();

    match run(&mut out) {
        Ok(()) => match out.flush() {
            Ok(()) => ExitCode::SUCCESS,
            Err(err) => {
                eprintln!("byteport: failed to flush stdout: {err}");
                ExitCode::FAILURE
            }
        },
        Err(err) => {
            drop(out);
            eprintln!("byteport: {err}");
            ExitCode::FAILURE
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // ---- CLI surface parsing -------------------------------------------

    #[test]
    fn parses_version_command() {
        let cli = Cli::try_parse_from(["byteport", "version"]).expect("version should parse");
        assert!(matches!(cli.command, Some(Command::Version)));
    }

    #[test]
    fn parses_transport_status_command() {
        let cli = Cli::try_parse_from(["byteport", "transport", "status"]).expect("transport status should parse");
        assert!(matches!(
            cli.command,
            Some(Command::Transport {
                command: TransportCommand::Status
            })
        ));
    }

    #[test]
    fn parses_transport_check_command() {
        let cli = Cli::try_parse_from(["byteport", "transport", "check"]).expect("transport check should parse");
        assert!(matches!(
            cli.command,
            Some(Command::Transport {
                command: TransportCommand::Check
            })
        ));
    }

    #[test]
    fn parses_desktop_command() {
        let cli = Cli::try_parse_from(["byteport", "desktop"]).expect("desktop should parse");
        assert!(matches!(cli.command, Some(Command::Desktop)));
    }

    #[test]
    fn no_subcommand_defers_to_desktop() {
        let cli = Cli::try_parse_from(["byteport"]).expect("bare invocation should parse");
        assert!(cli.command.is_none());
    }

    #[test]
    fn transport_requires_an_action() {
        assert!(Cli::try_parse_from(["byteport", "transport"]).is_err());
    }

    #[test]
    fn rejects_unknown_transport_action() {
        assert!(Cli::try_parse_from(["byteport", "transport", "frobnicate"]).is_err());
    }

    // ---- deploy ---------------------------------------------------------

    fn render_deploy(target: &str) -> serde_json::Value {
        let mut buffer = Vec::new();
        deploy::cmd_deploy(target, &mut buffer).expect("deploy should render");
        serde_json::from_slice(&buffer).expect("deploy output should be valid json")
    }

    #[test]
    fn parses_deploy_command_with_target_flag() {
        let cli = Cli::try_parse_from(["byteport", "deploy", "--target", "docker"]).expect("deploy should parse");

        assert!(
            matches!(&cli.command, Some(Command::Deploy { target }) if target == "docker"),
            "expected a deploy command for docker, got {:?}",
            cli.command
        );
    }

    #[test]
    fn deploy_requires_a_target() {
        assert!(Cli::try_parse_from(["byteport", "deploy"]).is_err());
    }

    #[test]
    fn deploy_emits_a_well_formed_plan_with_stages_and_warnings() {
        let value = render_deploy("docker");

        assert_eq!(value["target"], serde_json::json!("docker"));

        let stages = value["stages"].as_array().expect("plan should carry a stages array");
        assert!(!stages.is_empty(), "plan should have at least one stage");
        for stage in stages {
            assert!(stage["name"].is_string());
            assert!(stage["command"].is_string());
            assert!(stage["estimated_seconds"].is_u64());
        }

        let warnings = value["warnings"]
            .as_array()
            .expect("plan should carry a warnings array");
        assert!(!warnings.is_empty(), "plan should warn that it is plan-only");
        assert!(warnings.iter().all(serde_json::Value::is_string));
    }
}
