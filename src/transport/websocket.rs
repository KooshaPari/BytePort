//! WebSocket transport implementation for BytePort

use tokio_tungstenite::{accept_async, connect_async, tungstenite::Message, WebSocketStream, MaybeTlsStream};
use tokio::net::{TcpListener, TcpStream};
use futures::{SinkExt, StreamExt};
use bytes::Bytes;
use std::net::SocketAddr;
use tracing::{debug, error, info, warn};

use crate::error::{BytePortError, Result};
use crate::protocol::{Frame, MAX_FRAME_SIZE};

/// WebSocket server for BytePort protocol
pub struct WsServer {
    listener: TcpListener,
    local_addr: SocketAddr,
}

impl WsServer {
    /// Bind to a socket address
    pub async fn bind(addr: impl Into<SocketAddr>) -> Result<Self> {
        let addr = addr.into();
        let listener = TcpListener::bind(&addr).await?;
        let local_addr = listener.local_addr()?;
        info!("WebSocket server bound to {}", local_addr);
        
        Ok(Self { listener, local_addr })
    }
    
    /// Accept a WebSocket connection
    pub async fn accept(&self) -> Result<WsConnection> {
        let (stream, peer_addr) = self.listener.accept().await?;
        let ws_stream = accept_async(stream).await?;
        info!("WebSocket connection accepted from {}", peer_addr);
        
        Ok(WsConnection {
            ws_stream,
        })
    }
}

/// WebSocket connection
pub struct WsConnection {
    ws_stream: WebSocketStream<MaybeTlsStream<TcpStream>>,
}

impl WsConnection {
    /// Connect to a WebSocket server
    pub async fn connect(addr: impl AsRef<str>) -> Result<Self> {
        let url = addr.as_ref();
        let (ws_stream, _) = connect_async(url).await?;
        info!("Connected to WebSocket server at {}", url);
        
        Ok(Self {
            ws_stream,
        })
    }
    
    /// Send a frame as binary WebSocket message
    pub async fn send_frame(&mut self, frame: &Frame) -> Result<()> {
        let encoded = frame.encode();
        self.ws_stream.send(Message::Binary(encoded.to_vec())).await?;
        Ok(())
    }
    
    /// Receive a frame from WebSocket
    pub async fn recv_frame(&mut self) -> Result<Frame> {
        while let Some(msg) = self.ws_stream.next().await {
            match msg? {
                Message::Binary(data) => {
                    return Frame::decode(&data[..]);
                }
                Message::Close(_) => {
                    return Err(BytePortError::Connection("Connection closed".into()));
                }
                _ => continue,
            }
        }
        
        Err(BytePortError::Connection("WebSocket stream ended".into()))
    }
    
    /// Close the connection
    pub async fn close(mut self) -> Result<()> {
        self.ws_stream.close(None).await?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::protocol::FrameType;
    use tokio::time::{timeout, Duration};
    
    #[tokio::test]
    async fn test_websocket_roundtrip() {
        // Note: This test requires a WebSocket server to be running
        // For a proper test, you'd mock the connection
        // This is a basic smoke test that the types compile correctly
        let _ = Frame::data("test");
    }
}
