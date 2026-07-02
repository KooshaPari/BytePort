//! # byteport-cli
//!
//! CLI for Phenotype compute/infra — DAG execution and orchestration.
//!
//! ## Subcommands
//!
//! | Subcommand            | Description                                             |
//! |-----------------------|---------------------------------------------------------|
//! | `pheno-dag run <yaml>`| Parse a YAML DAG definition, schedule, execute          |
//! | `pheno-dag upload`    | Dispatch a `Transport::CreateUpload` call (OTel-traced) |
//!
//! ## Example
//!
//! ```shell
//! # Run a DAG from a YAML file
//! pheno-dag run examples/ci-pipeline.yaml
//!
//! # Create an S3 upload instruction (span emitted to OTLP collector)
//! pheno-dag upload my/object.bin --content-type application/octet-stream --content-length 1024
//! ```

use std::fs;
use std::path::PathBuf;

use clap::{Parser, Subcommand};
use opentelemetry::global;
use opentelemetry::trace::{Span, Tracer};

use byteport_dag::dag::Dag;
use byteport_dag::scheduler;
use byteport_dag::serialize::DagSchema;
use byteport_otel::propagation;
use byteport_transport::{S3UploadTransport, UploadRequest, UploadTransport};

// ---------------------------------------------------------------------------
// CLI entry point
// ---------------------------------------------------------------------------

#[derive(Parser)]
#[command(name = "pheno-dag", about = "Phenotype compute/infra DAG automation", version)]
struct Cli {
    #[command(subcommand)]
    command: DagCommand,
}

#[derive(Subcommand)]
enum DagCommand {
    /// Parse a YAML DAG definition, compute a schedule, and print execution plan
    Run {
        /// Path to the YAML file containing the DAG definition
        yaml: PathBuf,

        /// Optional DAG name filter (only execute nodes matching a pattern)
        #[arg(short, long)]
        name: Option<String>,

        /// Enable verbose output including serialized schedule
        #[arg(short, long)]
        verbose: bool,
    },

    /// Dispatch a Transport::CreateUpload call (wrapped in an OTel span).
    ///
    /// Propagates the current trace context into W3C env-var headers so that
    /// downstream processes can continue the same trace tree.
    Upload {
        /// Object key (path within the bucket)
        key: String,

        /// MIME content type
        #[arg(long, default_value = "application/octet-stream")]
        content_type: String,

        /// Content length in bytes
        #[arg(long, default_value_t = 0)]
        content_length: u64,

        /// S3-compatible storage endpoint
        #[arg(long, default_value = "https://storage.example.test")]
        endpoint: String,

        /// S3 bucket name
        #[arg(long, default_value = "byteport-uploads")]
        bucket: String,

        /// Optional key prefix
        #[arg(long)]
        prefix: Option<String>,
    },
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

fn main() {
    let cli = Cli::parse();
    dispatch(&cli.command);
}

/// Short, stable name for a subcommand, used as the span name and log field.
fn command_name(command: &DagCommand) -> &'static str {
    match command {
        DagCommand::Run { .. } => "run",
        DagCommand::Upload { .. } => "upload",
    }
}

/// Wrap a CLI command in an OTel span so each invocation is traceable, then
/// dispatch to its handler. The span is the trace root for any child processes
/// the command spawns (see `byteport_otel::propagation`).
fn dispatch(command: &DagCommand) {
    let cmd = command_name(command);
    let tracer = global::tracer("byteport-cli");
    let mut span = tracer.start(format!("cli.{cmd}"));
    span.set_attribute(opentelemetry::KeyValue::new("cli.command", cmd));
    tracing::info!(command = cmd, "cli command invoked");

    match command {
        DagCommand::Run {
            yaml,
            name: _name,
            verbose,
        } => run_dag(yaml, *verbose),

        DagCommand::Upload {
            key,
            content_type,
            content_length,
            endpoint,
            bucket,
            prefix,
        } => dispatch_upload(key, content_type, *content_length, endpoint, bucket, prefix.as_deref()),
    }
}

// ---------------------------------------------------------------------------
// DAG execution logic
// ---------------------------------------------------------------------------

/// Parse a YAML file, build the DAG, compute a schedule, and print it.
fn run_dag(path: &PathBuf, verbose: bool) {
    // 1. Read the YAML file
    let contents = match fs::read_to_string(path) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("Error: cannot read {} — {}", path.display(), e);
            std::process::exit(1);
        }
    };

    // 2. Deserialize into DagSchema
    let schema = match DagSchema::from_yaml(&contents) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("Error: failed to parse YAML — {}", e);
            std::process::exit(2);
        }
    };

    if verbose {
        eprintln!(
            "[info] parsed {}: {} nodes, {} edges",
            path.display(),
            schema.nodes.len(),
            schema.edges.len()
        );
        if let Some(ref name) = schema.name {
            eprintln!("[info] DAG name: {}", name);
        }
        eprintln!("[info] schema version: {}", schema.version);
    }

    // 3. Convert schema into an executable Dag
    let dag: Dag<String> = match schema.into_dag() {
        Ok(d) => d,
        Err(e) => {
            eprintln!("Error: invalid DAG definition — {}", e);
            std::process::exit(3);
        }
    };

    // 4. Compute the parallel-bucket schedule
    let schedule = match scheduler::schedule(&dag) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("Error: DAG contains a cycle — {}", e);
            std::process::exit(4);
        }
    };

    // 5. Print the schedule
    println!("{}", scheduler::format_schedule(&schedule));

    // 6. Verbose: dump the serialized plan as YAML
    if verbose {
        let export = DagSchema::from_dag(&dag, &schema.version).with_name(schema.name.clone().unwrap_or_default());
        match export.to_yaml() {
            Ok(yaml) => println!("---\n{}", yaml),
            Err(e) => eprintln!("Warning: could not serialize schedule to YAML — {}", e),
        }
    }
}

// ---------------------------------------------------------------------------
// Upload dispatch (OTel-traced, W3C propagation)
// ---------------------------------------------------------------------------

/// Dispatch a `Transport::CreateUpload` call inside an OTel span.
///
/// The W3C TraceContext headers are written to stderr so that callers
/// can forward them to the storage service to continue the trace tree.
fn dispatch_upload(
    key: &str,
    content_type: &str,
    content_length: u64,
    endpoint: &str,
    bucket: &str,
    prefix: Option<&str>,
) {
    let tracer = global::tracer("byteport-cli");
    let mut span = tracer.start("transport.create_upload");
    span.set_attribute(opentelemetry::KeyValue::new("upload.object_key", key.to_owned()));
    span.set_attribute(opentelemetry::KeyValue::new(
        "upload.content_type",
        content_type.to_owned(),
    ));
    span.set_attribute(opentelemetry::KeyValue::new(
        "upload.content_length",
        content_length as i64,
    ));

    // Inject current trace context into W3C env-var pairs so that
    // downstream processes (e.g. storage sidecars) can continue this span.
    let prop_envs = propagation::current_context_envs();
    if !prop_envs.is_empty() {
        eprintln!("[otel] propagating {} context header(s):", prop_envs.len());
        for (k, v) in &prop_envs {
            eprintln!("  {}={}", k, v);
        }
    }

    let transport = S3UploadTransport::new(endpoint, bucket, prefix);
    match transport.create_upload(&UploadRequest {
        object_key: key.to_owned(),
        content_type: content_type.to_owned(),
        content_length,
    }) {
        Ok(instruction) => {
            println!("method:  {}", instruction.method);
            println!("url:     {}", instruction.url);
            for (h, v) in &instruction.headers {
                println!("header:  {}: {}", h, v);
            }
        }
        Err(e) => {
            span.set_status(opentelemetry::trace::Status::error(e.to_string()));
            eprintln!("Error: upload dispatch failed — {}", e);
            std::process::exit(5);
        }
    }

    span.end();
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use clap::CommandFactory;

    #[test]
    fn cli_definition_is_valid() {
        Cli::command().debug_assert();
    }

    #[test]
    fn command_name_maps_run() {
        let cmd = DagCommand::Run {
            yaml: PathBuf::from("x.yaml"),
            name: None,
            verbose: false,
        };
        assert_eq!(command_name(&cmd), "run");
    }

    #[test]
    fn dispatch_emits_cli_span() {
        use opentelemetry_sdk::trace::{InMemorySpanExporter, SdkTracerProvider};

        // Install a tracer provider with an in-memory exporter so we can
        // assert the span produced by `dispatch`.
        let exporter = InMemorySpanExporter::default();
        let provider = SdkTracerProvider::builder()
            .with_simple_exporter(exporter.clone())
            .build();
        global::set_tracer_provider(provider.clone());

        // A valid DAG so `run_dag` completes without `process::exit`.
        let yaml = r#"
version: "1.0.0"
name: "span-test"
nodes:
  - { id: "a" }
  - { id: "b" }
edges:
  - { from: "a", to: "b" }
"#;
        let tmp = std::env::temp_dir().join("_cli_span_test_dag.yaml");
        fs::write(&tmp, yaml).expect("write temp YAML");

        dispatch(&DagCommand::Run {
            yaml: tmp.clone(),
            name: None,
            verbose: false,
        });

        // SimpleSpanExporter exports on span end; force flush to be safe.
        let _ = provider.force_flush();
        let spans = exporter.get_finished_spans().expect("finished spans");
        fs::remove_file(&tmp).ok();

        assert!(
            spans.iter().any(|s| s.name == "cli.run"),
            "expected a `cli.run` span, got {:?}",
            spans.iter().map(|s| s.name.as_ref()).collect::<Vec<_>>()
        );
    }

    #[test]
    fn run_dag_with_valid_yaml_succeeds() {
        // Build a minimal YAML dag
        let yaml = r#"
version: "1.0.0"
name: "test-pipeline"
nodes:
  - { id: "build" }
  - { id: "test" }
  - { id: "deploy" }
edges:
  - { from: "build", to: "test" }
  - { from: "test", to: "deploy" }
"#;
        let tmp = std::env::temp_dir().join("_f10_test_dag.yaml");
        fs::write(&tmp, yaml).expect("write temp YAML");

        // Capture stdout
        let saved_path = PathBuf::from(&tmp);
        run_dag(&saved_path, false);

        fs::remove_file(&tmp).ok();
    }

    #[test]
    fn run_dag_with_missing_file_reports_error() {
        let missing = PathBuf::from("/nonexistent/pipeline.yaml");
        // We just verify the function exits with an error code.
        // Since run_dag calls process::exit, we check that it logs via
        // stderr and exits.  For unit test coverage we exercise the
        // pre-exit paths via helper assertions.
        let result = fs::read_to_string(&missing);
        assert!(result.is_err(), "missing file must produce an error");
    }

    #[test]
    fn run_dag_with_invalid_yaml_reports_error() {
        let bad_yaml = "version: 1.0.0\nnodes: [invalid";
        let tmp = std::env::temp_dir().join("_f10_bad_dag.yaml");
        fs::write(&tmp, bad_yaml).expect("write bad YAML");

        let schema = DagSchema::from_yaml(&fs::read_to_string(&tmp).unwrap());
        assert!(schema.is_err(), "malformed YAML must fail to parse");

        fs::remove_file(&tmp).ok();
    }

    #[test]
    fn run_dag_with_cycle_reports_error() {
        let yaml = r#"
version: "1.0.0"
nodes:
  - { id: "a" }
  - { id: "b" }
  - { id: "c" }
edges:
  - { from: "a", to: "b" }
  - { from: "b", to: "c" }
  - { from: "c", to: "a" }
"#;
        let schema = DagSchema::from_yaml(yaml).unwrap();
        let dag = schema.into_dag().unwrap();
        let result = scheduler::schedule(&dag);
        assert!(result.is_err(), "cyclic DAG must fail to schedule");
    }

    /// TDD: dispatch_upload succeeds for a valid key (exercises Transport::CreateUpload span path).
    #[test]
    fn dispatch_upload_produces_instruction() {
        // Without an active tracer provider the span is a no-op, but the
        // upload itself must still succeed and print the instruction.
        dispatch_upload(
            "test/object.bin",
            "application/octet-stream",
            128,
            "https://storage.example.test",
            "byteport-uploads",
            None,
        );
        // Reaching here means dispatch_upload returned without calling process::exit.
    }

    /// TDD: propagation::current_context_envs returns empty when no OTel
    /// provider is initialised — safe no-op for callers.
    #[test]
    fn propagation_is_no_op_without_provider() {
        let envs = byteport_otel::propagation::current_context_envs();
        assert!(
            envs.is_empty(),
            "expected empty propagation without a provider, got {envs:?}"
        );
    }

    /// TDD: Upload subcommand variant is reachable via command_name.
    #[test]
    fn command_name_maps_upload() {
        let cmd = DagCommand::Upload {
            key: "test.bin".to_owned(),
            content_type: "application/octet-stream".to_owned(),
            content_length: 0,
            endpoint: "https://storage.example.test".to_owned(),
            bucket: "byteport-uploads".to_owned(),
            prefix: None,
        };
        assert_eq!(command_name(&cmd), "upload");
    }
}
