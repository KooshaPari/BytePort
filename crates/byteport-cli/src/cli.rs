//! Top-level clap definitions for the `byteport` CLI.
//!
//! Subcommand enum variants live here; per-command handlers live in
//! [`crate::commands`].

use std::path::PathBuf;

use clap::{Parser, Subcommand, ValueEnum};

use crate::output::OutputFormat;

/// Top-level CLI. Use [`Cli::parse`] for normal command-line parsing or
/// [`Cli::try_parse_from`] for tests.
#[derive(Debug, Parser)]
#[command(
    name = "byteport",
    version,
    about = "BytePort — self-hosted IaC + portfolio platform CLI",
    long_about = "BytePort turns an `odin.nvms` manifest into a deployed, portfolio-worthy project. \
                  Use this CLI to scaffold projects, trigger deploys, run DAGs, and manage your \
                  self-hosted BytePort installation.\n\n\
                  Run `byteport doctor` after install to verify connectivity, auth, and config.",
    author,
    propagate_version = true,
    infer_subcommands = true,
    disable_help_subcommand = false
)]
pub struct Cli {
    /// Path to the config file (default: `$BYTEPORT_CONFIG` or `$XDG_CONFIG_HOME/byteport/config.toml`).
    #[arg(long, global = true, env = "BYTEPORT_CONFIG")]
    pub config: Option<PathBuf>,

    /// Log level: trace|debug|info|warn|error.
    #[arg(long, global = true, default_value = "info", env = "RUST_LOG")]
    pub log_level: String,

    /// Disable colored output (also honors `NO_COLOR`).
    #[arg(long, global = true)]
    pub no_color: bool,

    /// Output machine-readable JSON for every command.
    #[arg(long, global = true)]
    pub json: bool,

    /// Skip the startup update check (also honored via `BYTEPORT_NO_UPDATE_CHECK`).
    #[arg(long, global = true)]
    pub no_update_check: bool,

    /// Output format override (default: human table; `--json` sets this to `json`).
    #[arg(long, global = true, value_enum, default_value_t = FormatArg::Auto)]
    pub format: FormatArg,

    #[command(subcommand)]
    pub command: Commands,
}

/// How to format subcommand output.
#[derive(Debug, Clone, Copy, ValueEnum, Default)]
pub enum FormatArg {
    /// Auto: human table on TTY, JSON when piped or when `--json` is set.
    #[default]
    Auto,
    /// Force human-readable table output.
    Table,
    /// Force JSON output.
    Json,
}

impl FormatArg {
    /// Resolve to the effective output format.
    pub fn resolve(self, json_flag: bool, no_color: bool) -> OutputFormat {
        match self {
            FormatArg::Auto => {
                if json_flag {
                    OutputFormat::Json
                } else {
                    OutputFormat::Table
                }
            }
            FormatArg::Table => OutputFormat::Table,
            FormatArg::Json => OutputFormat::Json,
        }
        .also(|_| {
            // Side-effect: surface the color decision to the trace layer.
            if no_color {
                colored::control::set_override(colored::control::ShouldColorize::Never);
            }
        })
    }
}

trait Also: Sized {
    fn also(self, f: impl FnOnce(&Self)) -> Self {
        f(&self);
        self
    }
}
impl<T> Also for T {}

/// All top-level subcommands.
#[derive(Debug, Subcommand)]
pub enum Commands {
    /// Show version + build info.
    #[command(visible_alias = "v")]
    Version,

    /// Scaffold a new `byteport.yaml` in the current directory.
    Init {
        /// Overwrite an existing `byteport.yaml` if present.
        #[arg(long)]
        force: bool,
        /// Create a sample project alongside the YAML.
        #[arg(long, default_value_t = true)]
        sample_project: bool,
    },

    /// Authenticate with a BytePort server via device-code flow.
    Login {
        /// Override the server URL.
        #[arg(long)]
        server: Option<String>,
    },

    /// Clear stored credentials.
    Logout,

    /// Print the currently authenticated user (or "Not authenticated").
    Whoami,

    /// Manage projects.
    #[command(visible_alias = "p", subcommand)]
    Project(ProjectCommand),

    /// Manage deploys.
    #[command(visible_alias = "d", subcommand)]
    Deploy(DeployCommand),

    /// DAG operations (run / upload / validate).
    #[command(subcommand)]
    Dag(DagCommand),

    /// Manage CLI configuration.
    #[command(subcommand)]
    Config(ConfigCommand),

    /// Generate shell completion scripts.
    Completions {
        /// Target shell.
        #[arg(value_enum)]
        shell: ShellArg,
    },

    /// Generate man pages (one per subcommand) to stdout.
    Man,

    /// Check for a newer BytePort release.
    Update,

    /// Diagnostics: config, auth, connectivity, version.
    Doctor,
}

/// Available shells for completion generation.
#[derive(Debug, Clone, Copy, ValueEnum)]
pub enum ShellArg {
    Bash,
    Zsh,
    Fish,
    PowerShell,
    Elvish,
}

/// Project subcommands.
#[derive(Debug, Subcommand)]
pub enum ProjectCommand {
    /// List all projects visible to the current user.
    List,
    /// Create a new project.
    Create {
        /// Project name (slug, unique within the active profile).
        name: String,
    },
    /// Delete a project by id.
    Delete {
        /// Project id.
        id: String,
    },
    /// Show a single project's details.
    Show {
        /// Project id.
        id: String,
    },
    /// Open the project's dashboard URL in your default browser.
    Open {
        /// Project id.
        id: String,
    },
}

/// Deploy subcommands.
#[derive(Debug, Subcommand)]
pub enum DeployCommand {
    /// List deploys (optionally filtered by project).
    List {
        /// Restrict to a single project.
        #[arg(long)]
        project: Option<String>,
    },
    /// Trigger a new deploy for the given project.
    Trigger {
        /// Project name or id.
        project: String,
    },
    /// Cancel an in-flight deploy.
    Cancel {
        /// Deploy id.
        id: String,
    },
    /// Tail deploy logs (or print all logs if the deploy is finished).
    Logs {
        /// Deploy id.
        id: String,
        /// Follow the log stream.
        #[arg(short, long)]
        follow: bool,
    },
    /// Print the status of a deploy.
    Status {
        /// Deploy id.
        id: String,
    },
}

/// DAG subcommands.
#[derive(Debug, Subcommand)]
pub enum DagCommand {
    /// Parse a YAML DAG definition, compute a schedule, and print the plan.
    Run {
        /// Path to the YAML file containing the DAG definition.
        yaml: PathBuf,
        /// Optional DAG name filter (only execute nodes matching a pattern).
        #[arg(short, long)]
        name: Option<String>,
        /// Enable verbose output including the serialized schedule.
        #[arg(short, long)]
        verbose: bool,
    },
    /// Dispatch a `Transport::CreateUpload` call (wrapped in an OTel span).
    Upload {
        /// Object key (path within the bucket).
        key: String,
        /// MIME content type.
        #[arg(long, default_value = "application/octet-stream")]
        content_type: String,
        /// Content length in bytes.
        #[arg(long, default_value_t = 0)]
        content_length: u64,
        /// S3-compatible storage endpoint.
        #[arg(long, default_value = "https://storage.example.test")]
        endpoint: String,
        /// S3 bucket name.
        #[arg(long, default_value = "byteport-uploads")]
        bucket: String,
        /// Optional key prefix.
        #[arg(long)]
        prefix: Option<String>,
    },
    /// Parse a YAML DAG file and print its structure (no execution).
    Validate {
        /// Path to the YAML file containing the DAG definition.
        yaml: PathBuf,
    },
}

/// CLI config subcommands.
#[derive(Debug, Subcommand)]
pub enum ConfigCommand {
    /// Get a single config value by dotted key (e.g. `profiles.default.server`).
    Get {
        /// Dotted key to look up.
        key: String,
    },
    /// Set a config value.
    Set {
        /// Dotted key to set.
        key: String,
        /// Value to assign.
        value: String,
    },
    /// List all config values.
    List,
    /// Print the resolved config file path.
    Path,
    /// Open the config file in `$EDITOR` (falls back to `vi`).
    Edit,
}

#[cfg(test)]
mod tests {
    use super::*;
    use clap::CommandFactory;

    #[test]
    fn cli_definition_is_valid() {
        Cli::command().debug_assert();
    }

    #[test]
    fn parse_help() {
        let cli = Cli::try_parse_from(["byteport", "--help"]).unwrap_err();
        assert_eq!(cli.kind(), clap::error::ErrorKind::DisplayHelp);
    }

    #[test]
    fn parse_version_flag() {
        let cli = Cli::try_parse_from(["byteport", "--version"]).unwrap_err();
        assert_eq!(cli.kind(), clap::error::ErrorKind::DisplayVersion);
    }

    #[test]
    fn parse_subcommand_dag_run() {
        let cli = Cli::try_parse_from(["byteport", "dag", "run", "x.yaml"]).unwrap();
        match cli.command {
            Commands::Dag(DagCommand::Run { yaml, .. }) => {
                assert_eq!(yaml, PathBuf::from("x.yaml"));
            }
            other => panic!("unexpected: {other:?}"),
        }
    }

    #[test]
    fn parse_global_json_flag() {
        let cli = Cli::try_parse_from(["byteport", "--json", "version"]).unwrap();
        assert!(cli.json);
    }
}