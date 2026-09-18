//! Criterion benchmarks for the desktop IPC bridge payloads.
//!
//! These measure the `serde_json` cost of the payload shapes that actually
//! cross the Tauri bridge: [`CreateUploadArgs`] is deserialized from the
//! webview on every upload request, [`CreateUploadResponse`] is serialized
//! back to it, and [`HealthStatus`] is re-serialized on every backend probe
//! the UI polls.
//!
//! The previous revision of this file benchmarked `app_lib::ipc::IpcEnvelope`,
//! a type that no longer exists (the standalone `ipc.rs` module that defined
//! it was deleted as dead code and replaced by the inline `pub mod ipc` in
//! `src/lib.rs`). The benchmark now targets that live surface, so the
//! `.github/workflows/bench.yml` job measures real IPC serialization cost
//! instead of failing to compile.
//!
//! Run with `cargo bench -p app --bench ipc`.

use std::collections::BTreeMap;
use std::hint::black_box;

use app_lib::ipc::{CreateUploadArgs, CreateUploadResponse, HealthStatus};
use criterion::{criterion_group, criterion_main, Criterion};

/// A representative `create_upload` argument payload as sent by the webview.
fn upload_args_json() -> String {
    serde_json::json!({
        "object_key": "/projects/demo/screenshot.png",
        "content_type": "image/png",
        "content_length": 4_194_304_u64,
    })
    .to_string()
}

/// A representative pre-signed upload instruction as returned to the webview.
fn upload_response() -> CreateUploadResponse {
    let mut headers = BTreeMap::new();
    headers.insert("content-type".to_string(), "image/png".to_string());
    headers.insert("content-length".to_string(), "4194304".to_string());
    headers.insert("x-amz-acl".to_string(), "private".to_string());
    CreateUploadResponse {
        method: "PUT".to_string(),
        url: "https://uploads.byteport.local/byteport-uploads/projects/demo/screenshot.png\
              ?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=demo&X-Amz-Expires=900"
            .to_string(),
        headers,
    }
}

/// A representative successful `/health` probe result as reported to the UI.
fn health_status() -> HealthStatus {
    HealthStatus {
        backend_reachable: true,
        status_code: Some(200),
        response_body: Some("{\"status\":\"ok\",\"db\":\"up\"}".to_string()),
        latency_ms: 3,
    }
}

/// Benchmarks the serde round-trip cost of each live IPC payload shape.
fn bench_ipc_payloads(c: &mut Criterion) {
    let args_json = upload_args_json();
    let response = upload_response();
    let health = health_status();

    c.bench_function("ipc_deserialize_create_upload_args", |b| {
        b.iter(|| {
            serde_json::from_str::<CreateUploadArgs>(black_box(&args_json)).expect("deserialize CreateUploadArgs")
        })
    });

    c.bench_function("ipc_serialize_create_upload_response", |b| {
        b.iter(|| serde_json::to_string(black_box(&response)).expect("serialize CreateUploadResponse"))
    });

    c.bench_function("ipc_serialize_health_status", |b| {
        b.iter(|| serde_json::to_string(black_box(&health)).expect("serialize HealthStatus"))
    });
}

criterion_group!(ipc_benches, bench_ipc_payloads);
criterion_main!(ipc_benches);
