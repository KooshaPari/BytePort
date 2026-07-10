//! byteport-engine-daemon — HTTP sidecar for the Engine trait.
//!
//! Listens on a Unix Domain Socket and exposes the Engine trait over REST.
//! The Go backend's `EngineDaemonClient` (HTTP over UDS) connects to this
//! socket at the path resolved from `BYTEPORT_ENGINE_SOCKET`
//! (default `/tmp/byteport-engine.sock`).
//!
//! # Configuration
//!
//! | Env var                  | Default                       | Description                      |
//! |--------------------------|-------------------------------|----------------------------------|
//! | `BYTEPORT_ENGINE_SOCKET` | `/tmp/byteport-engine.sock`   | UDS path the daemon binds to     |
//! | `NVMS_DAEMON_URL`        | *(not set)*                   | If set, registers nvms engine    |
//!
//! # Endpoints
//!
//! - `POST /deploy`                         — create a deployment
//! - `GET  /deployments/{id}`               — get deployment status
//! - `POST /deployments/{id}/stop`          — stop / destroy a deployment
//! - `GET  /deployments/{id}/logs`          — fetch recent log lines
//!
//! # Cleanup
//!
//! On startup the daemon removes any stale socket file left behind by a
//! previous crash. On shutdown the socket file is removed cleanly so the
//! next launch does not see a `EADDRINUSE` from the leftover inode.

use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use tower_http::cors::CorsLayer;
use tracing_subscriber::EnvFilter;

use byteport_engine::adapters::mock::MockEngine;
use byteport_engine::registry::EngineRegistry;
use byteport_engine::{
    DeploymentId, Engine, EngineError, EngineManifest, EnvVar, LogOptions, PortMapping,
    PortProtocol, ServiceSpec,
};

// ---------------------------------------------------------------------------
// Wire types
// ---------------------------------------------------------------------------

/// JSON body for `POST /deploy`.
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct DeployRequest {
    name: String,
    user: DeployUser,
    repository: String,
    services: Vec<ServiceEntry>,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct DeployUser {
    id: String,
    email: String,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct ServiceEntry {
    name: String,
    path: String,
    port: u16,
    env: Vec<EnvEntry>,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct EnvEntry {
    key: String,
    value: String,
}

/// JSON response for `GET /deployments/{id}`.
#[derive(Debug, Serialize)]
struct StatusResponse {
    deployment_id: String,
    state: String,
    urls: Vec<String>,
    ports: Vec<u16>,
    message: Option<String>,
}

/// JSON response for `GET /deployments/{id}/logs`.
#[derive(Debug, Serialize)]
struct LogLineResponse {
    line: String,
    stream: String,
    timestamp: String,
}

/// Uniform JSON error body.
#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    details: Option<String>,
}

// ---------------------------------------------------------------------------
// App state
// ---------------------------------------------------------------------------

#[derive(Clone)]
struct AppState {
    registry: Arc<EngineRegistry>,
}

// ---------------------------------------------------------------------------
// Engine selection
// ---------------------------------------------------------------------------

/// Pick the engine to dispatch to.
fn select_engine(state: &AppState) -> Result<&'static str, ErrorResponse> {
    let names = state.registry.names();
    if names.contains(&"nvms") {
        return Ok("nvms");
    }
    if names.contains(&"mock") {
        return Ok("mock");
    }
    Err(ErrorResponse {
        error: "no engines available".into(),
        details: None,
    })
}

/// Map an [`EngineError`] to an HTTP status and JSON error body.
fn err_to_parts(err: &EngineError) -> (StatusCode, ErrorResponse) {
    let (status, details) = match err {
        EngineError::NotFound(id) => (StatusCode::NOT_FOUND, Some(id.to_string())),
        EngineError::InvalidManifest(msg) => (StatusCode::BAD_REQUEST, Some(msg.clone())),
        EngineError::NotImplemented(feature) => {
            (StatusCode::NOT_IMPLEMENTED, Some((*feature).into()))
        }
        EngineError::Unavailable(msg) => (StatusCode::SERVICE_UNAVAILABLE, Some(msg.clone())),
        EngineError::RateLimited { .. } => (StatusCode::TOO_MANY_REQUESTS, None),
        EngineError::DeploymentFailed(msg) => {
            (StatusCode::INTERNAL_SERVER_ERROR, Some(msg.clone()))
        }
        EngineError::Io(e) => (StatusCode::INTERNAL_SERVER_ERROR, Some(e.to_string())),
        EngineError::Other(e) => (StatusCode::INTERNAL_SERVER_ERROR, Some(e.to_string())),
    };
    (
        status,
        ErrorResponse {
            error: err.to_string(),
            details,
        },
    )
}

/// Resolve an engine from the registry, short-circuiting with an error response
/// if none is available.
fn resolve_engine(
    state: &AppState,
) -> Result<&dyn Engine, (StatusCode, Json<ErrorResponse>)> {
    let name = select_engine(state).map_err(|e| (StatusCode::SERVICE_UNAVAILABLE, Json(e)))?;
    state
        .registry
        .get(name)
        .map_err(|e| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: format!("engine lookup failed: {e}"),
                    details: None,
                }),
            )
        })
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

/// `POST /deploy`
async fn handle_deploy(
    State(state): State<AppState>,
    Json(req): Json<DeployRequest>,
) -> impl IntoResponse {
    let engine = match resolve_engine(&state) {
        Ok(e) => e,
        Err(resp) => return resp.into_response(),
    };

    // Convert the wire-format deploy request into an EngineManifest.
    let entry = req.services.into_iter().next().unwrap_or(ServiceEntry {
        name: req.name.clone(),
        path: String::new(),
        port: 80,
        env: vec![],
    });

    let manifest = EngineManifest {
        name: req.name,
        service: ServiceSpec {
            image: entry.path,
            command: vec![],
            env: entry
                .env
                .into_iter()
                .map(|e| EnvVar {
                    key: e.key,
                    value: e.value,
                })
                .collect(),
            ports: vec![PortMapping {
                container_port: entry.port,
                host_port: None,
                protocol: PortProtocol::Tcp,
            }],
            resources: Default::default(),
            health_check_path: None,
            engine_config: None,
        },
        region: None,
    };

    match engine.deploy(manifest).await {
        Ok(id) => (
            StatusCode::CREATED,
            Json(serde_json::json!({ "deployment_id": id.0 })),
        )
            .into_response(),
        Err(e) => {
            let (status, body) = err_to_parts(&e);
            (status, Json(body)).into_response()
        }
    }
}

/// `GET /deployments/{id}`
async fn handle_get_deployment(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let engine = match resolve_engine(&state) {
        Ok(e) => e,
        Err(resp) => return resp.into_response(),
    };
    let did = DeploymentId(id);

    match engine.status(&did).await {
        Ok(status) => (
            StatusCode::OK,
            Json(StatusResponse {
                deployment_id: status.id.0,
                state: format!("{:?}", status.state),
                urls: status.urls,
                ports: status.ports,
                message: status.message,
            }),
        )
            .into_response(),
        Err(e) => {
            let (status, body) = err_to_parts(&e);
            (status, Json(body)).into_response()
        }
    }
}

/// `POST /deployments/{id}/stop`
async fn handle_stop_deployment(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let engine = match resolve_engine(&state) {
        Ok(e) => e,
        Err(resp) => return resp.into_response(),
    };
    let did = DeploymentId(id);

    match engine.stop(&did, true).await {
        Ok(()) => (
            StatusCode::OK,
            Json(serde_json::json!({ "status": "stopped" })),
        )
            .into_response(),
        Err(e) => {
            let (status, body) = err_to_parts(&e);
            (status, Json(body)).into_response()
        }
    }
}

/// `GET /deployments/{id}/logs`
async fn handle_get_logs(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let engine = match resolve_engine(&state) {
        Ok(e) => e,
        Err(resp) => return resp.into_response(),
    };

    let opts = LogOptions {
        tail: Some(100),
        follow: false,
        since: None,
    };

    let mut rx = match engine.logs(&DeploymentId(id), opts).await {
        Ok(rx) => rx,
        Err(e) => {
            let (status, body) = err_to_parts(&e);
            return (status, Json(body)).into_response();
        }
    };

    // Collect log lines with a reasonable timeout.
    let mut lines: Vec<LogLineResponse> = Vec::new();
    loop {
        match tokio::time::timeout(std::time::Duration::from_secs(3), rx.recv()).await {
            Ok(Some(Ok(line))) => {
                let stream = match line.stream {
                    byteport_engine::LogStream::Stdout => "stdout".into(),
                    byteport_engine::LogStream::Stderr => "stderr".into(),
                };
                lines.push(LogLineResponse {
                    line: line.line,
                    stream,
                    timestamp: line.timestamp.to_rfc3339(),
                });
            }
            Ok(Some(Err(e))) => {
                tracing::warn!(error = %e, "log line error");
            }
            Ok(None) => break,
            Err(_elapsed) => break,
        }
    }

    (StatusCode::OK, Json(lines)).into_response()
}

// ---------------------------------------------------------------------------
// Entrypoint
// ---------------------------------------------------------------------------

fn main() {
    // Initialise structured logging.
    tracing_subscriber::fmt()
        .with_env_filter(
            EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| EnvFilter::new("info")),
        )
        .init();

    // Read UDS socket path from the environment. The Go client uses the same
    // env var so a single env var drives both sides of the wire.
    let socket_path: String = std::env::var("BYTEPORT_ENGINE_SOCKET")
        .ok()
        .filter(|s| !s.is_empty())
        .unwrap_or_else(|| "/tmp/byteport-engine.sock".to_string());

    // Build the engine registry.
    let mut registry = EngineRegistry::new();

    // Always register the mock engine.
    registry
        .register("mock", Box::new(MockEngine::new()))
        .expect("mock engine registration");

    // Conditionally register the nvms engine.
    if std::env::var("NVMS_DAEMON_URL").is_ok() {
        #[cfg(feature = "nvms")]
        {
            let nvms = byteport_engine::adapters::nvms::NvmsHttpAdapter::from_env();
            registry
                .register("nvms", Box::new(nvms))
                .expect("nvms engine registration");
            tracing::info!("nvms engine registered");
        }
        #[cfg(not(feature = "nvms"))]
        {
            tracing::warn!(
                "NVMS_DAEMON_URL is set but the nvms feature is disabled; \
                 nvms engine will not be available"
            );
        }
    } else {
        tracing::info!("NVMS_DAEMON_URL not set — nvms engine skipped");
    }

    tracing::info!("registered engines: {:?}", registry.names());

    let state = AppState {
        registry: Arc::new(registry),
    };

    // Build the router with CORS support for development.
    let app = Router::new()
        .route("/deploy", post(handle_deploy))
        .route("/deployments/{id}", get(handle_get_deployment))
        .route("/deployments/{id}/stop", post(handle_stop_deployment))
        .route("/deployments/{id}/logs", get(handle_get_logs))
        .layer(CorsLayer::permissive())
        .with_state(state);

    // Create a single runtime for the whole program.
    let rt = tokio::runtime::Runtime::new().expect("tokio runtime");

    // Bind the Unix Domain Socket listener. We remove any stale socket file
    // first so a previous crash does not leave the path in `EADDRINUSE`.
    let listener = rt.block_on(async {
        if let Err(e) = std::fs::remove_file(&socket_path) {
            // ENOENT (file already gone) is the expected happy path; any
            // other error is worth surfacing before we attempt to bind.
            if e.kind() != std::io::ErrorKind::NotFound {
                tracing::warn!(error = %e, path = %socket_path, "could not remove stale socket file");
            }
        }
        tokio::net::UnixListener::bind(&socket_path)
    })
    .expect("bind UDS listener");

    tracing::info!("byteport-engine-daemon listening on unix://{socket_path}");

    // On shutdown, clean up the socket file so the next launch starts from
    // a clean slate. Use a guard so we run the cleanup even if axum panics.
    let socket_path_for_cleanup = socket_path.clone();
    let cleanup = move || {
        if let Err(e) = std::fs::remove_file(&socket_path_for_cleanup) {
            if e.kind() != std::io::ErrorKind::NotFound {
                tracing::warn!(error = %e, "failed to remove socket on shutdown");
            }
        }
    };

    // Serve requests. `axum::serve` accepts any `tokio::net::Listener`-like
    // type via the `IntoMakeService` adapter, but the `UnixListener` we hold
    // here satisfies the trait bound directly.
    let server = axum::serve(listener, app);
    rt.block_on(async {
        let result = server.await;
        cleanup();
        result.expect("axum server error");
    });
}
