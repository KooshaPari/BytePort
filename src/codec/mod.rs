//! Codec implementations for BytePort protocol

use bytes::{Buf, BytesMut, BufMut};  // Add Buf trait
use tokio_util::codec::{Encoder, Decoder};
use crate::protocol::Frame;
use crate::error::{BytePortError, Result};

/// Frame codec for BytePort protocol
pub struct FrameCodec;

impl Decoder for FrameCodec {
    type Item = Frame;
    type Error = BytePortError;
    
    fn decode(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>> {
        if src.len() < 8 {
            // Need at least header
            return Ok(None);
        }
        
        // Check magic bytes
        if &src[0..2] != &[0x42, 0x50] {
            // Discard until we find magic
            src.advance(1);
            return Ok(None);
        }
        
        // Get payload length
        let payload_len = u32::from_be_bytes([src[4], src[5], src[6], src[7]]) as usize;
        let total_len = 8 + payload_len + 4; // header + payload + checksum
        
        if src.len() < total_len {
            // Need more data
            src.reserve(total_len - src.len());
            return Ok(None);
        }
        
        // We have a complete frame
        let frame_data = src.split_to(total_len);
        Frame::decode(&frame_data[..]).map(Some)
    }
}

impl Encoder<Frame> for FrameCodec {
    type Error = BytePortError;
    
    fn encode(&mut self, item: Frame, dst: &mut BytesMut) -> Result<()> {
        let encoded = item.encode();
        dst.reserve(encoded.len());
        dst.extend_from_slice(&encoded);
        Ok(())
    }
}
