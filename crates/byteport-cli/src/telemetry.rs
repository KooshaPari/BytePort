//! Telemetry initialization (feature-gated).
//!
//! When the `otel` feature is enabled (the default), this module wires up
//! the [`byteport_otel`] stack. When disabled, it falls back to a plain
//! `tracing-subscriber` with an env filter.

use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

use byteport_otel::config::TelemetryConfig;

/// Initialize tracing + (optionally) OTel based on the runtime config.
///
/// Returns whether telemetry was set up successfully. Failures are
/// non-fatal — the CLI keeps running with a console-only subscriber.
pub fn init(log_level: &str) -> bool {
    // Honor RUST_LOG first; fall back to the CLI flag.
    let filter = EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(log_level));

    let base = tracing_subscriber::registry().with(filter);

    #[cfg(feature = "otel")]
    {
        let mut cfg = TelemetryConfig::default();
        cfg.log_level = log_level.to_string();
        cfg.enable_stdout_log = true;
        let _guard = byteport_otel::init::init_telemetry(cfg);
        // The OTel init already set up the global subscriber; bail.
        let _ = base;
        return true;
    }

    #[cfg(not(feature = "otel"))]
    {
        let _ = base
            .with(
                tracing_subscriber::fmt::layer()
                    .with_target(true)
                    .with_thread_ids(false),
            )
            .try_init();
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn init_does_not_panic() {
        assert!(init("info"));
    }
}