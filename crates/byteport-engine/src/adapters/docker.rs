//! Docker Engine — production stub.
//!
//! The default build returns `EngineError::NotImplemented` for every
//! operation. This stub exists so the engine registry has a `"docker"`
//! entry from day one without requiring a running Docker daemon or
//! any optional dependencies in the default test matrix.
//!
//! Future session: replace the body with a `bollard`-backed implementation
//! gated behind a `--features docker` Cargo feature. Spec lives in
//! `plans/2026-07-06-nvms-service-v1.md` (parallel architecture) and
//! `plans/2026-07-04-byteport-evolution-v1.md` Phase 1c. The bollard
//! crate's API surface is in flux (0.14 → 0.15 broke several methods
//! we relied on); the next session will pin a specific bollard version
//! and target that exact API.

use async_trait::async_trait;
use tokio::sync::mpsc;

use crate::engine::{
    DeploymentId, DeploymentStatus, Engine, EngineError, EngineManifest, LogLine, LogOptions,
};

/// Stub Docker engine.
#[derive(Debug, Default)]
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