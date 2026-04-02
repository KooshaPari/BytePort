//! BytePort - Binary protocol and serialization framework
//!
//! Provides efficient binary serialization and multiple transport options
//! for high-performance data transport across services.

pub mod error;
pub mod protocol;
pub mod transport;
pub mod codec;

pub use error::{BytePortError, Result};
pub use protocol::{Frame, FrameType, PROTOCOL_VERSION, MAX_FRAME_SIZE, varint};
pub use codec::FrameCodec;
pub use transport::{TcpServer, TcpConnection, UdpServer, UdpClient, WsServer, WsConnection};

use serde::{Serialize, Deserialize};
use bytes::Bytes;

/// A message that can be serialized and sent over BytePort
pub trait Message: Serialize + for<'de> Deserialize<'de> + Send + Sync + 'static {
    /// Encode message to bytes
    fn encode(&self) -> Result<Bytes> {
        let json = serde_json::to_vec(self)
            .map_err(|e| BytePortError::Serialization(e.to_string()))?;
        Ok(Bytes::from(json))
    }
    
    /// Decode message from bytes
    fn decode(data: &[u8]) -> Result<Self> {
        serde_json::from_slice(data)
            .map_err(|e| BytePortError::Deserialization(e.to_string()))
    }
}

/// Blanket implementation for types that implement Serialize + Deserialize
impl<T> Message for T 
where 
    T: Serialize + for<'de> Deserialize<'de> + Send + Sync + 'static 
{}

/// Client for BytePort protocol
pub struct Client {
    connection: transport::TcpConnection,
}

impl Client {
    /// Connect to a remote server
    pub async fn connect(addr: impl Into<std::net::SocketAddr>) -> Result<Self> {
        let connection = transport::TcpConnection::connect(addr).await?;
        Ok(Self { connection })
    }
    
    /// Send a message
    pub async fn send<M: Message>(&mut self, message: &M) -> Result<()> {
        let payload = message.encode()?;
        let frame = Frame::data(payload);
        self.connection.send_frame(&frame).await
    }
    
    /// Receive a message
    pub async fn recv<M: Message>(&mut self) -> Result<M> {
        let frame = self.connection.recv_frame().await?;
        M::decode(&frame.payload)
    }
    
    /// Close the connection
    pub async fn close(self) -> Result<()> {
        self.connection.close().await
    }
}

/// Server for BytePort protocol
pub struct Server {
    inner: transport::TcpServer,
}

impl Server {
    /// Bind to a socket address
    pub async fn bind(addr: impl Into<std::net::SocketAddr>) -> Result<Self> {
        let inner = transport::TcpServer::bind(addr).await?;
        Ok(Self { inner })
    }
    
    /// Accept a new connection
    pub async fn accept(&self) -> Result<Client> {
        let connection = self.inner.accept().await?;
        Ok(Client { connection })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde::{Serialize, Deserialize};
    
    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct TestMessage {
        id: u64,
        content: String,
    }
    
    #[test]
    fn test_message_encode_decode() {
        let msg = TestMessage {
            id: 42,
            content: "Hello".to_string(),
        };
        
        let encoded = msg.encode().unwrap();
        let decoded = TestMessage::decode(&encoded).unwrap();
        
        assert_eq!(msg, decoded);
    }
    
    #[test]
    fn test_frame_encode_decode() {
        let frame = Frame::data("Hello, World!");
        let encoded = frame.encode();
        let decoded = Frame::decode(&encoded).unwrap();
        
        assert_eq!(frame.frame_type, decoded.frame_type);
        assert_eq!(frame.payload, decoded.payload);
    }
}
