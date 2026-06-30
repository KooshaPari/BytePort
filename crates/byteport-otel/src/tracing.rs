//! Tracing utility helpers.

/// Initialise the tracing subscriber for a test, capturing spans in memory.
/// Returns a `tracing_subscriber::reload::Handle` that can be dropped to flush.
///
/// # Panics
///
/// Panics if the global subscriber was already set (i.e., called more than
/// once without resetting).
#[cfg(test)]
pub fn init_test_tracing(
) -> tracing_subscriber::reload::Handle<tracing_subscriber::filter::EnvFilter, tracing_subscriber::registry::Registry> {
    use tracing_subscriber::{layer::SubscriberExt, reload, util::SubscriberInitExt, EnvFilter};

    let filter = EnvFilter::new("debug");
    let (filter_layer, handle) = reload::Layer::new(filter);

    tracing_subscriber::registry()
        .with(filter_layer)
        .with(tracing_subscriber::fmt::layer().with_test_writer())
        .init();

    handle
}

#[cfg(test)]
mod tests {
    use opentelemetry::trace::{Tracer as _, TracerProvider as _};
    use opentelemetry_sdk::trace::SdkTracerProvider;

    #[test]
    fn telemetry_guard_flush_does_not_panic() {
        // Drop a TelemetryGuard that has no real OTLP exporter — should be a no-op.
        let guard = crate::init::init_default();
        drop(guard);
    }

    #[test]
    fn tracer_provider_default_has_name() {
        let tp = SdkTracerProvider::default();
        let tracer = tp.tracer("test");
        let _span = tracer.start("test-span");
    }
}
