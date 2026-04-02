//! Error types for BytePort protocol

use thiserror::Error;
use std::io;

#[derive(Error, Debug, Clone)]
pub enum BytePortError {
    #[error("IO error: {0}")]
    Io(String),
    
    #[error("Serialization error: {0}")]
    Serialization(String),
    
    #[error("Deserialization error: {0}")]
    Deserialization(String),
    
    #[error("Protocol error: {0}")]
    Protocol(String),
    
    #[error("Transport error: {0}")]
    Transport(String),
    
    #[error("Connection error: {0}")]
    Connection(String),
    
    #[error("Compression error: {0}")]
    Compression(String),
    
    #[error("Encryption error: {0}")]
    Encryption(String),
    
    #[error("Timeout")]
    Timeout,
    
    #[error("Invalid frame: {0}")]
    InvalidFrame(String),
    
    #[error("Version mismatch: expected {expected}, got {actual}")]
    VersionMismatch { expected: u8, actual: u8 },
    
    #[error("CRC checksum failed")]
    CrcMismatch,
}

impl From<io::Error> for BytePortError {
    fn from(e: io::Error) -> Self {
        BytePortError::Io(e.to_string())
    }
}

pub type Result<T> = std::result::Result<T, BytePortError>;
