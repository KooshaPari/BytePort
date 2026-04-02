//! BytePort Examples

use byteport::{Client, Server, Frame, Message};
use serde::{Serialize, Deserialize};

/// Example message type
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ChatMessage {
    pub sender: String,
    pub content: String,
    pub timestamp: u64,
}

/// Echo server example
pub async fn run_echo_server(addr: &str) -> anyhow::Result<()> {
    let server = Server::bind(addr).await?;
    println!("Echo server listening on {}", addr);

    loop {
        let mut client = server.accept().await?;
        tokio::spawn(async move {
            loop {
                match client.recv::<ChatMessage>().await {
                    Ok(msg) => {
                        println!("Received from {}: {}", msg.sender, msg.content);
                        if let Err(e) = client.send(&msg).await {
                            eprintln!("Failed to echo: {}", e);
                            break;
                        }
                    }
                    Err(e) => {
                        eprintln!("Client disconnected: {}", e);
                        break;
                    }
                }
            }
        });
    }
}
