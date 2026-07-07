//! HTTP daemon adapter for NanoVMS.
//!
//! Communicates with a running `nanovms` Spin Wasm service over HTTP.
//! The NVMS Go service exposes [`POST /deploy`] and [`POST /terminate`]
//! endpoints that create / tear down AWS infrastructure (EC2, S3, ALB, etc.).
//!
//! # Configuration
//!
//! | Env var | Default | Description |
//! |---------|---------|-------------|
//! | `NVMS_DAEMON_URL` | `http://127.0.0.1:9700` | Base URL of the NVMS Spin service |
//! | `NVMS_DAEMON_TIMEOUT` | `30` | Request timeout in seconds |
//!
//! # Request / Response Contract
//!
//! The NVMS `/deploy` endpoint accepts a JSON body with the project/nvms manifest
//! and returns a JSON object keyed by service name → `InstanceInfo`:
//!
//! ```json
//! {
//!   "instanceId": "i-abc123",
//!   "publicIp": "1.2.3.4",
//!   "region": "us-east-1",
//!   "status": "running"
//! }
//! ```

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use tokio::sync::mpsc;

use crate::engine::{DeploymentId, DeploymentStatus, Engine, EngineError, EngineManifest, LogLine, LogOptions};

// ---------------------------------------------------------------------------
// NVMS wire types
// ---------------------------------------------------------------------------

/// Request body sent to the NVMS `/deploy` endpoint.
///
/// Mirrors the fields expected by the Go `projectManager.DeployProject` handler.
#[derive(Debug, Serialize)]
struct NvmsDeployRequest {
    name: String,
    user: NvmsUser,
    repository: String,
}

/// Subset of the Go `models.User` that NVMS reads for AWS credential resolution.
#[derive(Debug, Serialize)]
struct NvmsUser {
    id: String,
    email: String,
}

/// Response from the NVMS `/deploy` endpoint.
///
/// Keyed by service name (e.g. `"api"`, `"web"`), values describe the
/// provisioned EC2 instance(s).
#[derive(Debug, Deserialize)]
struct NvmsDeployResponse {
    #[serde(flatten)]
    services: std::collections::HashMap<String, NvmsInstanceInfo>,
}

/// A single provisioned instance returned by NVMS.
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct NvmsInstanceInfo {
    #[serde(alias = "instanceId", alias = "InstanceID")]
    instance_id: String,
    #[serde(alias = "publicIp", alias = "PublicIP", alias = "public_ip")]
    public_ip: Option<String>,
    region: String,
    status: String,
}

// ---------------------------------------------------------------------------
// Adapter
// ---------------------------------------------------------------------------

/// NVMS HTTP daemon adapter.
///
/// Communicates with the NVMS Spin Wasm service over plain HTTP.
/// All methods return [`EngineError::Unavailable`] if the NVMS daemon
/// cannot be reached.
#[derive(Debug)]
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
            .expect("valid reqwest Client builder");
        Self {
            base_url: base_url.trim_end_matches('/').to_owned(),
            client,
        }
    }

    /// Build the NVMS deploy URL.
    fn deploy_url(&self) -> String {
        format!("{}/deploy", self.base_url)
    }

    /// Build the NVMS terminate URL.
    fn terminate_url(&self) -> String {
        format!("{}/terminate", self.base_url)
    }

    /// Common error mapping for HTTP failures.
    fn map_err(e: reqwest::Error) -> EngineError {
        if e.is_timeout() {
            EngineError::Unavailable(format!("NVMS daemon timeout: {e}"))
        } else if e.is_connect() {
            EngineError::Unavailable(format!("NVMS daemon unreachable: {e}"))
        } else {
            EngineError::DeploymentFailed(format!("NVMS HTTP error: {e}"))
        }
    }
}

#[async_trait]
impl Engine for NvmsHttpAdapter {
    fn name(&self) -> &'static str {
        "nvms"
    }

    /// Deploy a project through NVMS.
    ///
    /// Sends a POST to `/deploy` with the manifest's name and inferred user.
    /// Returns a deployment ID derived from the first provisioned instance.
    ///
    /// # Errors
    ///
    /// Returns [`EngineError::Unavailable`] if the daemon is unreachable,
    /// [`EngineError::InvalidArgument`] if the response is unparseable,
    /// or [`EngineError::Internal`] for other HTTP errors.
    async fn deploy(&self, manifest: EngineManifest) -> Result<DeploymentId, EngineError> {
        let req_body = NvmsDeployRequest {
            name: manifest.name.clone(),
            user: NvmsUser {
                id: "byteport".into(),
                email: "byteport@localhost".into(),
            },
            repository: manifest.name.clone(),
        };

        let resp = self
            .client
            .post(&self.deploy_url())
            .json(&req_body)
            .send()
            .await
            .map_err(Self::map_err)?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            return Err(EngineError::DeploymentFailed(format!(
                "NVMS deploy returned {status}: {body}"
            )));
        }

        let deploy_resp: NvmsDeployResponse =
            resp.json().await.map_err(|e| {
                EngineError::InvalidManifest(format!(
                "NVMS deploy response parse error: {e}"
            ))
            })?;

        // Use the first service's instance ID as the deployment identifier.
        let first_id = deploy_resp
            .services
            .into_iter()
            .next()
            .map(|(name, info)| {
                if info.instance_id.is_empty() {
                    name
                } else {
                    info.instance_id
                }
            })
            .unwrap_or_else(|| manifest.name.clone());

        Ok(DeploymentId(first_id))
    }

    /// Get deployment status.
    ///
    /// NVMS does not expose a standalone status endpoint in the current
    /// contract, so we return [`EngineError::NotImplemented`].
    async fn status(&self, _id: &DeploymentId) -> Result<DeploymentStatus, EngineError> {
        // The current NVMS Go service has no GET /status endpoint.
        // When one is added (Phase 2), this will call it.
        Err(EngineError::NotImplemented("nvms::status"))
    }

    /// Stop (terminate) a deployment.
    ///
    /// Sends a POST to `/terminate` with the deployment identifier.
    async fn stop(&self, id: &DeploymentId, _destroy: bool) -> Result<(), EngineError> {
        let req_body = serde_json::json!({
            "name": id.0,
            "repository": "",
            "user": {
                "id": "byteport",
                "email": "byteport@localhost"
            }
        });

        let resp = self
            .client
            .post(&self.terminate_url())
            .json(&req_body)
            .send()
            .await
            .map_err(Self::map_err)?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            return Err(EngineError::DeploymentFailed(format!(
                "NVMS terminate returned {status}: {body}"
            )));
        }

        Ok(())
    }

    /// Stream logs for a deployment.
    ///
    /// NVMS does not expose a log endpoint in the current contract.
    /// Returns [`EngineError::NotImplemented`].
    async fn logs(
        &self,
        _id: &DeploymentId,
        _opts: LogOptions,
    ) -> Result<mpsc::Receiver<Result<LogLine, EngineError>>, EngineError> {
        Err(EngineError::NotImplemented("nvms::logs"))
    }

    /// List all active deployments.
    ///
    /// NVMS does not expose a list endpoint in the current contract.
    /// Returns [`EngineError::NotImplemented`].
    async fn list(&self) -> Result<Vec<DeploymentStatus>, EngineError> {
        Err(EngineError::NotImplemented("nvms::list"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::engine::{EngineManifest, ServiceSpec};

    #[test]
    fn test_from_env_defaults() {
        // Unset any existing env to test fallback.
        let prev = std::env::var("NVMS_DAEMON_URL").ok();
        std::env::remove_var("NVMS_DAEMON_URL");

        let adapter = NvmsHttpAdapter::from_env();
        assert_eq!(adapter.base_url, "http://127.0.0.1:9700");

        if let Some(url) = prev {
            std::env::set_var("NVMS_DAEMON_URL", url);
        }
    }

    #[test]
    fn test_custom_base_url() {
        let adapter = NvmsHttpAdapter::new("http://nvms:9800");
        assert_eq!(adapter.base_url, "http://nvms:9800");
        assert_eq!(adapter.deploy_url(), "http://nvms:9800/deploy");
        assert_eq!(
            adapter.terminate_url(),
            "http://nvms:9800/terminate"
        );
    }

    #[test]
    fn test_trailing_slash_stripped() {
        let adapter = NvmsHttpAdapter::new("http://localhost:9700/");
        assert_eq!(adapter.base_url, "http://localhost:9700");
    }

    #[test]
    fn test_deploy_request_serialization() {
        let req = NvmsDeployRequest {
            name: "my-app".into(),
            user: NvmsUser {
                id: "test-user".into(),
                email: "test@example.com".into(),
            },
            repository: "my-app".into(),
        };
        let json = serde_json::to_value(&req).unwrap();
        assert_eq!(json["name"], "my-app");
        assert_eq!(json["user"]["id"], "test-user");
        assert_eq!(json["repository"], "my-app");
    }

    #[tokio::test]
    async fn test_deploy_returns_error_without_daemon() {
        // Without a running NVMS daemon, deploy should fail with Unavailable.
        let adapter = NvmsHttpAdapter::new("http://127.0.0.1:1");
        let manifest = EngineManifest {
            name: "test".into(),
            service: ServiceSpec {
                image: "nginx:alpine".into(),
                ..Default::default()
            },
            region: None,
        };
        let result = adapter.deploy(manifest).await;
        assert!(
            result.is_err(),
            "expected error without running NVMS daemon"
        );
        match result {
            Err(EngineError::Unavailable(_)) => {} // expected
            other => panic!("expected Unavailable, got {other:?}"),
        }
    }

    // ------------------------------------------------------------------
    // Wiremock integration tests
    // ------------------------------------------------------------------

    #[tokio::test]
    async fn wiremock_deploy_success() {
        let mock_server = wiremock::MockServer::start().await;

        let body = r#"{"api":{"instanceId":"i-test01","publicIp":"1.2.3.4","region":"us-east-1","status":"running"}}"#;

        wiremock::Mock::given(wiremock::matchers::method("POST"))
            .and(wiremock::matchers::path("/deploy"))
            .respond_with(wiremock::ResponseTemplate::new(200).set_body_string(body))
            .expect(1)
            .mount(&mock_server)
            .await;

        let adapter = NvmsHttpAdapter::new(&mock_server.uri());
        let manifest = EngineManifest {
            name: "demo".into(),
            service: ServiceSpec {
                image: "nginx:alpine".into(),
                ..Default::default()
            },
            region: None,
        };
        let result = adapter.deploy(manifest).await.expect("deploy ok");
        assert_eq!(result.0, "i-test01");
    }

    #[tokio::test]
    async fn wiremock_deploy_http_500() {
        let mock_server = wiremock::MockServer::start().await;
        wiremock::Mock::given(wiremock::matchers::method("POST"))
            .and(wiremock::matchers::path("/deploy"))
            .respond_with(wiremock::ResponseTemplate::new(500).set_body_string("nvms kaput"))
            .mount(&mock_server)
            .await;

        let adapter = NvmsHttpAdapter::new(&mock_server.uri());
        let manifest = EngineManifest {
            name: "demo".into(),
            service: ServiceSpec::default(),
            region: None,
        };
        let result = adapter.deploy(manifest).await;
        match result {
            Err(EngineError::DeploymentFailed(msg)) => {
                assert!(msg.contains("500"), "msg = {msg}");
                assert!(msg.contains("nvms kaput"), "msg = {msg}");
            }
            other => panic!("expected DeploymentFailed(500), got {other:?}"),
        }
    }

    #[tokio::test]
    async fn wiremock_deploy_malformed_body() {
        let mock_server = wiremock::MockServer::start().await;
        wiremock::Mock::given(wiremock::matchers::method("POST"))
            .and(wiremock::matchers::path("/deploy"))
            .respond_with(wiremock::ResponseTemplate::new(200).set_body_string("not json"))
            .mount(&mock_server)
            .await;

        let adapter = NvmsHttpAdapter::new(&mock_server.uri());
        let manifest = EngineManifest {
            name: "demo".into(),
            service: ServiceSpec::default(),
            region: None,
        };
        let result = adapter.deploy(manifest).await;
        match result {
            Err(EngineError::InvalidManifest(_)) => {}
            other => panic!("expected InvalidManifest, got {other:?}"),
        }
    }

    #[tokio::test]
    async fn wiremock_stop_no_services() {
        // NVMS returns {} → adapter should fall back to manifest.name as ID fallback path
        let mock_server = wiremock::MockServer::start().await;
        wiremock::Mock::given(wiremock::matchers::method("POST"))
            .and(wiremock::matchers::path("/terminate"))
            .respond_with(wiremock::ResponseTemplate::new(200).set_body_string("{}"))
            .mount(&mock_server)
            .await;

        let adapter = NvmsHttpAdapter::new(&mock_server.uri());
        let id = crate::engine::DeploymentId("my-deploy".into());
        let result = adapter.stop(&id, false).await;
        assert!(result.is_ok(), "expected ok, got {result:?}");
    }

    #[tokio::test]
    async fn wiremock_stop_http_500() {
        let mock_server = wiremock::MockServer::start().await;
        wiremock::Mock::given(wiremock::matchers::method("POST"))
            .and(wiremock::matchers::path("/terminate"))
            .respond_with(wiremock::ResponseTemplate::new(500).set_body_string("boom"))
            .mount(&mock_server)
            .await;

        let adapter = NvmsHttpAdapter::new(&mock_server.uri());
        let id = crate::engine::DeploymentId("my-deploy".into());
        let result = adapter.stop(&id, false).await;
        match result {
            Err(EngineError::DeploymentFailed(msg)) => assert!(msg.contains("500")),
            other => panic!("expected DeploymentFailed, got {other:?}"),
        }
    }
}
