//! # byteport-cli — WASM-compatible CLI subset
//!
//! A subset of the BytePort CLI that compiles to `wasm32-wasi`.
//! Supports transport upload creation commands.
//!
//! ## Commands
//!
//! | Command                  | Description                              |
//! |--------------------------|------------------------------------------|
//! | `transport create-upload`| Create an S3 upload instruction          |

use clap::{Parser, Subcommand};

use byteport_transport::{S3UploadTransport, UploadRequest, UploadTransport};

/// BytePort WASM CLI — compute/infra automation subset.
#[derive(Debug, Parser)]
#[command(
    name = "byteport",
    version,
    about = "BytePort WASM CLI (compute/infra subset)",
    long_about = None
)]
struct Cli {
    #[command(subcommand)]
    command: Command,
}

#[derive(Debug, Subcommand)]
enum Command {
    /// Transport operations (S3 upload instructions).
    Transport {
        #[command(subcommand)]
        action: TransportAction,
    },
}

#[derive(Debug, Subcommand)]
enum TransportAction {
    /// Create an S3 pre-signed upload instruction.
    CreateUpload {
        /// S3 endpoint URL.
        #[arg(long, default_value = "https://uploads.byteport.local")]
        endpoint: String,
        /// S3 bucket name.
        #[arg(long, default_value = "byteport-uploads")]
        bucket: String,
        /// Object key to upload to.
        #[arg(long)]
        object_key: String,
        /// MIME content type.
        #[arg(long, default_value = "application/octet-stream")]
        content_type: String,
        /// Content length in bytes.
        #[arg(long, default_value_t = 0)]
        content_length: u64,
    },
}

fn main() {
    let cli = Cli::parse();

    match cli.command {
        Command::Transport { action } => match action {
            TransportAction::CreateUpload {
                endpoint,
                bucket,
                object_key,
                content_type,
                content_length,
            } => cmd_transport_create_upload(
                &endpoint,
                &bucket,
                &object_key,
                &content_type,
                content_length,
            ),
        },
    }
}

fn cmd_transport_create_upload(
    endpoint: &str,
    bucket: &str,
    object_key: &str,
    content_type: &str,
    content_length: u64,
) {
    let transport = S3UploadTransport::new(endpoint, bucket, None::<&str>);

    match transport.create_upload(&UploadRequest {
        object_key: object_key.to_string(),
        content_type: content_type.to_string(),
        content_length,
    }) {
        Ok(instruction) => {
            let json = serde_json::to_string_pretty(&instruction)
                .unwrap_or_else(|e| format!("{{\"error\": \"{e}\"}}"));
            println!("{json}");
        }
        Err(e) => {
            eprintln!("Error: {e}");
            std::process::exit(1);
        }
    }
}
