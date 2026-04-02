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

/// Run a simple client example
pub async fn run_client(addr: &str) -> anyhow::Result<()> {
    let mut client = Client::connect(addr).await?;
    
    let msg = ChatMessage {
        sender: "client".to_string(),
        content: "Hello!".to_string(),
        timestamp: 1234567890,
    };
    
    client.send(&msg).await?;
    println!("Sent message");
    
    let response = client.recv::<ChatMessage>().await?;
    println!("Received: {:?}", response);
    
    Ok(())
}

/// Benchmark throughput between client and server
pub async fn benchmark_throughput(addr: &str, message_count: usize) -> anyhow::Result<f64> {
    let mut client = Client::connect(addr).await?;
    let msg = ChatMessage {
        sender: "bench".to_string(),
        content: "x".repeat(100),
        timestamp: 0,
    };
    
    let start = std::time::Instant::now();
    
    for _ in 0..message_count {
        client.send(&msg).await?;
        let _ = client.recv::<ChatMessage>().await?;
    }
    
    let elapsed = start.elapsed().as_secs_f64();
    let throughput = message_count as f64 / elapsed;
    
    Ok(throughput)
}
