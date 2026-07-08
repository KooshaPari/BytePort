//! Startup update check.
//!
//! BytePort CLI checks for newer versions once every 24h (configurable in
//! [`Config`](crate::config::Config)) and prints a friendly reminder if
//! an upgrade is available. The check is:
//!
//! - **Opt-out**: `BYTEPORT_NO_UPDATE_CHECK=1` or `--no-update-check` disables it.
//! - **CI-aware**: when `CI`, `CONTINUOUS_INTEGRATION`, or `BYTEPORT_CI=1`
//!   is set, the check is skipped entirely.
//! - **Async** but synchronous at the call site: the check uses
//!   [`tokio::task::spawn_blocking`] under the hood so the rest of the
//!   CLI is never blocked on GitHub.
//!
//! The remote endpoint is `https://api.github.com/repos/KooshaPari/BytePort/releases/latest`.

use chrono::Utc;
use update_checker::{Check, Update};

use crate::config::Config;
use crate::errors::{CliError, Result};

/// GitHub repository in `owner/name` form.
const REPO: &str = "KooshaPari/BytePort";

/// `BYTEPORT_NO_UPDATE_CHECK` env var — disables the update check.
const ENV_NO_UPDATE_CHECK: &str = "BYTEPORT_NO_UPDATE_CHECK";

/// `BYTEPORT_CI` env var — signal that we're running in CI.
const ENV_CI_FLAG: &str = "BYTEPORT_CI";

/// Check whether the user has opted out of the update check.
///
/// Returns `true` if the check should be **skipped**.
pub fn is_disabled() -> bool {
    if std::env::var_os(ENV_NO_UPDATE_CHECK)
        .map(|v| !v.is_empty())
        .unwrap_or(false)
    {
        return true;
    }
    if std::env::var_os(ENV_CI_FLAG)
        .map(|v| !v.is_empty())
        .unwrap_or(false)
    {
        return true;
    }
    if std::env::var_os("CI")
        .map(|v| !v.is_empty())
        .unwrap_or(false)
    {
        return true;
    }
    if std::env::var_os("CONTINUOUS_INTEGRATION")
        .map(|v| !v.is_empty())
        .unwrap_or(false)
    {
        return true;
    }
    false
}

/// Synchronous wrapper around `update_checker::Check`.
///
/// Returns the newer version string if one is available, `None` otherwise.
/// Network failures and parse errors are swallowed (logged at debug) so
/// the CLI never fails to start because of an update-check hiccup.
pub fn check_blocking() -> Option<String> {
    if is_disabled() {
        return None;
    }
    let current = env!("CARGO_PKG_VERSION");
    let check = Check::builder()
        .with_owner("KooshaPari")
        .with_repository("BytePort")
        .with_current_version(current)
        .build()
        .ok()?;
    let update = check.check().ok()?;
    update.is_newer().then(|| update.version.to_string())
}

/// Async version — runs the blocking check on a blocking thread.
pub async fn check() -> Option<String> {
    if is_disabled() {
        return None;
    }
    tokio::task::spawn_blocking(check_blocking)
        .await
        .ok()
        .flatten()
}

/// Run the update check, honoring the 24h cache in [`Config`].
///
/// If a new version is available, returns the version string. The caller
/// is responsible for printing it.
pub fn run_with_cache(cfg: &mut Config) -> Result<Option<String>> {
    if is_disabled() {
        return Ok(None);
    }
    if !cfg.should_check_for_updates(Utc::now()) {
        return Ok(None);
    }
    // Touch the timestamp *before* the network call so we don't hammer
    // GitHub even if the request fails.
    cfg.touch_update_check();
    let _ = cfg.save();
    Ok(check_blocking())
}

/// Convenience: convert the result into a printable one-line notice.
pub fn format_notice(newer: &str) -> String {
    format!(
        "📦 BytePort {newer} is available — run `byteport update` to learn more (or visit https://github.com/{REPO}/releases)."
    )
}

/// One-shot helper: load config, run the cached check, save config, print notice.
pub fn run_startup_check() -> Result<()> {
    if is_disabled() {
        return Ok(());
    }
    let mut cfg = match Config::load() {
        Ok(c) => c,
        Err(e) => {
            tracing::debug!("update_check: skipping — config load failed: {e}");
            return Ok(());
        }
    };
    match run_with_cache(&mut cfg) {
        Ok(Some(newer)) => {
            eprintln!("\n{}\n", format_notice(&newer));
        }
        Ok(None) => {}
        Err(e) => {
            tracing::debug!("update_check: error: {e}");
        }
    }
    Ok(())
}

/// Internal: lookup the repository name (exposed for tests).
pub fn repo() -> &'static str {
    REPO
}

/// Suppress update-check warnings for testing (returns true if currently disabled).
pub fn disabled_for_test() -> bool {
    is_disabled()
}

/// A stub result for environments where the check is impossible.
pub fn unavailable(reason: impl Into<String>) -> CliError {
    CliError::Network(format!("update check unavailable: {}", reason.into()))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn repo_is_set() {
        assert_eq!(repo(), "KooshaPari/BytePort");
    }

    #[test]
    fn disabled_when_env_set() {
        let prev = std::env::var_os(ENV_NO_UPDATE_CHECK);
        // SAFETY: tests run single-threaded.
        unsafe {
            std::env::set_var(ENV_NO_UPDATE_CHECK, "1");
        }
        assert!(is_disabled());
        match prev {
            Some(v) => unsafe {
                std::env::set_var(ENV_NO_UPDATE_CHECK, v);
            },
            None => unsafe {
                std::env::remove_var(ENV_NO_UPDATE_CHECK);
            },
        }
    }

    #[test]
    fn disabled_when_ci_env_set() {
        let prev = std::env::var_os("CI");
        unsafe {
            std::env::set_var("CI", "1");
        }
        assert!(is_disabled());
        match prev {
            Some(v) => unsafe {
                std::env::set_var("CI", v);
            },
            None => unsafe {
                std::env::remove_var("CI");
            },
        }
    }

    #[test]
    fn format_notice_is_friendly() {
        let s = format_notice("0.2.0");
        assert!(s.contains("0.2.0"));
        assert!(s.contains("https://"));
    }

    #[test]
    fn check_blocking_returns_none_when_disabled() {
        let prev = std::env::var_os(ENV_NO_UPDATE_CHECK);
        unsafe {
            std::env::set_var(ENV_NO_UPDATE_CHECK, "1");
        }
        assert_eq!(check_blocking(), None);
        match prev {
            Some(v) => unsafe {
                std::env::set_var(ENV_NO_UPDATE_CHECK, v);
            },
            None => unsafe {
                std::env::remove_var(ENV_NO_UPDATE_CHECK);
            },
        }
    }
}