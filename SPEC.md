# BytePort Specification

> Binary protocol and serialization framework for efficient data transport

## Overview

BytePort provides a binary protocol and serialization framework optimized for high-performance data transport across services in the Phenotype ecosystem.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          BytePort Architecture                                 │
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                      Protocol Layer                                     │ │
│   │                                                                       │ │
│   │   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │ │
│   │   │  Binary  │ │   VarInt │ │  Schema  │ │ Version  │          │ │
│   │   │  Encoder │ │  Coding  │ │  Registry│ │ Control  │          │ │
│   │   └──────────┘ └──────────┘ └──────────┘ └──────────┘          │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                    │                                         │
│                                    ▼                                         │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                      Transport Layer                                  │ │
│   │                                                                       │ │
│   │   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │ │
│   │   │   TCP    │ │   UDP    │ │ WebSocket│ │   QUIC   │          │ │
│   │   │ Transport│ │ Transport│ │ Transport│ │ Transport│          │ │
│   │   └──────────┘ └──────────┘ └──────────┘ └──────────┘          │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Features

- **Binary Encoding**: Compact, efficient serialization
- **Multiple Transports**: TCP, UDP, WebSocket, QUIC support
- **Schema Registry**: Versioned message schemas
- **Compression**: Optional compression (zstd, lz4)
- **Encryption**: TLS/QUIC built-in

## Quick Start

```rust
use byteport::{Client, Server, Message};

#[derive(Message)]
struct MyMessage {
    id: u64,
    payload: Vec<u8>,
}

#[tokio::main]
async fn main() {
    // Server
    let server = Server::bind("0.0.0.0:8080").await.unwrap();
    server.handle(|msg: MyMessage| {
        println!("Received: {:?}", msg);
    }).await;
    
    // Client
    let client = Client::connect("localhost:8080").await.unwrap();
    client.send(MyMessage { id: 1, payload: vec![1, 2, 3] }).await.unwrap();
}
```

## Protocol Format

```
Frame Structure:
┌─────────┬─────────┬─────────┬─────────┬─────────┐
│  Magic  │ Version │  Type   │ Length  │ Payload │
│ 2 bytes │ 1 byte  │ 1 byte  │ 4 bytes │ N bytes │
└─────────┴─────────┴─────────┴─────────┴─────────┘

Total: 8 bytes header + variable payload
```

## Performance

| Metric | Target |
|--------|--------|
| Serialization | <1μs |
| Throughput | 1M+ msg/sec |
| Latency | <100μs |
| Overhead | <10% |

## References

- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Cap'n Proto](https://capnproto.org/)
- [FlatBuffers](https://google.github.io/flatbuffers/)
