//! Core [`Engine`] trait and value types for the BytePort pluggable
//! deployment-engine abstraction.
//!
//! Every concrete backend (Docker, Firecracker via NanoVMS, Kubernetes,
//! AWS ECS, …) implements [`Engine`]. Callers dispatch by name through an
//! [`EngineRegistry`](crate::registry::EngineRegistry).
//!
//! # Type hierarchy
//!
//! ```text
//! Engine (trait)            ← implemented per backend
//! ├── deploy(Manifest)      → DeploymentId
//! ├── status(id)            → DeploymentStatus
//! ├── stop(id)              → ()
//! ├── logs(id, opts)        → LogStream
//! └── list()                → Vec<DeploymentStatus>
//!
//! EngineManifest            ← input to deploy()
//!   ├── service_spec        → container / task definition
//!   └── resources           → cpu, memory, replicas
//!
//! DeploymentStatus          ← returned by status()
//!   ├── id                  → DeploymentId
//!   ├── state               → DeploymentState enum
//!   ├── url / ports         → access info
//!   └── error               → optional error reason
//!
//! EngineError               ← returned by all fallible methods
//! ```

use std::time::Duration;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use thiserror::Error;

// ---------------------------------------------------------------------------
// Identifiers
// ---------------------------------------------------------------------------

/// Opaque, engine-assigned deployment identifier.
///
/// Engines are free to use any scheme (UUID, cloud-provider ARN, etc.).
/// The string form is for external references (HTTP API, CLI, DB).
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct DeploymentId(pub String);

impl std::fmt::Display for DeploymentId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        self.0.fmt(f)
    }
}

// ---------------------------------------------------------------------------
// Value types
// ---------------------------------------------------------------------------

/// Resources allocated to a deployment.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Resources {
    /// CPU shares (1024 = 1 core).
    pub cpu_shares: u64,
    /// Memory limit in MiB.
    pub memory_mib: u64,
    /// Desired number of replicas / instances.
    pub replicas: u32,
}

impl Default for Resources {
    fn default() -> Self {
        Self {
            cpu_shares: 1024,
            memory_mib: 512,
            replicas: 1,
        }
    }
}

/// Network port mapping.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PortMapping {
    /// Container / task port.
    pub container_port: u16,
    /// Host / load-balancer port (None → engine-assigned).
    pub host_port: Option<u16>,
    /// Protocol.
    pub protocol: PortProtocol,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PortProtocol {
    Tcp,
    Udp,
}

impl Default for PortProtocol {
    fn default() -> Self {
        Self::Tcp
    }
}

/// Environment variable.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnvVar {
    pub key: String,
    /// Value or secret reference (e.g. `secret://projects/42/DB_PASS`).
    pub value: String,
}

/// Full service specification passed to [`Engine::deploy`].
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ServiceSpec {
    /// OCI image reference (e.g. `docker.io/nginx:alpine`).
    pub image: String,
    /// Command override (empty = use image default).
    pub command: Vec<String>,
    /// Environment variables.
    pub env: Vec<EnvVar>,
    /// Ports to expose.
    pub ports: Vec<PortMapping>,
    /// Resource limits.
    pub resources: Resources,
    /// Health-check endpoint path (e.g. `/healthz`).
    pub health_check_path: Option<String>,
    /// Arbitrary engine-specific configuration (passed verbatim).
    pub engine_config: Option<serde_json::Value>,
}

/// Input to [`Engine::deploy`].
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EngineManifest {
    /// Deployment name / label.
    pub name: String,
    /// Service specification.
    pub service: ServiceSpec,
    /// Target region / zone (empty = engine default).
    pub region: Option<String>,
}

// ---------------------------------------------------------------------------
// Deployment state machine
// ---------------------------------------------------------------------------

/// High-level deployment state.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DeploymentState {
    /// Being provisioned.
    Deploying,
    /// Running and accepting traffic.
    Running,
    /// Stopped / scaled to zero.
    Stopped,
    /// Terminated / destroyed.
    Terminated,
    /// Transient failure; will retry.
    Degraded,
    /// Permanent failure; manual intervention needed.
    Failed,
}

impl DeploymentState {
    pub fn is_terminal(self) -> bool {
        matches!(self, Self::Terminated | Self::Failed)
    }

    pub fn is_running(self) -> bool {
        self == Self::Running
    }
}

/// Full deployment status returned by [`Engine::status`].
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeploymentStatus {
    pub id: DeploymentId,
    pub state: DeploymentState,
    /// Public URL(s) where the deployment is reachable.
    pub urls: Vec<String>,
    /// Assigned host ports, if applicable.
    pub ports: Vec<u16>,
    /// Human-readable status message or error description.
    pub message: Option<String>,
    /// Engine-specific status detail.
    pub engine_detail: Option<serde_json::Value>,
}

// ---------------------------------------------------------------------------
// Log streaming
// ---------------------------------------------------------------------------

/// A single log line from a deployment.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogLine {
    pub line: String,
    pub stream: LogStream,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum LogStream {
    Stdout,
    Stderr,
}

/// Options for [`Engine::logs`].
#[derive(Debug, Clone)]
pub struct LogOptions {
    /// Number of recent lines to tail.
    pub tail: Option<u32>,
    /// Follow (stream) new lines.
    pub follow: bool,
    /// Start time for log range (ignored for `follow: true`).
    pub since: Option<chrono::DateTime<chrono::Utc>>,
}

impl Default for LogOptions {
    fn default() -> Self {
        Self {
            tail: Some(100),
            follow: false,
            since: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

/// Errors returned by [`Engine`] methods.
#[derive(Debug, Error)]
pub enum EngineError {
    #[error("engine not implemented: {0}")]
    NotImplemented(&'static str),

    #[error("deployment failed: {0}")]
    DeploymentFailed(String),

    #[error("deployment not found: {0}")]
    NotFound(DeploymentId),

    #[error("invalid manifest: {0}")]
    InvalidManifest(String),

    #[error("rate-limited by upstream: retry after {wait:?}")]
    RateLimited { wait: Duration },

    #[error("engine unavailable: {0}")]
    Unavailable(String),

    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    #[error(transparent)]
    Other(Box<dyn std::error::Error + Send + Sync + 'static>),
}

impl From<String> for EngineError {
    fn from(s: String) -> Self {
        EngineError::DeploymentFailed(s)
    }
}

// ---------------------------------------------------------------------------
// Engine trait
// ---------------------------------------------------------------------------

/// Pluggable deployment engine.
///
/// Each variant (Docker, Firecracker, Kubernetes, AWS ECS, …) implements
/// this trait. The [`EngineRegistry`](crate::registry::EngineRegistry)
/// dispatches calls by name.
#[async_trait]
pub trait Engine: Send + Sync + std::fmt::Debug {
    /// Human-readable engine name (e.g. `"docker"`, `"firecracker"`,
    /// `"kubernetes"`, `"aws-ecs"`).
    fn name(&self) -> &'static str;

    /// Deploy a new service from a manifest.
    ///
    /// Returns a [`DeploymentId`] that can be used with [`status`],
    /// [`stop`], and [`logs`].
    async fn deploy(&self, manifest: EngineManifest) -> Result<DeploymentId, EngineError>;

    /// Poll the current status of a deployment.
    async fn status(&self, id: &DeploymentId) -> Result<DeploymentStatus, EngineError>;

    /// Stop (and optionally destroy) a deployment.
    ///
    /// If `destroy` is `true`, the deployment is permanently removed.
    /// If `false`, the deployment is stopped but may be restarted later.
    async fn stop(&self, id: &DeploymentId, destroy: bool) -> Result<(), EngineError>;

    /// Stream logs from a deployment.
    ///
    /// Returns a channel receiver that yields `LogLine`s as they arrive.
    /// The implementation is responsible for closing the channel when
    /// the log stream ends.
    async fn logs(
        &self,
        id: &DeploymentId,
        opts: LogOptions,
    ) -> Result<tokio::sync::mpsc::Receiver<Result<LogLine, EngineError>>, EngineError>;

    /// List all deployments managed by this engine.
    async fn list(&self) -> Result<Vec<DeploymentStatus>, EngineError>;
}
