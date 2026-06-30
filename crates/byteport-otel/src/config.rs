//! Telemetry configuration types.

/// Configuration for the BytePort OpenTelemetry stack.
#[derive(Debug, Clone)]
pub struct TelemetryConfig {
    /// Service name reported to the observability backend.
    pub service_name: String,
    /// Service version (e.g. crate version).
    pub service_version: String,
    /// OTLP gRPC endpoint.
    pub otlp_endpoint: String,
    /// Whether to export metrics.
    pub enable_metrics: bool,
    /// Whether to export traces.
    pub enable_tracing: bool,
    /// Whether to log structured JSON to stdout as well.
    pub enable_stdout_log: bool,
    /// Log level filter (e.g. "info", "debug").
    pub log_level: String,
}

impl Default for TelemetryConfig {
    fn default() -> Self {
        Self {
            service_name: "byteport".into(),
            service_version: env!("CARGO_PKG_VERSION").into(),
            otlp_endpoint: "http://localhost:4317".into(),
            enable_metrics: true,
            enable_tracing: true,
            enable_stdout_log: true,
            log_level: "info".into(),
        }
    }
}

impl TelemetryConfig {
    /// Builder-style constructor.
    pub fn builder() -> TelemetryConfigBuilder {
        TelemetryConfigBuilder::default()
    }
}

/// Builder for [`TelemetryConfig`].
#[derive(Default)]
pub struct TelemetryConfigBuilder {
    service_name: Option<String>,
    service_version: Option<String>,
    otlp_endpoint: Option<String>,
    enable_metrics: Option<bool>,
    enable_tracing: Option<bool>,
    enable_stdout_log: Option<bool>,
    log_level: Option<String>,
}

impl TelemetryConfigBuilder {
    /// Set the service name.
    pub fn service_name(mut self, v: impl Into<String>) -> Self {
        self.service_name = Some(v.into());
        self
    }
    /// Set the service version.
    pub fn service_version(mut self, v: impl Into<String>) -> Self {
        self.service_version = Some(v.into());
        self
    }
    /// Set the OTLP endpoint.
    pub fn otlp_endpoint(mut self, v: impl Into<String>) -> Self {
        self.otlp_endpoint = Some(v.into());
        self
    }
    /// Enable or disable metrics.
    pub fn enable_metrics(mut self, v: bool) -> Self {
        self.enable_metrics = Some(v);
        self
    }
    /// Enable or disable tracing.
    pub fn enable_tracing(mut self, v: bool) -> Self {
        self.enable_tracing = Some(v);
        self
    }
    /// Enable or disable stdout structured logging.
    pub fn enable_stdout_log(mut self, v: bool) -> Self {
        self.enable_stdout_log = Some(v);
        self
    }
    /// Set the log level filter.
    pub fn log_level(mut self, v: impl Into<String>) -> Self {
        self.log_level = Some(v.into());
        self
    }
    /// Build the config.
    pub fn build(self) -> TelemetryConfig {
        let base = TelemetryConfig::default();
        TelemetryConfig {
            service_name: self.service_name.unwrap_or(base.service_name),
            service_version: self.service_version.unwrap_or(base.service_version),
            otlp_endpoint: self.otlp_endpoint.unwrap_or(base.otlp_endpoint),
            enable_metrics: self.enable_metrics.unwrap_or(base.enable_metrics),
            enable_tracing: self.enable_tracing.unwrap_or(base.enable_tracing),
            enable_stdout_log: self.enable_stdout_log.unwrap_or(base.enable_stdout_log),
            log_level: self.log_level.unwrap_or(base.log_level),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn config_uses_defaults() {
        let cfg = TelemetryConfig::default();
        assert_eq!(cfg.service_name, "byteport");
        assert_eq!(cfg.otlp_endpoint, "http://localhost:4317");
        assert!(cfg.enable_metrics);
        assert!(cfg.enable_tracing);
    }

    #[test]
    fn builder_overrides() {
        let cfg = TelemetryConfig::builder()
            .service_name("bp-test")
            .otlp_endpoint("http://otel:4317")
            .enable_metrics(false)
            .build();
        assert_eq!(cfg.service_name, "bp-test");
        assert_eq!(cfg.otlp_endpoint, "http://otel:4317");
        assert!(!cfg.enable_metrics);
        assert!(cfg.enable_tracing);
    }
}
