//! Metric instruments used by BytePort components.

use std::sync::OnceLock;

use opentelemetry::{metrics::Meter, KeyValue};
use opentelemetry_sdk::metrics::MetricResult;

static METER: OnceLock<Meter> = OnceLock::new();

fn meter() -> &'static Meter {
    METER.get_or_init(|| opentelemetry::global::meter("byteport"))
}

/// Initialise globally-scoped metric instruments. Safe to call multiple times.
pub fn init_metrics() -> MetricResult<()> {
    let m = meter();

    // ── HTTP / API metrics ───────────────────────────────────
    let _http_requests = m
        .u64_counter("http.requests.total")
        .with_description("Total number of HTTP requests received")
        .with_unit("1")
        .build();

    // For now, we just ensure instruments are constructed without error.
    // Explicitly ignore unused variables to avoid clippy warnings.
    std::mem::drop(_http_requests);

    Ok(())
}

/// Record a counter increment.
pub fn record_request(path: &str, status: u16) {
    let m = meter();
    let counter = m
        .u64_counter("http.requests.count")
        .with_description("Count of HTTP requests per route")
        .with_unit("1")
        .build();

    counter.add(
        1,
        &[
            KeyValue::new("http.route", path.to_string()),
            KeyValue::new("http.status_code", status as i64),
        ],
    );
}

/// Record a CLI command invocation.
pub fn record_cli_invocation(command: &str) {
    let m = meter();
    let counter = m
        .u64_counter("cli.invocations")
        .with_description("Number of CLI command invocations")
        .with_unit("1")
        .build();
    counter.add(1, &[KeyValue::new("cli.command", command.to_owned())]);
}

/// Record a CLI command error.
pub fn record_cli_error(command: &str, error_kind: &str) {
    let m = meter();
    let counter = m
        .u64_counter("cli.errors")
        .with_description("Number of CLI command errors")
        .with_unit("1")
        .build();
    counter.add(
        1,
        &[
            KeyValue::new("cli.command", command.to_owned()),
            KeyValue::new("error.kind", error_kind.to_owned()),
        ],
    );
}
