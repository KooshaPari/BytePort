//! Core protocol definitions for BytePort

use bytes::{Bytes, BytesMut, Buf, BufMut};
use crate::error::{BytePortError, Result};

/// Protocol version
pub const PROTOCOL_VERSION: u8 = 1;

/// Magic bytes for frame identification
pub const FRAME_MAGIC: [u8; 2] = [0x42, 0x50]; // "BP"

/// Maximum frame size (16MB)
pub const MAX_FRAME_SIZE: u32 = 16 * 1024 * 1024;

/// Frame types
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FrameType {
    Data = 0x01,
    Control = 0x02,
    Heartbeat = 0x03,
    Ack = 0x04,
    Error = 0x05,
    Compression = 0x06,
}

impl TryFrom<u8> for FrameType {
    type Error = BytePortError;
    
    fn try_from(value: u8) -> Result<Self> {
        match value {
            0x01 => Ok(FrameType::Data),
            0x02 => Ok(FrameType::Control),
            0x03 => Ok(FrameType::Heartbeat),
            0x04 => Ok(FrameType::Ack),
            0x05 => Ok(FrameType::Error),
            0x06 => Ok(FrameType::Compression),
            _ => Err(BytePortError::InvalidFrame(format!("Unknown frame type: {}", value))),
        }
    }
}

/// Protocol frame
#[derive(Debug, Clone)]
pub struct Frame {
    pub version: u8,
    pub frame_type: FrameType,
    pub payload: Bytes,
    pub checksum: u32,
}

impl Frame {
    /// Create a new data frame
    pub fn data(payload: impl Into<Bytes>) -> Self {
        let payload = payload.into();
        let checksum = crc32fast::hash(&payload);
        
        Self {
            version: PROTOCOL_VERSION,
            frame_type: FrameType::Data,
            payload,
            checksum,
        }
    }
    
    /// Create a heartbeat frame
    pub fn heartbeat() -> Self {
        Self {
            version: PROTOCOL_VERSION,
            frame_type: FrameType::Heartbeat,
            payload: Bytes::new(),
            checksum: 0,
        }
    }
    
    /// Create an acknowledgment frame
    pub fn ack(sequence: u64) -> Self {
        let payload = Bytes::from(sequence.to_be_bytes().to_vec());
        let checksum = crc32fast::hash(&payload);
        
        Self {
            version: PROTOCOL_VERSION,
            frame_type: FrameType::Ack,
            payload,
            checksum,
        }
    }
    
    /// Encode frame to bytes
    pub fn encode(&self) -> Bytes {
        let mut buf = BytesMut::with_capacity(8 + self.payload.len());
        
        // Magic bytes
        buf.extend_from_slice(&FRAME_MAGIC);
        
        // Version
        buf.put_u8(self.version);
        
        // Frame type
        buf.put_u8(self.frame_type as u8);
        
        // Payload length (4 bytes, big endian)
        buf.put_u32(self.payload.len() as u32);
        
        // Payload
        buf.extend_from_slice(&self.payload);
        
        // Checksum
        buf.put_u32(self.checksum);
        
        buf.freeze()
    }
    
    /// Decode frame from bytes
    pub fn decode(mut buf: impl Buf) -> Result<Self> {
        // Check magic bytes
        let mut magic = [0u8; 2];
        buf.copy_to_slice(&mut magic);
        if magic != FRAME_MAGIC {
            return Err(BytePortError::InvalidFrame("Invalid magic bytes".to_string()));
        }
        
        // Version
        let version = buf.get_u8();
        if version != PROTOCOL_VERSION {
            return Err(BytePortError::VersionMismatch {
                expected: PROTOCOL_VERSION,
                actual: version,
            });
        }
        
        // Frame type
        let frame_type = FrameType::try_from(buf.get_u8())?;
        
        // Payload length
        let payload_len = buf.get_u32();
        if payload_len > MAX_FRAME_SIZE {
            return Err(BytePortError::InvalidFrame(format!(
                "Payload too large: {} > {}",
                payload_len, MAX_FRAME_SIZE
            )));
        }
        
        // Payload
        let mut payload = vec![0u8; payload_len as usize];
        buf.copy_to_slice(&mut payload);
        let payload = Bytes::from(payload);
        
        // Checksum
        let checksum = buf.get_u32();
        let computed_checksum = crc32fast::hash(&payload);
        if checksum != computed_checksum {
            return Err(BytePortError::CrcMismatch);
        }
        
        Ok(Self {
            version,
            frame_type,
            payload,
            checksum,
        })
    }
}

/// Variable-length integer encoding (VarInt)
pub mod varint {
    use bytes::{Buf, BufMut};
    
    /// Encode u64 as VarInt
    pub fn encode_u64(value: u64, buf: &mut impl BufMut) {
        let mut value = value;
        while value >= 0x80 {
            buf.put_u8((value as u8) | 0x80);
            value >>= 7;
        }
        buf.put_u8(value as u8);
    }
    
    /// Decode VarInt to u64
    pub fn decode_u64(buf: &mut impl Buf) -> Option<u64> {
        let mut result = 0u64;
        let mut shift = 0;
        
        loop {
            if !buf.has_remaining() {
                return None;
            }
            
            let byte = buf.get_u8();
            result |= ((byte & 0x7f) as u64) << shift;
            
            if (byte & 0x80) == 0 {
                return Some(result);
            }
            
            shift += 7;
            if shift > 63 {
                return None; // Overflow
            }
        }
    }
    
    /// Encode u32 as VarInt
    pub fn encode_u32(value: u32, buf: &mut impl BufMut) {
        encode_u64(value as u64, buf);
    }
    
    /// Decode VarInt to u32
    pub fn decode_u32(buf: &mut impl Buf) -> Option<u32> {
        decode_u64(buf).and_then(|v| if v <= u32::MAX as u64 { Some(v as u32) } else { None })
    }
}
