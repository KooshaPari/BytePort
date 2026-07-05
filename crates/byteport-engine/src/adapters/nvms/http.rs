//! HTTP daemon adapter for NanoVMS.
//!
//! Communicates with a running `nanovms` daemon over REST. Both the daemon
//! and the API contract are under active development — this adapter returns
//! [`EngineError::NotImplemented`] for all operations until the daemon's API
//! surface is stable.
//!
//! # Configuration
//!
//! | Env var | Default | Description |
//! |---------|---------|-------------|
//! | `NVMS_DAEMON_URL` | `http://127.0.0.1:9700` | Base URL of the NanoVMS daemon |
//! | `NVMS_DAEMON_TIMEOUT` | `30` | Request timeout in seconds |

use async_trait::async_trait;
use tokio::sync::mpsc;

use crate::engine::{
    DeploymentId, DeploymentStatus, Engine, EngineError, EngineManifest, LogLine, LogOptions,
};

/// NanoVMS HTTP daemon adapter.
///
/// Requires a running `nanovms daemon` process. Without one, every method
/// returns [`EngineError::Unavailable`] with a descriptive message.
#[derive(Debug)]
#[allow(dead_code)]
pub struct NvmsHttpAdapter {
    base_url: String,
    client: reqwest::Client,
}

impl NvmsHttpAdapter {
    /// Create a new adapter reading config from the environment.
    pub fn from_env() -> Self {
        let base_url = std::env::var("NVMS_DAEMON_URL")
            .unwrap_or_else(|_| "http://127.0.0.1:9700".into());
        Self::new(&base_url)
    }

    /// Create a new adapter with an explicit base URL.
    pub fn new(base_url: &str) -> Self {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(
                std::env::var("NVMS_DAEMON_TIMEOUT")
                    .ok()
                    .and_then(|s| s.parse().ok())
                    .unwrap_or(30),
            ))
            .build()
            .expect("valid reqwest Client config");
        Self {
            base_url: base_url.trim_end_matches('/').to_owned(),
            client,
        }
    }
}

#[async_trait]
impl Engine for NvmsHttpAdapter {
    fn name(&self) -> &'static str {
        "nanovms"
    }

    async fn deploy(&self, _manifest: EngineManifest) -> Result<DeploymentId, EngineError> {
        Err(EngineError::NotImplemented("nanovms::deploy"))
    }

    async fn status(&self, _id: &DeploymentId) -> Result<DeploymentStatus, EngineError> {
        Err(EngineError::NotImplemented("nanovms::status"))
    }

    async fn stop(&self, _id: &DeploymentId, _destroy: bool) -> Result<(), EngineError> {
        Err(EngineError::NotImplemented("nanovms::stop"))
    }

    async fn logs(
        &self,
        _id: &DeploymentId,
        _opts: LogOptions,
    ) -> Result<mpsc::Receiver<Result<LogLine, EngineError>>, EngineError> {
        Err(EngineError::NotImplemented("nanovms::logs"))
    }

    async fn list(&self) -> Result<Vec<DeploymentStatus>, EngineError> {
        Err(EngineError::NotImplemented("nanovms::list"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn adapter_creates_with_default_url() {
        let adapter = NvmsHttpAdapter::new("http://localhost:9999");
        assert_eq!(adapter.base_url, "http://localhost:9999");
    }

    #[test]
    fn adapter_strips_trailing_slash() {
        let adapter = NvmsHttpAdapter::new("http://localhost:9999/");
        assert_eq!(adapter.base_url, "http://localhost:9999");
    }

    #[tokio::test]
    async fn deploy_returns_not_implemented() {
        let adapter = NvmsHttpAdapter::new("http://localhost:1");
        let manifest = EngineManifest {
            name: "x".into(),
            service: crate::engine::ServiceSpec {
                image: "i".into(),
                ..Default::default()
            },
            region: None,
        };
        let err = adapter.deploy(manifest).await.unwrap_err();
        assert!(matches!(err, EngineError::NotImplemented(_)));
    }
}
