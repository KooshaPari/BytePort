//! Stub [`DockerEngine`] — returns `NotImplemented` for all operations.
//!
//! This placeholder exists so the engine registry has a `"docker"` entry
//! from day one without requiring the `bollard` crate or a running Docker
//! daemon. Real implementation is a follow-up task.

use async_trait::async_trait;
use tokio::sync::mpsc;

use crate::engine::{
    DeploymentId, DeploymentStatus, Engine, EngineError, EngineManifest, LogLine, LogOptions,
};

/// Stub Docker engine.
///
/// # Errors
///
/// All methods return [`EngineError::NotImplemented`].
#[derive(Debug)]
pub struct DockerEngine;

#[async_trait]
impl Engine for DockerEngine {
    fn name(&self) -> &'static str {
        "docker"
    }

    async fn deploy(&self, _manifest: EngineManifest) -> Result<DeploymentId, EngineError> {
        Err(EngineError::NotImplemented("docker::deploy"))
    }

    async fn status(&self, _id: &DeploymentId) -> Result<DeploymentStatus, EngineError> {
        Err(EngineError::NotImplemented("docker::status"))
    }

    async fn stop(&self, _id: &DeploymentId, _destroy: bool) -> Result<(), EngineError> {
        Err(EngineError::NotImplemented("docker::stop"))
    }

    async fn logs(
        &self,
        _id: &DeploymentId,
        _opts: LogOptions,
    ) -> Result<mpsc::Receiver<Result<LogLine, EngineError>>, EngineError> {
        Err(EngineError::NotImplemented("docker::logs"))
    }

    async fn list(&self) -> Result<Vec<DeploymentStatus>, EngineError> {
        Err(EngineError::NotImplemented("docker::list"))
    }
}
