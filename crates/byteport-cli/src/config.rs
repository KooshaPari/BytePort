//! XDG-aware configuration management for the BytePort CLI.
//!
//! The config file lives at:
//!
//! - `$XDG_CONFIG_HOME/byteport/config.toml` if `XDG_CONFIG_HOME` is set, or
//! - `$HOME/.config/byteport/config.toml` otherwise.
//!
//! It can also be overridden with the `BYTEPORT_CONFIG` environment variable
//! or the global `--config <path>` flag.
//!
//! ## Schema
//!
//! ```toml
//! current_profile = "default"
//!
//! [profiles.default]
//! name = "default"
//! server = "https://api.byteport.dev"
//! token = "<opaque-token>"  # encrypted at rest in a future release
//! default_project = "demo"
//!
//! [profiles.staging]
//! name = "staging"
//! server = "http://localhost:8080"
//! token = "dev-token"
//!
//! last_update_check = "2026-07-04T12:00:00Z"
//! ```
//!
//! Use [`Config::load`] to read the active config, [`Config::save`] to
//! persist it, and [`Config::path`] to query the resolved file location.

use std::collections::HashMap;
use std::fs;
use std::io;
use std::path::{Path, PathBuf};

use chrono::{DateTime, Utc};
use directories::ProjectDirs;
use serde::{Deserialize, Serialize};

use crate::errors::{CliError, Result};

/// A single named profile (server + credentials + default project).
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Profile {
    /// Profile name (e.g. "default", "staging").
    pub name: String,
    /// Base URL of the BytePort API.
    pub server: String,
    /// Optional opaque auth token (PAT / OAuth refresh / etc.).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub token: Option<String>,
    /// Optional default project for command invocations.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub default_project: Option<String>,
}

impl Profile {
    /// Create a new profile with the given name and server.
    pub fn new(name: impl Into<String>, server: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            server: server.into(),
            token: None,
            default_project: None,
        }
    }

    /// Builder-style: set the auth token.
    pub fn with_token(mut self, token: impl Into<String>) -> Self {
        self.token = Some(token.into());
        self
    }

    /// Builder-style: set the default project.
    pub fn with_default_project(mut self, project: impl Into<String>) -> Self {
        self.default_project = Some(project.into());
        self
    }
}

/// Top-level on-disk configuration.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, Default)]
pub struct Config {
    /// Name of the active profile.
    #[serde(default = "default_profile_name")]
    pub current_profile: String,
    /// All known profiles keyed by name.
    #[serde(default)]
    pub profiles: HashMap<String, Profile>,
    /// Timestamp of the last update check (cached for 24h).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub last_update_check: Option<DateTime<Utc>>,
}

fn default_profile_name() -> String {
    "default".to_string()
}

impl Config {
    /// Return the resolved config file path.
    ///
    /// Honors, in order:
    /// 1. Explicit override (e.g. from `--config <path>`).
    /// 2. `BYTEPORT_CONFIG` env var.
    /// 3. `XDG_CONFIG_HOME/byteport/config.toml` (or `$HOME/.config/byteport/config.toml`).
    pub fn path() -> Result<PathBuf> {
        if let Ok(p) = std::env::var("BYTEPORT_CONFIG") {
            if !p.is_empty() {
                return Ok(PathBuf::from(p));
            }
        }
        let dirs = ProjectDirs::from("dev", "byteport", "byteport")
            .ok_or_else(|| CliError::Config("could not resolve config directory".into()))?;
        Ok(dirs.config_dir().join("config.toml"))
    }

    /// Load the config from disk. If the file doesn't exist yet, return
    /// a default [`Config`] with a single "default" profile.
    pub fn load() -> Result<Self> {
        Self::load_from(&Self::path()?)
    }

    /// Load the config from an explicit path.
    pub fn load_from(path: &Path) -> Result<Self> {
        if !path.exists() {
            return Ok(Self::with_default_profile());
        }
        let raw = fs::read_to_string(path)?;
        if raw.trim().is_empty() {
            return Ok(Self::with_default_profile());
        }
        let cfg: Self = toml::from_str(&raw)?;
        // Ensure at least one profile exists.
        if cfg.profiles.is_empty() {
            return Ok(Self::with_default_profile());
        }
        Ok(cfg)
    }

    /// Construct a default config with a single "default" profile.
    pub fn with_default_profile() -> Self {
        let mut profiles = HashMap::new();
        profiles.insert(
            "default".to_string(),
            Profile::new("default", "https://api.byteport.dev"),
        );
        Self {
            current_profile: "default".to_string(),
            profiles,
            last_update_check: None,
        }
    }

    /// Save the config to disk, creating parent directories as needed.
    pub fn save(&self) -> Result<()> {
        let path = Self::path()?;
        self.save_to(&path)
    }

    /// Save the config to an explicit path.
    pub fn save_to(&self, path: &Path) -> Result<()> {
        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }
        let raw = toml::to_string_pretty(self)?;
        let tmp = path.with_extension("toml.tmp");
        fs::write(&tmp, raw)?;
        // Atomic rename.
        match fs::rename(&tmp, path) {
            Ok(()) => Ok(()),
            Err(e) => {
                // Clean up the temp file if rename fails.
                let _ = fs::remove_file(&tmp);
                Err(e.into())
            }
        }
    }

    /// Get the currently active profile. Falls back to "default".
    pub fn active_profile(&self) -> Result<&Profile> {
        self.profiles
            .get(&self.current_profile)
            .or_else(|| self.profiles.get("default"))
            .ok_or_else(|| {
                CliError::Config(format!(
                    "no profile named '{}' and no 'default' profile to fall back to",
                    self.current_profile
                ))
            })
    }

    /// Get a mutable reference to the active profile, creating it if missing.
    pub fn active_profile_mut(&mut self) -> Result<&mut Profile> {
        let name = self.current_profile.clone();
        if !self.profiles.contains_key(&name) {
            self.profiles
                .insert(name.clone(), Profile::new(&name, "https://api.byteport.dev"));
        }
        Ok(self.profiles.get_mut(&name).expect("just inserted"))
    }

    /// Touch the last-update-check timestamp to `now`.
    pub fn touch_update_check(&mut self) {
        self.last_update_check = Some(Utc::now());
    }

    /// Whether the update check should run (i.e. hasn't run in the last 24h).
    pub fn should_check_for_updates(&self, now: DateTime<Utc>) -> bool {
        match self.last_update_check {
            None => true,
            Some(last) => (now - last).num_hours() >= 24,
        }
    }
}

/// Clear stored credentials for the active profile (logout helper).
pub fn clear_active_token() -> Result<()> {
    let path = Config::path()?;
    if !path.exists() {
        return Ok(());
    }
    let mut cfg = Config::load_from(&path)?;
    if let Ok(profile) = cfg.active_profile_mut() {
        profile.token = None;
    }
    cfg.save_to(&path)?;
    Ok(())
}

/// Standard IO error helper for the `?` operator.
pub(crate) fn io_err(e: io::Error) -> CliError {
    CliError::Io(e)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn default_config_has_default_profile() {
        let cfg = Config::with_default_profile();
        assert_eq!(cfg.current_profile, "default");
        let profile = cfg.active_profile().unwrap();
        assert_eq!(profile.name, "default");
        assert!(profile.server.starts_with("http"));
    }

    #[test]
    fn profile_builder() {
        let p = Profile::new("prod", "https://prod.example.com")
            .with_token("secret")
            .with_default_project("alpha");
        assert_eq!(p.token.as_deref(), Some("secret"));
        assert_eq!(p.default_project.as_deref(), Some("alpha"));
    }

    #[test]
    fn round_trip_toml() {
        let tmp = tempfile::tempdir().unwrap();
        let p = tmp.path().join("config.toml");
        let mut cfg = Config::with_default_profile();
        cfg.profiles.insert(
            "staging".to_string(),
            Profile::new("staging", "http://localhost:8080").with_token("tok"),
        );
        cfg.save_to(&p).unwrap();
        let back = Config::load_from(&p).unwrap();
        assert_eq!(back.current_profile, cfg.current_profile);
        assert_eq!(back.profiles.len(), 2);
        assert_eq!(back.profiles["staging"].token.as_deref(), Some("tok"));
    }

    #[test]
    fn missing_file_yields_default() {
        let tmp = tempfile::tempdir().unwrap();
        let p = tmp.path().join("nope.toml");
        let cfg = Config::load_from(&p).unwrap();
        assert!(cfg.profiles.contains_key("default"));
    }

    #[test]
    fn empty_file_yields_default() {
        let tmp = tempfile::tempdir().unwrap();
        let p = tmp.path().join("empty.toml");
        fs::write(&p, "").unwrap();
        let cfg = Config::load_from(&p).unwrap();
        assert!(cfg.profiles.contains_key("default"));
    }

    #[test]
    fn should_check_for_updates_is_24h_cached() {
        let mut cfg = Config::default();
        let now = Utc::now();
        assert!(cfg.should_check_for_updates(now));
        cfg.touch_update_check();
        assert!(!cfg.should_check_for_updates(now));
        let later = now + chrono::Duration::hours(25);
        assert!(cfg.should_check_for_updates(later));
    }

    #[test]
    fn active_profile_falls_back_to_default() {
        let mut cfg = Config::default();
        cfg.profiles.clear();
        cfg.profiles.insert(
            "default".to_string(),
            Profile::new("default", "https://api.byteport.dev"),
        );
        cfg.current_profile = "missing".to_string();
        assert_eq!(cfg.active_profile().unwrap().name, "default");
    }
}