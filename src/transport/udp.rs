//! UDP transport implementation for BytePort

use tokio::net::{UdpSocket, ToSocketAddrs};
use bytes::{Bytes, BytesMut, Buf, BufMut};
use std::net::SocketAddr;
use tracing::{debug, error, info, warn};

use crate::error::{BytePortError, Result};
use crate::protocol::{Frame, MAX_FRAME_SIZE};

/// UDP server for BytePort protocol
pub struct UdpServer {
    socket: UdpSocket,
    local_addr: SocketAddr,
}

impl UdpServer {
    /// Bind to a socket address
    pub async fn bind(addr: impl ToSocketAddrs) -> Result<Self> {
        let socket = UdpSocket::bind(addr).await?;
        let local_addr = socket.local_addr()?;
        info!("UDP server bound to {}", local_addr);
        
        Ok(Self { socket, local_addr })
    }
    
    /// Get local address
    pub fn local_addr(&self) -> SocketAddr {
        self.local_addr
    }
    
    /// Receive a frame from any peer
    pub async fn recv_frame(&self) -> Result<(Frame, SocketAddr)> {
        let mut buf = vec![0u8; MAX_FRAME_SIZE as usize];
        let (len, peer) = self.socket.recv_from(&mut buf).await?;
        buf.truncate(len);
        
        let frame = Frame::decode(&buf[..])?;
        debug!("Received {} bytes from {} via UDP", len, peer);
        
        Ok((frame, peer))
    }
    
    /// Send a frame to a specific peer
    pub async fn send_frame(&self, frame: &Frame, peer: SocketAddr) -> Result<()> {
        let encoded = frame.encode();
        let len = self.socket.send_to(&encoded, peer).await?;
        debug!("Sent {} bytes to {} via UDP", len, peer);
        Ok(())
    }
}

/// UDP client for BytePort protocol
pub struct UdpClient {
    socket: UdpSocket,
    server_addr: SocketAddr,
}

impl UdpClient {
    /// Create a UDP client connected to a server
    pub async fn connect(server_addr: impl ToSocketAddrs) -> Result<Self> {
        let socket = UdpSocket::bind("0.0.0.0:0").await?;
        let server_addr = tokio::net::lookup_host(server_addr).await?
            .next()
            .ok_or_else(|| BytePortError::Connection("Invalid server address".to_string()))?;
        
        info!("UDP client connected to {}", server_addr);
        
        Ok(Self { socket, server_addr })
    }
    
    /// Send a frame to the server
    pub async fn send_frame(&self, frame: &Frame) -> Result<()> {
        let encoded = frame.encode();
        let len = self.socket.send_to(&encoded, self.server_addr).await?;
        debug!("Sent {} bytes to server via UDP", len);
        Ok(())
    }
    
    /// Receive a frame from the server
    pub async fn recv_frame(&self) -> Result<Frame> {
        let mut buf = vec![0u8; MAX_FRAME_SIZE as usize];
        let (len, _peer) = self.socket.recv_from(&mut buf).await?;
        
        buf.truncate(len);
        Frame::decode(&buf[..])
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::time::{timeout, Duration};
    
    #[tokio::test]
    async fn test_udp_server_client() {
        let server = UdpServer::bind("127.0.0.1:0").await.unwrap();
        let server_addr = server.local_addr();
        
        let client = UdpClient::connect(server_addr).await.unwrap();
        
        let server_task = tokio::spawn(async move {
            let (frame, peer) = timeout(Duration::from_secs(1), server.recv_frame()).await.unwrap().unwrap();
            assert_eq!(frame.frame_type, crate::protocol::FrameType::Data);
            
            let response = Frame::ack(1);
            server.send_frame(&response, peer).await.unwrap();
        });
        
        let frame = Frame::data("Hello UDP!");
        client.send_frame(&frame).await.unwrap();
        
        let response = timeout(Duration::from_secs(1), client.recv_frame()).await.unwrap().unwrap();
        assert_eq!(response.frame_type, crate::protocol::FrameType::Ack);
        
        server_task.await.unwrap();
    }
}
