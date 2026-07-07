//! In-memory [`MockEngine`] for tests and local development.
//!
//! Stores deployments in a `HashMap` behind a `tokio::sync::RwLock`.
//! Deploy always succeeds; the caller controls behaviour via initial
//! configuration through [`MockEngine::with_config`].
//!
//! # Failure-injection
//!
//! ```rust
//! use byteport_engine::adapters::mock::{MockEngine, MockConfig};
//!
//! // Fail every deploy.
//! let engine = MockEngine::with_config(MockConfig {
//!     deploy_error: Some("image pull failed".into()),
//!     ..Default::default()
//! });
//! ```
//!
//! # Latency simulation
//!
//! ```rust
//! use std::time::Duration;
//! use byteport_engine::adapters::mock::{MockEngine, MockConfig};
//!
//! let engine = MockEngine::with_config(MockConfig {
//!     operation_delay: Some(Duration::from_millis(250)),
//!     ..Default::default()
//! });
//! ```

use std::collections::HashMap;
use std::time::Duration;

use async_trait::async_trait;
use tokio::sync::{mpsc, RwLock};
use tokio::time::sleep;

use crate::engine::{
    DeploymentId, DeploymentState, DeploymentStatus, Engine, EngineError, EngineManifest, LogLine,
    LogOptions, LogStream,
};

/// Tunable behaviour for [`MockEngine`]. Used by tests to simulate failure
/// modes and latency.
#[derive(Debug, Clone, Default)]
pub struct MockConfig {
    /// If `Some`, every [`Engine::deploy`] call returns
    /// [`EngineError::DeploymentFailed`] with this message.
    pub deploy_error: Option<String>,
    /// Optional artificial delay applied to every operation.
    pub operation_delay: Option<Duration>,
    /// Initial deployment ID → state to seed the registry with.
    pub seed: HashMap<DeploymentId, DeploymentState>,
    /// If `true`, log streaming produces one synthetic line and closes.
    /// If `false` (default), log streaming produces nothing (matches
    /// Docker's behaviour when a container has not written anything).
    pub emit_one_log_line: bool,
}

/// In-memory mock engine for testing and local development.
///
/// See [module docs](self) for failure-injection examples.
#[derive(Debug)]
pub struct MockEngine {
    deployments: RwLock<HashMap<DeploymentId, DeploymentState>>,
    config: MockConfig,
}

impl Default for MockEngine {
    fn default() -> Self {
        Self::new()
    }
}

impl MockEngine {
    /// Create an empty mock engine with default config (no failures, no
    /// latency).
    pub fn new() -> Self {
        Self::with_config(MockConfig::default())
    }

    /// Create a mock engine with the given config. Seeded deployments
    /// become immediately visible via [`list`](Engine::list).
    pub fn with_config(config: MockConfig) -> Self {
        Self {
            deployments: RwLock::new(config.seed.clone()),
            config,
        }
    }

    /// Test helper: snapshot the current deployment map.
    pub async fn snapshot(&self) -> HashMap<DeploymentId, DeploymentState> {
        self.deployments.read().await.clone()
    }

    async fn apply_delay(&self) {
        if let Some(d) = self.config.operation_delay {
            sleep(d).await;
        }
    }
}

#[async_trait]
impl Engine for MockEngine {
    fn name(&self) -> &'static str {
        "mock"
    }

    async fn deploy(&self, manifest: EngineManifest) -> Result<DeploymentId, EngineError> {
        self.apply_delay().await;
        if let Some(ref msg) = self.config.deploy_error {
            return Err(EngineError::DeploymentFailed(msg.clone()));
        }
        let id = DeploymentId(uuid::Uuid::new_v4().to_string());
        let mut map = self.deployments.write().await;
        map.insert(id.clone(), DeploymentState::Running);
        tracing::info!(deployment = %id, name = %manifest.name, "mock deployed");
        Ok(id)
    }

    async fn status(&self, id: &DeploymentId) -> Result<DeploymentStatus, EngineError> {
        self.apply_delay().await;
        let map = self.deployments.read().await;
        let state = map.get(id).copied().ok_or_else(|| EngineError::NotFound(id.clone()))?;
        Ok(DeploymentStatus {
            id: id.clone(),
            state,
            urls: vec![],
            ports: vec![],
            message: None,
            engine_detail: None,
        })
    }

    async fn stop(&self, id: &DeploymentId, destroy: bool) -> Result<(), EngineError> {
        self.apply_delay().await;
        let mut map = self.deployments.write().await;
        if destroy {
            map.remove(id);
        } else {
            map.insert(id.clone(), DeploymentState::Stopped);
        }
        Ok(())
    }

    async fn logs(
        &self,
        _id: &DeploymentId,
        _opts: LogOptions,
    ) -> Result<mpsc::Receiver<Result<LogLine, EngineError>>, EngineError> {
        self.apply_delay().await;
        let (tx, rx) = mpsc::channel(16);
        if self.config.emit_one_log_line {
            let line = LogLine {
                line: "mock: hello world".to_string(),
                stream: LogStream::Stdout,
                timestamp: chrono::Utc::now(),
            };
            // Best-effort send; receiver may be dropped.
            let _ = tx.send(Ok(line)).await;
        }
        // Drop tx → channel closes → receiver returns None on next poll.
        drop(tx);
        Ok(rx)
    }

    async fn list(&self) -> Result<Vec<DeploymentStatus>, EngineError> {
        self.apply_delay().await;
        let map = self.deployments.read().await;
        Ok(map
            .iter()
            .map(|(id, state)| DeploymentStatus {
                id: id.clone(),
                state: *state,
                urls: vec![],
                ports: vec![],
                message: None,
                engine_detail: None,
            })
            .collect())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::engine::{Resources, ServiceSpec};

    fn manifest(name: &str) -> EngineManifest {
        EngineManifest {
            name: name.into(),
            service: ServiceSpec {
                image: "nginx:alpine".into(),
                ..Default::default()
            },
            region: None,
        }
    }

    #[tokio::test]
    async fn deploy_returns_unique_id() {
        let engine = MockEngine::new();
        let a = engine.deploy(manifest("a")).await.unwrap();
        let b = engine.deploy(manifest("b")).await.unwrap();
        assert_ne!(a, b);
    }

    #[tokio::test]
    async fn status_unknown_returns_not_found() {
        let engine = MockEngine::new();
        let err = engine.status(&DeploymentId("missing".into())).await.unwrap_err();
        assert!(matches!(err, EngineError::NotFound(_)));
    }

    #[tokio::test]
    async fn stop_destroy_removes() {
        let engine = MockEngine::new();
        let id = engine.deploy(manifest("a")).await.unwrap();
        engine.stop(&id, true).await.unwrap();
        let snap = engine.snapshot().await;
        assert!(!snap.contains_key(&id));
    }

    #[tokio::test]
    async fn stop_keep_keeps_in_map() {
        let engine = MockEngine::new();
        let id = engine.deploy(manifest("a")).await.unwrap();
        engine.stop(&id, false).await.unwrap();
        let status = engine.status(&id).await.unwrap();
        assert_eq!(status.state, DeploymentState::Stopped);
    }

    #[tokio::test]
    async fn deploy_error_configured_returns_err() {
        let engine = MockEngine::with_config(MockConfig {
            deploy_error: Some("image pull failed".into()),
            ..Default::default()
        });
        let err = engine.deploy(manifest("a")).await.unwrap_err();
        assert!(matches!(err, EngineError::DeploymentFailed(m) if m == "image pull failed"));
    }

    #[tokio::test]
    async fn operation_delay_is_applied() {
        use std::time::Instant;
        let engine = MockEngine::with_config(MockConfig {
            operation_delay: Some(Duration::from_millis(50)),
            ..Default::default()
        });
        let start = Instant::now();
        let _ = engine.list().await.unwrap();
        let elapsed = start.elapsed();
        assert!(elapsed >= Duration::from_millis(50));
    }

    #[tokio::test]
    async fn seed_makes_deployments_visible() {
        let mut seed = HashMap::new();
        let id = DeploymentId("seeded".into());
        seed.insert(id.clone(), DeploymentState::Running);
        let engine = MockEngine::with_config(MockConfig {
            seed,
            ..Default::default()
        });
        let listed = engine.list().await.unwrap();
        assert_eq!(listed.len(), 1);
        assert_eq!(listed[0].id, id);
    }

    #[tokio::test]
    async fn logs_emit_one_line_when_configured() {
        let engine = MockEngine::with_config(MockConfig {
            emit_one_log_line: true,
            ..Default::default()
        });
        let id = DeploymentId("any".into());
        let mut rx = engine.logs(&id, LogOptions::default()).await.unwrap();
        let first = rx.recv().await.expect("one line").unwrap();
        assert_eq!(first.line, "mock: hello world");
        assert_eq!(first.stream, LogStream::Stdout);
        let next = rx.recv().await;
        assert!(next.is_none(), "channel closed after one line");
    }

    #[tokio::test]
    async fn logs_default_emit_nothing() {
        let engine = MockEngine::new();
        let id = DeploymentId("any".into());
        let mut rx = engine.logs(&id, LogOptions::default()).await.unwrap();
        let next = rx.recv().await;
        assert!(next.is_none(), "no lines emitted by default");
    }

    #[test]
    fn default_resources_apply_to_manifest() {
        let m = manifest("x");
        assert_eq!(m.service.resources, Resources::default());
        assert_eq!(m.service.resources.replicas, 1);
    }
}