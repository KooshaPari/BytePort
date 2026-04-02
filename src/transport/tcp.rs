//! TCP transport implementation for BytePort

use tokio::net::{TcpListener, TcpStream};
use tokio::io::{AsyncReadExt, AsyncWriteExt, BufReader, BufWriter};
use bytes::{Buf, Bytes, BytesMut};
use std::net::SocketAddr;
use tracing::{debug, error, info, warn};

use crate::error::{BytePortError, Result};
use crate::protocol::{Frame, MAX_FRAME_SIZE};

/// TCP server for BytePort protocol
pub struct TcpServer {
    listener: TcpListener,
}

impl TcpServer {
    /// Bind to a socket address
    pub async fn bind(addr: impl Into<SocketAddr>) -> Result<Self> {
        let addr = addr.into();
        let listener = TcpListener::bind(&addr).await?;
        info!("TCP server bound to {}", addr);
        
        Ok(Self { listener })
    }
    
    /// Accept incoming connections
    pub async fn accept(&self) -> Result<TcpConnection> {
        let (stream, addr) = self.listener.accept().await?;
        debug!("Accepted connection from {}", addr);
        
        Ok(TcpConnection::new(stream, addr))
    }
}

/// TCP connection for BytePort protocol
pub struct TcpConnection {
    stream: TcpStream,
    addr: SocketAddr,
    read_buffer: BytesMut,
}

impl TcpConnection {
    /// Create a new TCP connection
    pub fn new(stream: TcpStream, addr: SocketAddr) -> Self {
        Self {
            stream,
            addr,
            read_buffer: BytesMut::with_capacity(4096),
        }
    }
    
    /// Connect to a remote address
    pub async fn connect(addr: impl Into<SocketAddr>) -> Result<Self> {
        let addr = addr.into();
        let stream = TcpStream::connect(&addr).await?;
        info!("Connected to {}", addr);
        
        Ok(Self::new(stream, addr))
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
        debug!("Sent {} bytes to {}", encoded.len(), self.addr);
        Ok(())
    }
    
    /// Receive a frame
    pub async fn recv_frame(&mut self) -> Result<Frame> {
        loop {
            // Try to parse a frame from the buffer
            if self.read_buffer.len() >= 8 {
                // Check magic bytes
                if &self.read_buffer[0..2] == &[0x42, 0x50] {
                    // Get payload length from header
                    let payload_len = u32::from_be_bytes([
                        self.read_buffer[4],
                        self.read_buffer[5],
                        self.read_buffer[6],
                        self.read_buffer[7],
                    ]) as usize;
                    
                    let total_len = 8 + payload_len + 4; // header + payload + checksum
                    
                    if self.read_buffer.len() >= total_len {
                        // We have a complete frame
                        let frame_data = self.read_buffer.split_to(total_len);
                        return Frame::decode(&frame_data[..]);
                    }
                } else {
                    // Invalid magic, discard bytes until we find magic
                    warn!("Invalid magic bytes, discarding byte");
                    self.read_buffer.advance(1);
                }
            }
            
            // Need more data
            let mut temp_buf = [0u8; 4096];
            let n = self.stream.read(&mut temp_buf).await?;
            
            if n == 0 {
                return Err(BytePortError::Connection("Connection closed".to_string()));
            }
            
            self.read_buffer.extend_from_slice(&temp_buf[..n]);
            
            // Prevent buffer from growing too large
            if self.read_buffer.len() > MAX_FRAME_SIZE as usize * 2 {
                return Err(BytePortError::InvalidFrame("Buffer overflow".to_string()));
            }
        }
    }
    
    /// Send raw bytes
    pub async fn send_bytes(&mut self, data: impl Into<Bytes>) -> Result<()> {
        let data = data.into();
        self.stream.write_all(&data).await?;
        self.stream.flush().await?;
        Ok(())
    }
    
    /// Close the connection
    pub async fn close(mut self) -> Result<()> {
        self.stream.shutdown().await?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::time::{timeout, Duration};
    
    #[tokio::test]
    async fn test_tcp_connection() {
        // Start server
        let server = TcpServer::bind("127.0.0.1:0".parse::<std::net::SocketAddr>().unwrap()).await.unwrap();
        let addr = server.listener.local_addr().unwrap();
        
        // Spawn server task
        let server_task = tokio::spawn(async move {
            let mut conn = server.accept().await.unwrap();
            let frame = conn.recv_frame().await.unwrap();
            assert_eq!(frame.frame_type, crate::protocol::FrameType::Data);
            
            let response = Frame::ack(1);
            conn.send_frame(&response).await.unwrap();
        });
        
        // Connect client
        let mut client = TcpConnection::connect(addr).await.unwrap();
        
        // Send frame
        let frame = Frame::data("Hello, World!");
        client.send_frame(&frame).await.unwrap();
        
        // Receive response
        let response = timeout(Duration::from_secs(1), client.recv_frame()).await.unwrap().unwrap();
        assert_eq!(response.frame_type, crate::protocol::FrameType::Ack);
        
        server_task.await.unwrap();
    }
}
