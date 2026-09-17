//! `byteport` CLI entry point — Tauri desktop shell plus headless subcommands.
//!
//! Invoked with no arguments, the binary launches the Tauri desktop shell
//! (see [`app_lib::run`]). The headless subcommands (`transport check`,
//! `transport status`, `deploy`, `version`) emit machine-readable JSON on
//! stdout so CI and scripts can assert on transport configuration or review a
//! deployment plan without a display server. Errors are reported on stderr
//! with a non-zero exit code.
//!
//! `deploy` is a planner, not an executor: see [`deploy`].

mod deploy;

use std::io::{self, Write};
use std::process::ExitCode;

use byteport_transport::{S3UploadTransport, UploadRequest, UploadTransport};
use clap::{Parser, Subcommand};
use serde_json::json;

/// Environment variable holding the S3-compatible upload endpoint.
const ENV_UPLOAD_URL: &str = "BYTEPORT_UPLOAD_URL";

/// Environment variable holding the target upload bucket.
const ENV_UPLOAD_BUCKET: &str = "BYTEPORT_UPLOAD_BUCKET";

/// Dev default endpoint, mirroring the desktop wiring in `lib.rs`.
const DEFAULT_UPLOAD_URL: &str = "https://uploads.byteport.local";

/// Dev default bucket, mirroring the desktop wiring in `lib.rs`.
const DEFAULT_UPLOAD_BUCKET: &str = "byteport-uploads";

/// Key prefix applied by the desktop shell, reused here so `transport
/// check` exercises the same wiring the GUI uses.
const UPLOAD_KEY_PREFIX: &str = "desktop";

/// Synthetic object key used by `transport check` to prove the transport
/// can mint an upload instruction.
const PROBE_OBJECT_KEY: &str = "_healthcheck/transport-check";

/// Content type sent with the synthetic probe upload.
const PROBE_CONTENT_TYPE: &str = "application/json";

/// Content length sent with the synthetic probe upload.
const PROBE_CONTENT_LENGTH: u64 = 2;

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

/// Upload transport configuration resolved from the environment.
#[derive(Debug, Clone, PartialEq, Eq)]
struct UploadConfig {
    /// Effective endpoint URL.
    url: String,
    /// Effective bucket name.
    bucket: String,
    /// Whether `url` came from the environment rather than the default.
    url_from_env: bool,
    /// Whether `bucket` came from the environment rather than the default.
    bucket_from_env: bool,
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
        Command::Version => with_stdout(cmd_version),
        Command::Deploy { target } => with_stdout(|out| deploy::cmd_deploy(&target, out)),
        Command::Transport { command } => {
            let lookup = default_env_lookup();
            let config = resolve_upload_config(&lookup);
            match command {
                TransportCommand::Check => with_stdout(|out| cmd_transport_check(&config, out)),
                TransportCommand::Status => with_stdout(|out| cmd_transport_status(&config, out)),
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

/// Environment lookup used by the real binary; tests inject their own.
fn default_env_lookup() -> impl Fn(&str) -> Option<String> {
    |key| std::env::var(key).ok()
}

/// Resolve the effective upload configuration.
///
/// A variable that is unset *or* blank falls back to the dev default, so
/// `url_from_env`/`bucket_from_env` mean "the effective value came from
/// the environment".
fn resolve_upload_config(lookup: &dyn Fn(&str) -> Option<String>) -> UploadConfig {
    let env_url = lookup(ENV_UPLOAD_URL).filter(|value| !value.trim().is_empty());
    let env_bucket = lookup(ENV_UPLOAD_BUCKET).filter(|value| !value.trim().is_empty());

    UploadConfig {
        url_from_env: env_url.is_some(),
        bucket_from_env: env_bucket.is_some(),
        url: env_url.unwrap_or_else(|| DEFAULT_UPLOAD_URL.to_string()),
        bucket: env_bucket.unwrap_or_else(|| DEFAULT_UPLOAD_BUCKET.to_string()),
    }
}

/// `byteport version` — name, crate version, and git revision as JSON.
fn cmd_version(out: &mut dyn Write) -> Result<(), CliError> {
    let payload = json!({
        "name": "byteport",
        "version": env!("CARGO_PKG_VERSION"),
        "build_id": build_id(),
    });

    write_json(out, &payload)
}

/// `byteport transport status` — resolved endpoint/bucket plus provenance.
fn cmd_transport_status(config: &UploadConfig, out: &mut dyn Write) -> Result<(), CliError> {
    let payload = json!({
        "url": config.url,
        "bucket": config.bucket,
        "configured_via_env": {
            "url": config.url_from_env,
            "bucket": config.bucket_from_env,
        },
    });

    write_json(out, &payload)
}

/// `byteport transport check` — mint a probe upload instruction and
/// assert the transport resolved it as expected.
///
/// Fails (non-zero exit) when the endpoint or bucket is unusable, when
/// the transport rejects the probe, or when the returned instruction
/// disagrees with the configured endpoint/bucket/key prefix.
fn cmd_transport_check(config: &UploadConfig, out: &mut dyn Write) -> Result<(), CliError> {
    validate_endpoint(&config.url)?;

    if config.bucket.trim().is_empty() {
        return Err(CliError::Config("upload bucket is empty".to_string()));
    }

    let transport = S3UploadTransport::new(
        config.url.as_str(),
        config.bucket.as_str(),
        Some(UPLOAD_KEY_PREFIX),
    );

    let instruction = transport.create_upload(&UploadRequest {
        object_key: PROBE_OBJECT_KEY.to_string(),
        content_type: PROBE_CONTENT_TYPE.to_string(),
        content_length: PROBE_CONTENT_LENGTH,
    })?;

    let expected_key = format!(
        "{UPLOAD_KEY_PREFIX}/{}",
        PROBE_OBJECT_KEY.trim_start_matches('/')
    );
    let resolved_key = object_key_from_url(&instruction.url, &config.url, &config.bucket)
        .ok_or_else(|| {
            CliError::Config(format!(
                "transport returned an upload URL outside {}/{}/: {}",
                config.url.trim_end_matches('/'),
                config.bucket,
                instruction.url
            ))
        })?;

    if resolved_key != expected_key {
        return Err(CliError::Config(format!(
            "transport resolved object key {resolved_key:?}, expected {expected_key:?}"
        )));
    }

    if instruction.method != "PUT" {
        return Err(CliError::Config(format!(
            "transport returned method {:?}, expected \"PUT\"",
            instruction.method
        )));
    }

    // The transport exposes no server-side session id; the object key is
    // this upload's stable identity, so it is reported as `upload_id` and
    // tagged with `upload_id_kind` to keep the semantics unambiguous.
    let payload = json!({
        "ok": true,
        "upload_id": resolved_key,
        "upload_id_kind": "object_key",
        "endpoint": config.url,
        "bucket": config.bucket,
        "method": instruction.method,
        "target_url": instruction.url,
        "headers": instruction.headers,
    });

    write_json(out, &payload)
}

/// Serialize `value` as pretty JSON followed by a newline.
fn write_json(out: &mut dyn Write, value: &serde_json::Value) -> Result<(), CliError> {
    serde_json::to_writer_pretty(&mut *out, value)?;
    out.write_all(b"\n")?;
    Ok(())
}

/// Reject endpoints the S3 transport cannot build a URL from.
fn validate_endpoint(endpoint: &str) -> Result<(), CliError> {
    let trimmed = endpoint.trim();

    if trimmed.is_empty() {
        return Err(CliError::Config("upload endpoint is empty".to_string()));
    }

    if !(trimmed.starts_with("http://") || trimmed.starts_with("https://")) {
        return Err(CliError::Config(format!(
            "upload endpoint must be an http(s) URL, got {trimmed:?}"
        )));
    }

    Ok(())
}

/// Extract the object key from a transport-built upload URL, or `None`
/// when the URL does not sit under `endpoint/bucket/`.
fn object_key_from_url(url: &str, endpoint: &str, bucket: &str) -> Option<String> {
    let prefix = format!("{}/{}/", endpoint.trim_end_matches('/'), bucket);
    url.strip_prefix(&prefix).map(str::to_string)
}

/// Short git revision of the working tree, or `dev` when unavailable.
fn build_id() -> String {
    git_short_revision().unwrap_or_else(|| "dev".to_string())
}

/// Run `git rev-parse --short HEAD`, returning `None` on any failure so
/// the CLI still works from a release tarball without a `.git` directory.
fn git_short_revision() -> Option<String> {
    let output = std::process::Command::new("git")
        .arg("rev-parse")
        .arg("--short")
        .arg("HEAD")
        .output()
        .ok()?;

    if !output.status.success() {
        return None;
    }

    let revision = String::from_utf8(output.stdout).ok()?;
    let revision = revision.trim();

    if revision.is_empty() {
        None
    } else {
        Some(revision.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn env_with(url: Option<&str>, bucket: Option<&str>) -> impl Fn(&str) -> Option<String> {
        let url = url.map(str::to_string);
        let bucket = bucket.map(str::to_string);
        move |key: &str| match key {
            ENV_UPLOAD_URL => url.clone(),
            ENV_UPLOAD_BUCKET => bucket.clone(),
            _ => None,
        }
    }

    fn render_check(config: &UploadConfig) -> serde_json::Value {
        let mut buffer = Vec::new();
        cmd_transport_check(config, &mut buffer).expect("transport check should succeed");
        serde_json::from_slice(&buffer).expect("check output should be valid json")
    }

    // ---- CLI surface parsing -------------------------------------------

    #[test]
    fn parses_version_command() {
        let cli = Cli::try_parse_from(["byteport", "version"]).expect("version should parse");
        assert!(matches!(cli.command, Some(Command::Version)));
    }

    #[test]
    fn parses_transport_status_command() {
        let cli = Cli::try_parse_from(["byteport", "transport", "status"])
            .expect("transport status should parse");
        assert!(matches!(
            cli.command,
            Some(Command::Transport {
                command: TransportCommand::Status
            })
        ));
    }

    #[test]
    fn parses_transport_check_command() {
        let cli = Cli::try_parse_from(["byteport", "transport", "check"])
            .expect("transport check should parse");
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

    // ---- version -------------------------------------------------------

    #[test]
    fn version_output_is_json_with_name_version_and_build_id() {
        let mut buffer = Vec::new();
        cmd_version(&mut buffer).expect("version should render");

        let value: serde_json::Value =
            serde_json::from_slice(&buffer).expect("version output should be valid json");

        assert_eq!(value["name"], serde_json::json!("byteport"));
        assert_eq!(value["version"], serde_json::json!(env!("CARGO_PKG_VERSION")));
        assert!(
            value["build_id"]
                .as_str()
                .map(|id| !id.is_empty())
                .unwrap_or(false),
            "build_id should be a non-empty string, got {value:?}"
        );
    }

    // ---- config resolution ---------------------------------------------

    #[test]
    fn config_falls_back_to_dev_defaults() {
        let config = resolve_upload_config(&env_with(None, None));

        assert_eq!(config.url, DEFAULT_UPLOAD_URL);
        assert_eq!(config.bucket, DEFAULT_UPLOAD_BUCKET);
        assert!(!config.url_from_env);
        assert!(!config.bucket_from_env);
    }

    #[test]
    fn blank_env_values_are_treated_as_unset() {
        let config = resolve_upload_config(&env_with(Some("   "), Some("")));

        assert_eq!(config.url, DEFAULT_UPLOAD_URL);
        assert_eq!(config.bucket, DEFAULT_UPLOAD_BUCKET);
        assert!(!config.url_from_env);
        assert!(!config.bucket_from_env);
    }

    #[test]
    fn env_values_win_and_are_reported_as_env_sourced() {
        let config = resolve_upload_config(&env_with(
            Some("https://uploads.example.test"),
            Some("byteport-ci-uploads"),
        ));

        assert_eq!(config.url, "https://uploads.example.test");
        assert_eq!(config.bucket, "byteport-ci-uploads");
        assert!(config.url_from_env);
        assert!(config.bucket_from_env);
    }

    // ---- transport status ----------------------------------------------

    #[test]
    fn status_reports_defaults_and_partial_env_provenance() {
        let mut buffer = Vec::new();
        let config = resolve_upload_config(&env_with(Some("https://uploads.example.test"), None));
        cmd_transport_status(&config, &mut buffer).expect("status should render");

        let value: serde_json::Value =
            serde_json::from_slice(&buffer).expect("status output should be valid json");

        assert_eq!(value["url"], serde_json::json!("https://uploads.example.test"));
        assert_eq!(value["bucket"], serde_json::json!(DEFAULT_UPLOAD_BUCKET));
        assert_eq!(value["configured_via_env"]["url"], serde_json::json!(true));
        assert_eq!(value["configured_via_env"]["bucket"], serde_json::json!(false));
    }

    // ---- transport check ------------------------------------------------

    #[test]
    fn check_succeeds_and_reports_resolved_upload() {
        let config = resolve_upload_config(&env_with(
            Some("https://uploads.example.test"),
            Some("byteport-uploads"),
        ));

        let value = render_check(&config);

        let expected_key = format!("{UPLOAD_KEY_PREFIX}/{PROBE_OBJECT_KEY}");

        assert_eq!(value["ok"], serde_json::json!(true));
        assert_eq!(value["upload_id"], serde_json::json!(expected_key.clone()));
        assert_eq!(value["upload_id_kind"], serde_json::json!("object_key"));
        assert_eq!(
            value["endpoint"],
            serde_json::json!("https://uploads.example.test")
        );
        assert_eq!(value["bucket"], serde_json::json!("byteport-uploads"));
        assert_eq!(value["method"], serde_json::json!("PUT"));
        assert_eq!(
            value["target_url"],
            serde_json::json!(format!(
                "https://uploads.example.test/byteport-uploads/{expected_key}"
            ))
        );
        assert_eq!(
            value["headers"]["content-type"],
            serde_json::json!(PROBE_CONTENT_TYPE)
        );
        assert_eq!(
            value["headers"]["content-length"],
            serde_json::json!(PROBE_CONTENT_LENGTH.to_string())
        );
    }

    #[test]
    fn check_succeeds_on_default_configuration() {
        let config = resolve_upload_config(&env_with(None, None));
        let value = render_check(&config);

        assert_eq!(value["ok"], serde_json::json!(true));
        assert_eq!(value["endpoint"], serde_json::json!(DEFAULT_UPLOAD_URL));
    }

    #[test]
    fn check_rejects_non_http_endpoint() {
        let config = UploadConfig {
            url: "ftp://uploads.example.test".to_string(),
            bucket: "byteport-uploads".to_string(),
            url_from_env: true,
            bucket_from_env: false,
        };

        let error = cmd_transport_check(&config, &mut Vec::new())
            .expect_err("non-http endpoint should fail");

        assert!(
            matches!(error, CliError::Config(_)),
            "expected a config error, got {error:?}"
        );
    }

    #[test]
    fn check_rejects_empty_endpoint_and_bucket() {
        let empty_url = UploadConfig {
            url: String::new(),
            bucket: "byteport-uploads".to_string(),
            url_from_env: true,
            bucket_from_env: false,
        };
        assert!(cmd_transport_check(&empty_url, &mut Vec::new()).is_err());

        let empty_bucket = UploadConfig {
            url: "https://uploads.example.test".to_string(),
            bucket: "  ".to_string(),
            url_from_env: false,
            bucket_from_env: true,
        };
        assert!(cmd_transport_check(&empty_bucket, &mut Vec::new()).is_err());
    }

    // ---- deploy ---------------------------------------------------------

    fn render_deploy(target: &str) -> serde_json::Value {
        let mut buffer = Vec::new();
        deploy::cmd_deploy(target, &mut buffer).expect("deploy should render");
        serde_json::from_slice(&buffer).expect("deploy output should be valid json")
    }

    #[test]
    fn parses_deploy_command_with_target_flag() {
        let cli = Cli::try_parse_from(["byteport", "deploy", "--target", "docker"])
            .expect("deploy should parse");

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

        let stages = value["stages"]
            .as_array()
            .expect("plan should carry a stages array");
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

    // ---- helpers --------------------------------------------------------

    #[test]
    fn object_key_from_url_strips_endpoint_and_bucket() {
        let key = object_key_from_url(
            "https://uploads.example.test/byteport-uploads/desktop/demo.png",
            "https://uploads.example.test/",
            "byteport-uploads",
        );
        assert_eq!(key.as_deref(), Some("desktop/demo.png"));
    }

    #[test]
    fn object_key_from_url_rejects_foreign_urls() {
        assert_eq!(
            object_key_from_url(
                "https://other.example.test/byteport-uploads/desktop/demo.png",
                "https://uploads.example.test",
                "byteport-uploads",
            ),
            None
        );
    }
}
