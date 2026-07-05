//! In-memory [`MockEngine`] for tests and local development.
//!
//! Stores deployments in a `HashMap` behind a `tokio::sync::RwLock`.
//! Deploy always succeeds; the caller controls behaviour via initial
//! configuration.

use async_trait::async_trait;
use tokio::sync::{mpsc, RwLock};

use crate::engine::{
    DeploymentId, DeploymentState, DeploymentStatus, Engine, EngineError, EngineManifest,
    LogLine, LogOptions,
};

/// In-memory mock engine for testing.
///
/// # Example
///
/// ```
/// use byteport_engine::engine::{
///     Engine, EngineManifest, ServiceSpec, Resources, PortMapping, PortProtocol,
/// };
/// use byteport_engine::adapters::mock::MockEngine;
///
/// #[tokio::test]
/// async fn mock_deploy_returns_id() {
///     let engine = MockEngine::new();
///     let manifest = EngineManifest {
///         name: "test".into(),
///         service: ServiceSpec {
///             image: "nginx:alpine".into(),
///             ..Default::default()
///         },
///         region: None,
///     };
///     let id = engine.deploy(manifest).await.expect("deploy");
///     assert!(!id.0.is_empty(), "deployment id must not be empty");
/// }
/// ```
#[derive(Debug)]
pub struct MockEngine {
    deployments: RwLock<std::collections::HashMap<DeploymentId, DeploymentState>>,
}

impl Default for MockEngine {
    fn default() -> Self {
        Self::new()
    }
}

impl MockEngine {
    pub fn new() -> Self {
        Self {
            deployments: RwLock::new(std::collections::HashMap::new()),
        }
    }
}

#[async_trait]
impl Engine for MockEngine {
    fn name(&self) -> &'static str {
        "mock"
    }

    async fn deploy(&self, manifest: EngineManifest) -> Result<DeploymentId, EngineError> {
        let id = DeploymentId(uuid::Uuid::new_v4().to_string());
        let mut map = self.deployments.write().await;
        map.insert(id.clone(), DeploymentState::Running);
        tracing::info!(deployment = %id, name = %manifest.name, "mock deployed");
        Ok(id)
    }

    async fn status(&self, id: &DeploymentId) -> Result<DeploymentStatus, EngineError> {
        let map = self.deployments.read().await;
        let state = map.get(id).copied().ok_or_else(|| {
            EngineError::NotFound(id.clone())
        })?;
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
        let (tx, rx) = mpsc::channel(16);
        // Drop tx immediately: no logs to send.
        drop(tx);
        Ok(rx)
    }

    async fn list(&self) -> Result<Vec<DeploymentStatus>, EngineError> {
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
