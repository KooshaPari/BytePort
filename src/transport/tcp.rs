//! TCP transport implementation for BytePort

use bytes::{Bytes, BytesMut, Buf};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use std::net::SocketAddr;

use crate::error::{BytePortError, Result};
use crate::protocol::{Frame, MAX_FRAME_SIZE};

/// TCP server for BytePort
pub struct TcpServer {
    listener: TcpListener,
    local_addr: SocketAddr,
}

impl TcpServer {
    /// Bind to a socket address
    pub async fn bind(addr: impl Into<SocketAddr>) -> Result<Self> {
        let addr = addr.into();
        let listener = TcpListener::bind(&addr).await?;
        let local_addr = listener.local_addr()?;
        
        Ok(Self { listener, local_addr })
    }
    
    /// Get local address
    pub fn local_addr(&self) -> SocketAddr {
        self.local_addr
    }
    
    /// Accept a connection
    pub async fn accept(&self) -> Result<TcpConnection> {
        let (stream, addr) = self.listener.accept().await?;
        Ok(TcpConnection { stream, addr, buffer: BytesMut::with_capacity(4096) })
    }
}

/// TCP connection
pub struct TcpConnection {
    stream: TcpStream,
    addr: SocketAddr,
    buffer: BytesMut,
}

impl TcpConnection {
    /// Connect to a remote address
    pub async fn connect(addr: impl Into<SocketAddr>) -> Result<Self> {
        let addr = addr.into();
        let stream = TcpStream::connect(&addr).await?;
        
        Ok(Self { stream, addr, buffer: BytesMut::with_capacity(4096) })
    }
    
    /// Get peer address
    pub fn peer_addr(&self) -> SocketAddr {
        self.addr
    }
    
    /// Send a frame
    pub async fn send_frame(&mut self, frame: &Frame) -> Result<()> {
        let encoded = frame.encode();
        self.stream.write_all(&encoded).await?;
        self.stream.flush().await?;
        Ok(())
    }
    
    /// Receive a frame
    pub async fn recv_frame(&mut self) -> Result<Frame> {
        loop {
            // Check if we have a complete frame
            if let Some(frame) = try_parse_frame(&mut self.buffer)? {
                return Ok(frame);
            }
            
            // Read more data
            let mut temp_buf = vec![0u8; 4096];
            let n = self.stream.read(&mut temp_buf).await?;
            
            if n == 0 {
                return Err(BytePortError::Connection("Connection closed".into()));
            }
            
            self.buffer.extend_from_slice(&temp_buf[..n]);
            
            // Prevent buffer overflow
            if self.buffer.len() > MAX_FRAME_SIZE as usize * 2 {
                return Err(BytePortError::InvalidFrame("Buffer overflow".into()));
            }
        }
    }
    
    /// Close the connection
    pub async fn close(mut self) -> Result<()> {
        self.stream.shutdown().await?;
        Ok(())
    }
}

/// Try to parse a frame from the buffer
fn try_parse_frame(buf: &mut BytesMut) -> Result<Option<Frame>> {
    if buf.len() < 8 {
        return Ok(None);
    }
    
    // Check magic
    if &buf[0..2] != &[0x42, 0x50] {
        // Discard until magic
        buf.advance(1);
        return Ok(None);
    }
    
    // Get payload length
    let payload_len = u32::from_be_bytes([buf[4], buf[5], buf[6], buf[7]]) as usize;
    let total_len = 8 + payload_len + 4; // header + payload + checksum
    
    if buf.len() < total_len {
        return Ok(None);
    }
    
    // Parse frame
    let frame_data = buf.split_to(total_len);
    Frame::decode(&frame_data[..]).map(Some)
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_tcp_frame() {
        let frame = Frame::data("test");
        let encoded = frame.encode();
        
        let decoded = Frame::decode(&encoded[..]).unwrap();
        assert_eq!(frame.payload, decoded.payload);
    }
}
