//! BytePort integration crate — wires shared Phenotype crates
//! (phenotype-crypto, phenotype-observability, phenotype-health).

use serde::Deserialize;
use std::env;

/// BytePort application configuration.
#[derive(Debug, Clone, Deserialize, PartialEq)]
pub struct BytePortConfig {
    pub name: String,
    pub port: Option<u16>,
}

/// Error type for BytePort application operations.
#[derive(Debug, thiserror::Error)]
pub enum BytePortError {
    #[error("config error: {0}")]
    Config(String),
}

/// Canonical BytePort application struct.
///
/// Wires `phenotype-crypto`, `phenotype-observability`, and `phenotype-health`
/// into a single entry point.
#[derive(Debug, Clone)]
pub struct BytePortApp {
    pub config: BytePortConfig,
}

impl BytePortApp {
    /// Create a new `BytePortApp` by loading configuration from
    /// environment variables with the given `prefix`.
    ///
    /// Looks for `{PREFIX}_NAME` and `{PREFIX}_PORT`.
    pub fn new(prefix: &str) -> Result<Self, BytePortError> {
        let name =
            env::var(format!("{prefix}NAME")).map_err(|e| BytePortError::Config(format!("{prefix}NAME: {e}")))?;
        let port = env::var(format!("{prefix}PORT")).ok().and_then(|v| v.parse().ok());
        Ok(Self {
            config: BytePortConfig { name, port },
        })
    }

    /// Create a new `BytePortApp` with an explicit config value.
    pub fn with_config(config: BytePortConfig) -> Self {
        Self { config }
    }

    /// Run the application.
    pub fn run(&self) -> Result<(), BytePortError> {
        tracing::info!(app.name = %self.config.name, "BytePortApp started");
        Ok(())
    }
}

/// Re-export shared crypto utilities for downstream crates.
pub use phenotype_crypto;

/// Re-export shared observability for downstream crates.
pub use phenotype_observability;

/// Re-export shared health checks for downstream crates.
pub use phenotype_health;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn app_loads_config() {
        env::set_var("BPTEST_NAME", "byteport-test");
        env::set_var("BPTEST_PORT", "8080");

        let app = BytePortApp::new("BPTEST_").expect("app should load config");
        assert_eq!(app.config.name, "byteport-test");
        assert_eq!(app.config.port, Some(8080));
    }

    #[test]
    fn app_with_explicit_config() {
        let app = BytePortApp::with_config(BytePortConfig {
            name: "explicit".into(),
            port: Some(9090),
        });
        assert_eq!(app.config.name, "explicit");
        assert_eq!(app.config.port, Some(9090));
    }
}
