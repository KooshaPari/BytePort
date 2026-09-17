//! `byteport` desktop app — Tauri 2.x entry point.
//!
//! Wires the [`byteport_transport::S3UploadTransport`] into a Tauri
//! `State` so the frontend can request upload instructions over the IPC
//! bridge. Tracing is initialised once per process via the
//! `tracing-subscriber` JSON layer (the `tauri-plugin-log` plugin
//! augments this with per-window log capture for the in-app console).

#![deny(missing_docs)]
#![forbid(unsafe_code)]
#![warn(clippy::all)]

use std::sync::Arc;

use byteport_transport::{S3UploadTransport, UploadTransport};
use tauri::Manager;

/// Shared application state — kept as an `Arc<dyn UploadTransport>` so
/// additional transports (SOCKS5, WebDAV, …) can be swapped in via
/// feature flags without changing the IPC contract.
pub struct AppState {
    /// The active upload transport.
    pub uploader: Arc<dyn UploadTransport>,
}

/// Application-level error type for the desktop IPC bridge.
///
/// Wraps transport errors and configuration errors with enough context
/// for the frontend to render a useful message. The `thiserror` derive
/// is intentionally not used here to keep the crate free of one more
/// dependency; a manual impl is only ~15 lines.
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    /// The requested transport is not configured (missing env var,
    /// invalid endpoint URL, etc.).
    #[error("transport not configured: {0}")]
    NotConfigured(String),

    /// The underlying transport returned an error.
    #[error("upload transport error: {0}")]
    Transport(#[from] byteport_transport::UploadTransportError),

    /// Generic I/O error.
    #[error("i/o error: {0}")]
    Io(#[from] std::io::Error),

    /// Serialization error (JSON, base64, etc.).
    #[error("serialization error: {0}")]
    Serde(#[from] serde_json::Error),

    /// Catch-all for any other error, preserving the source message.
    #[error("internal error: {0}")]
    Other(String),
}

// `tauri::ipc::IpcResponse` only requires `serde::Serialize`, but the
// frontend (SvelteKit) expects a stable JSON shape with a `kind` tag so
// the renderer can pick the right user-facing message.
impl serde::Serialize for AppError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut s = serializer.serialize_struct("AppError", 2)?;
        let kind = match self {
            Self::NotConfigured(_) => "not_configured",
            Self::Transport(_) => "transport",
            Self::Io(_) => "io",
            Self::Serde(_) => "serde",
            Self::Other(_) => "other",
        };
        s.serialize_field("kind", kind)?;
        s.serialize_field("message", &self.to_string())?;
        s.end()
    }
}

/// Read the upload endpoint from `BYTEPORT_UPLOAD_URL` or fall back to the
/// local dev default. Centralised so tests can stub the env.
fn upload_endpoint() -> String {
    std::env::var("BYTEPORT_UPLOAD_URL").unwrap_or_else(|_| "https://uploads.byteport.local".to_string())
}

/// Read the upload bucket from `BYTEPORT_UPLOAD_BUCKET` or fall back to
/// the dev default.
fn upload_bucket() -> String {
    std::env::var("BYTEPORT_UPLOAD_BUCKET").unwrap_or_else(|_| "byteport-uploads".to_string())
}

/// Tauri entry point.
#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    // Logging is handled by tauri_plugin_log below; do NOT initialise
    // tracing_subscriber here — it claims the global `log` facade and
    // conflicts with tauri_plugin_log's logger registration.

    let uploader: Arc<dyn UploadTransport> = Arc::new(S3UploadTransport::new(
        upload_endpoint(),
        upload_bucket(),
        Some("desktop"),
    ));

    tauri::Builder::default()
        .plugin(
            tauri_plugin_log::Builder::default()
                .level(log::LevelFilter::Info)
                .build(),
        )
        .setup(move |app| {
            // Register shared state for IPC handlers.
            app.manage(AppState {
                uploader: Arc::clone(&uploader),
            });
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![ipc::create_upload, ipc::health_check])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

/// IPC handler module — kept in its own file as the surface grows.
pub mod ipc {
    use std::sync::Arc;

    use byteport_transport::{UploadRequest, UploadTransport};
    use serde::{Deserialize, Serialize};
    use tauri::State;

    use super::AppState;

    /// Environment variable that overrides the backend base URL probed by
    /// [`health_check`].
    const BACKEND_URL_ENV: &str = "BYTEPORT_BACKEND_URL";

    /// Default backend base URL — matches the Go API server's default port
    /// (`PORT` unset falls back to `8080` in `backend/main.go`).
    const BACKEND_URL_DEFAULT: &str = "http://localhost:8080";

    /// Health endpoint path exposed by the Go API server (`GET /health`).
    const HEALTH_PATH: &str = "/health";

    /// Hard timeout for a single probe, in seconds (`curl --max-time`).
    const HEALTH_TIMEOUT_SECS: u64 = 2;

    /// Upper bound on the response body handed back to the frontend.
    const HEALTH_BODY_LIMIT: usize = 4_096;

    /// IPC request body for `create_upload`.
    #[derive(Debug, Deserialize)]
    pub struct CreateUploadArgs {
        /// Object key (e.g. `/projects/demo/screenshot.png`).
        pub object_key: String,
        /// MIME type (e.g. `image/png`).
        pub content_type: String,
        /// Payload size in bytes.
        pub content_length: u64,
    }

    /// IPC response body for `create_upload`.
    #[derive(Debug, Serialize)]
    pub struct CreateUploadResponse {
        /// HTTP method to use for the upload.
        pub method: String,
        /// Pre-signed URL to `PUT` the bytes to.
        pub url: String,
        /// Headers to attach to the `PUT` (content-type, content-length, etc.).
        pub headers: std::collections::BTreeMap<String, String>,
    }

    /// Tauri command: build a pre-signed upload instruction for the
    /// frontend to perform. Returns the method, URL, and required
    /// headers.
    #[tauri::command]
    pub fn create_upload(state: State<'_, AppState>, args: CreateUploadArgs) -> Result<CreateUploadResponse, String> {
        let uploader: Arc<dyn UploadTransport> = Arc::clone(&state.uploader);
        let req = UploadRequest {
            object_key: args.object_key,
            content_type: args.content_type,
            content_length: args.content_length,
        };
        uploader
            .create_upload(&req)
            .map(|instr| CreateUploadResponse {
                method: instr.method,
                url: instr.url,
                headers: instr.headers,
            })
            .map_err(|e| e.to_string())
    }

    /// IPC response body for [`health_check`].
    #[derive(Debug, Clone, PartialEq, Eq, Serialize)]
    pub struct HealthStatus {
        /// `true` when the backend completed an HTTP exchange within the
        /// timeout. A refused connection or timeout yields `false` rather
        /// than an error.
        pub backend_reachable: bool,
        /// HTTP status code, or `None` when no response was received.
        pub status_code: Option<u16>,
        /// Response body truncated to [`HEALTH_BODY_LIMIT`]. When the probe
        /// fails before an HTTP exchange, this carries the `curl` diagnostic
        /// instead so the UI can explain the failure.
        pub response_body: Option<String>,
        /// Probe wall-clock latency in milliseconds.
        pub latency_ms: u64,
    }

    /// Resolve the absolute health endpoint URL.
    ///
    /// Honours [`BACKEND_URL_ENV`] so alternate ports/ hosts can be probed
    /// without a rebuild, and tolerates a trailing slash on the base URL.
    fn health_endpoint() -> String {
        let base = std::env::var(BACKEND_URL_ENV).unwrap_or_else(|_| BACKEND_URL_DEFAULT.to_string());
        format!("{}{HEALTH_PATH}", base.trim_end_matches('/'))
    }

    /// Truncate a body to `limit` bytes on a UTF-8 char boundary.
    fn truncate_body(raw: &str, limit: usize) -> String {
        if raw.len() <= limit {
            return raw.to_string();
        }
        let mut end = limit;
        while end > 0 && !raw.is_char_boundary(end) {
            end -= 1;
        }
        format!("{}...", &raw[..end])
    }

    /// Interpret `curl` output into a [`HealthStatus`].
    ///
    /// `curl` is invoked with `--write-out '\n%{http_code}'`, so the final
    /// line of stdout is the status code and everything before it is the
    /// body. `%{http_code}` is `000` when no response was received.
    fn parse_probe_output(stdout: &str, stderr: &str, latency_ms: u64) -> HealthStatus {
        let (body, code) = match stdout.rsplit_once('\n') {
            Some((body, code)) => (body, code.trim()),
            None => ("", stdout.trim()),
        };

        let status_code = code.parse::<u16>().ok().filter(|code| *code != 0);
        let backend_reachable = status_code.is_some();

        let trimmed_body = body.trim_matches('\n');
        let response_body = if !trimmed_body.is_empty() {
            Some(truncate_body(trimmed_body, HEALTH_BODY_LIMIT))
        } else if backend_reachable {
            None
        } else {
            let diagnostic = stderr.trim();
            if diagnostic.is_empty() {
                None
            } else {
                Some(truncate_body(diagnostic, HEALTH_BODY_LIMIT))
            }
        };

        HealthStatus {
            backend_reachable,
            status_code,
            response_body,
            latency_ms,
        }
    }

    /// Blocking probe of `url` using `curl`.
    ///
    /// `reqwest` is deliberately not a dependency of this crate, so the
    /// probe shells out to the system `curl` (always present on macOS and
    /// shipped with Windows 10+). `Err` is reserved for the case where the
    /// probe could not run at all.
    fn probe_blocking(url: &str) -> Result<HealthStatus, String> {
        use std::process::Command;
        use std::time::Instant;

        let started = Instant::now();
        let output = Command::new("curl")
            .arg("--max-time")
            .arg(HEALTH_TIMEOUT_SECS.to_string())
            .arg("--silent")
            .arg("--show-error")
            .arg("--write-out")
            .arg("\n%{http_code}")
            .arg(url)
            .output()
            .map_err(|e| format!("failed to run curl: {e}"))?;

        let latency_ms = u64::try_from(started.elapsed().as_millis()).unwrap_or(u64::MAX);
        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);

        Ok(parse_probe_output(&stdout, &stderr, latency_ms))
    }

    /// Tauri command: probe the Go backend's `/health` endpoint.
    ///
    /// Returns reachability, HTTP status, a bounded response body, and the
    /// latency. A refused connection or timeout reports
    /// `backend_reachable: false`; `Err` is only returned when the probe
    /// itself could not run (for example `curl` is unavailable).
    #[tauri::command]
    pub async fn health_check() -> Result<HealthStatus, String> {
        let url = health_endpoint();
        // `curl` is blocking; keep it off the async runtime's worker threads.
        tauri::async_runtime::spawn_blocking(move || probe_blocking(&url))
            .await
            .map_err(|e| format!("health probe task failed: {e}"))?
    }

    #[cfg(test)]
    mod tests {
        use super::*;

        #[test]
        fn health_endpoint_default_is_local_backend() {
            // Env vars are shared across test threads; only assert shape.
            let url = health_endpoint();
            assert!(url.ends_with("/health"), "unexpected endpoint: {url}");
            assert!(url.starts_with("http"), "unexpected endpoint: {url}");
            assert!(!url.contains("//health"), "double slash in: {url}");
        }

        #[test]
        fn parse_probe_output_reports_success() {
            let status = parse_probe_output("{\"status\":\"ok\"}\n200", "", 12);
            assert!(status.backend_reachable);
            assert_eq!(status.status_code, Some(200));
            assert_eq!(status.response_body.as_deref(), Some("{\"status\":\"ok\"}"));
            assert_eq!(status.latency_ms, 12);
        }

        #[test]
        fn parse_probe_output_reports_server_error_body() {
            let status = parse_probe_output("db down\n503", "", 5);
            assert!(status.backend_reachable);
            assert_eq!(status.status_code, Some(503));
            assert_eq!(status.response_body.as_deref(), Some("db down"));
        }

        #[test]
        fn parse_probe_output_treats_curl_failure_as_unreachable() {
            let status = parse_probe_output(
                "\n000",
                "curl: (7) Failed to connect to localhost port 8080: Connection refused",
                2001,
            );
            assert!(!status.backend_reachable);
            assert_eq!(status.status_code, None);
            assert!(status
                .response_body
                .as_deref()
                .is_some_and(|body| body.contains("Connection refused")));
        }

        #[test]
        fn parse_probe_output_drops_empty_body_when_reachable() {
            let status = parse_probe_output("\n204", "", 1);
            assert!(status.backend_reachable);
            assert_eq!(status.status_code, Some(204));
            assert_eq!(status.response_body, None);
        }

        #[test]
        fn truncate_body_is_char_boundary_safe() {
            let raw = "ééééé";
            let truncated = truncate_body(raw, 5);
            assert_eq!(truncated, "éé...");
            assert_eq!(truncate_body("short", 5), "short");
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn upload_endpoint_default() {
        // We don't clear env vars in tests; just assert the function does
        // not panic and returns a non-empty string.
        let ep = upload_endpoint();
        assert!(!ep.is_empty());
    }

    #[test]
    fn upload_bucket_default() {
        let b = upload_bucket();
        assert!(!b.is_empty());
    }
}
