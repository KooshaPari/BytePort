//! Telemetry initialisation: sets up the global `TracerProvider`, `MeterProvider`,
//! and integrates with `tracing-subscriber` for structured logging.
//!
//! ## Shutdown guard
//!
//! [`TelemetryGuard`] flushes all pending spans and metric exports when dropped.
//! It should be held for the lifetime of the application.

use std::time::Duration;

use opentelemetry::trace::{TraceError, TracerProvider};
use opentelemetry::KeyValue;
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::{
    metrics::{MeterProviderBuilder, MetricResult},
    trace::SdkTracerProvider,
    Resource,
};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

use crate::config::TelemetryConfig;

/// A guard that flushes and shuts down the telemetry pipeline on drop.
pub struct TelemetryGuard {
    _tracer_provider: Option<SdkTracerProvider>,
}

impl Drop for TelemetryGuard {
    fn drop(&mut self) {
        if let Some(tp) = self._tracer_provider.take() {
            let _ = tp.shutdown();
        }
    }
}

/// Initialise the full BytePort telemetry stack.
///
/// Returns a [`TelemetryGuard`] that must be kept alive for the application's
/// lifetime. Dropping it triggers a graceful flush of all pending telemetry.
pub fn init_telemetry(config: TelemetryConfig) -> TelemetryGuard {
    let resource = Resource::builder()
        .with_attributes(vec![
            KeyValue::new("service.name", config.service_name.clone()),
            KeyValue::new("service.version", config.service_version.clone()),
            #[cfg(feature = "semconv")]
            KeyValue::new(
                opentelemetry_semantic_conventions::resource::SERVICE_NAME,
                config.service_name.clone(),
            ),
        ])
        .build();

    // ── Trace provider ────────────────────────────────────────────────
    let tracer_provider = if config.enable_tracing {
        match build_tracer_provider(&config, resource.clone()) {
            Ok(tp) => {
                let _ = opentelemetry::global::set_tracer_provider(tp.clone());
                Some(tp)
            }
            Err(e) => {
                // If OTLP init fails, fall back to stdout trace.
                eprintln!("byteport-otel: OTLP tracer init failed ({e}), falling back to stdout");
                None
            }
        }
    } else {
        None
    };

    // ── Metric provider ────────────────────────────────────────────────
    if config.enable_metrics {
        match build_meter_provider(&config, resource.clone()) {
            Ok(mp) => {
                opentelemetry::global::set_meter_provider(mp);
            }
            Err(e) => {
                eprintln!("byteport-otel: OTLP meter init failed ({e}), metrics disabled");
            }
        }
    }

    // ── Tracing subscriber ────────────────────────────────────────────
    if config.enable_tracing {
        let filter = EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(&config.log_level));

        let subscriber = tracing_subscriber::registry().with(filter);

        match (config.enable_stdout_log, tracer_provider.as_ref()) {
            (true, Some(tp)) => {
                let tracer = tp.tracer("byteport");
                subscriber
                    .with(
                        tracing_subscriber::fmt::layer()
                            .json()
                            .with_target(true)
                            .with_thread_ids(true),
                    )
                    .with(tracing_opentelemetry::layer().with_tracer(tracer))
                    .init();
            }
            (true, None) => {
                subscriber
                    .with(
                        tracing_subscriber::fmt::layer()
                            .json()
                            .with_target(true)
                            .with_thread_ids(true),
                    )
                    .init();
            }
            (false, Some(tp)) => {
                let tracer = tp.tracer("byteport");
                subscriber
                    .with(tracing_opentelemetry::layer().with_tracer(tracer))
                    .init();
            }
            (false, None) => {
                subscriber.init();
            }
        }
    } else if config.enable_stdout_log {
        let filter = EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(&config.log_level));

        tracing_subscriber::registry()
            .with(filter)
            .with(
                tracing_subscriber::fmt::layer()
                    .json()
                    .with_target(true)
                    .with_thread_ids(true),
            )
            .init();
    }

    TelemetryGuard {
        _tracer_provider: tracer_provider,
    }
}

/// Initialise telemetry with default configuration.
pub fn init_default() -> TelemetryGuard {
    init_telemetry(TelemetryConfig::default())
}

// ── Internal helpers ─────────────────────────────────────────────────

fn build_tracer_provider(config: &TelemetryConfig, resource: Resource) -> Result<SdkTracerProvider, TraceError> {
    let exporter = opentelemetry_otlp::SpanExporter::builder()
        .with_tonic()
        .with_endpoint(&config.otlp_endpoint)
        .with_timeout(Duration::from_secs(10))
        .build()?;

    Ok(SdkTracerProvider::builder()
        .with_batch_exporter(exporter)
        .with_resource(resource)
        .build())
}

fn build_meter_provider(
    config: &TelemetryConfig,
    resource: Resource,
) -> MetricResult<opentelemetry_sdk::metrics::SdkMeterProvider> {
    let exporter = opentelemetry_otlp::MetricExporter::builder()
        .with_tonic()
        .with_endpoint(&config.otlp_endpoint)
        .with_timeout(Duration::from_secs(10))
        .build()?;

    Ok(MeterProviderBuilder::default()
        .with_reader(
            opentelemetry_sdk::metrics::PeriodicReader::builder(exporter)
                .with_interval(Duration::from_secs(60))
                .build(),
        )
        .with_resource(resource)
        .build())
}
