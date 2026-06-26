//! OTel metrics for the BytePort CLI.
//!
//! Tracks invocation counters and error rates per command using
//! OpenTelemetry metric instruments. The exporter writes to stdout
//! for local development; a production deployment would swap in an
//! OTLP exporter.
//!
//! # Instruments
//!
//! | Instrument                 | Type      | Labels                  | Description                        |
//! |----------------------------|-----------|-------------------------|------------------------------------|
//! | `byteport.cli.invocations` | Counter   | `command`               | Total CLI invocations per command  |
//! | `byteport.cli.errors`      | Counter   | `command`, `error_type` | CLI errors per command and kind    |

use opentelemetry::{
    global,
    metrics::{Counter, Meter},
    KeyValue,
};
use opentelemetry_sdk::metrics::SdkMeterProvider;
use std::sync::OnceLock;

static METRICS: OnceLock<CliMetrics> = OnceLock::new();

/// Singleton OTel metrics handle for the CLI.
pub struct CliMetrics {
    /// Invocation counter: `byteport.cli.invocations`.
    invocations: Counter<u64>,
    /// Error counter: `byteport.cli.errors`.
    errors: Counter<u64>,
}

/// Command categories tracked by metrics.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum CommandKind {
    Codec,
    Transport,
    Ui,
    Upload,
}

impl CommandKind {
    fn as_str(&self) -> &'static str {
        match self {
            CommandKind::Codec => "codec",
            CommandKind::Transport => "transport",
            CommandKind::Ui => "ui",
            CommandKind::Upload => "upload",
        }
    }
}

/// Error classifications for metric labels.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum ErrorKind {
    /// User provided invalid input (e.g. bad hex string, unknown view).
    InvalidInput,
    /// An operation failed at the transport / codec layer.
    OperationFailed,
    /// An unexpected internal error.
    Internal,
}

impl ErrorKind {
    fn as_str(&self) -> &'static str {
        match self {
            ErrorKind::InvalidInput => "invalid_input",
            ErrorKind::OperationFailed => "operation_failed",
            ErrorKind::Internal => "internal",
        }
    }
}

/// Initialise the global OTel meter provider and return a handle to the
/// CLI metrics instruments. Safe to call multiple times — subsequent
/// calls are no-ops.
pub fn init() -> &'static CliMetrics {
    METRICS.get_or_init(|| {
        let provider = SdkMeterProvider::builder()
            .with_reader(opentelemetry_stdout::MetricsExporterBuilder::default().build())
            .build();
        global::set_meter_provider(provider.clone());

        let meter: Meter = global::meter("byteport-cli");
        let invocations: Counter<u64> = meter
            .u64_counter("byteport.cli.invocations")
            .with_description("Total CLI invocations per command")
            .with_unit("{invocation}")
            .init();
        let errors: Counter<u64> = meter
            .u64_counter("byteport.cli.errors")
            .with_description("CLI errors per command and kind")
            .with_unit("{error}")
            .init();

        CliMetrics { invocations, errors }
    })
}

impl CliMetrics {
    /// Record a successful command invocation.
    pub fn record_invocation(&self, command: CommandKind) {
        self.invocations.add(
            1,
            &[KeyValue::new("command", command.as_str())],
        );
    }

    /// Record a command error.
    pub fn record_error(&self, command: CommandKind, kind: ErrorKind) {
        self.errors.add(
            1,
            &[
                KeyValue::new("command", command.as_str()),
                KeyValue::new("error_type", kind.as_str()),
            ],
        );
    }

    /// Convenience: record invocation then return the metrics handle
    /// for optional error tracking in the caller.
    pub fn track(command: CommandKind) -> &'static Self {
        let m = init();
        m.record_invocation(command);
        m
    }
}
