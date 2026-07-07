//! NanoVMS adapter — bridges BytePort's [`Engine`](crate::engine::Engine) trait
//! to NanoVMS sandbox / MicroVM management.
//!
//! NanoVMS is a Go-based hexagonal-architecture sandbox manager (Firecracker,
//! KVM, Lima-VZ, WASM). This adapter communicates with it over an HTTP/gRPC
//! daemon interface (or directly via the `nvms` CLI for local dev).
//!
//! ## Status
//!
//! - **HTTP daemon adapter**: stub (`EngineError::NotImplemented`). Ready for
//!   wiring once `nanovms/cmd/nanovms` exposes an HTTP API.
//! - **Process adapter** (direct `nvms` CLI calls): planned (Phase 3C).
//!
//! ## NanoVMS port mapping
//!
//! | NanoVMS port | byteport-engine method | Status |
//! |--------------|----------------------|--------|
//! | `SandboxPort.Create` → `deploy` | Creates a new sandbox/MicroVM | 🔄 |
//! | `SandboxPort.Start` | Starts the sandbox | 🔄 |
//! | `SandboxPort.Stop` → `stop` | Stops/destroys deployment | 🔄 |
//! | `SandboxPort.Get` → `status` | Poll deployment state | 🔄 |
//! | `SandboxPort.Logs` → `logs` | Stream logs | 🔄 |
//! | `SandboxPort.List` → `list` | List all deployments | 🔄 |
//! | `NetworkPort.*` | Network configuration | 🔄 |
//! | `VolumePort.*` | Volume/disk operations | 🔄 |
//!
//! Legend: ✅ implemented, 🔄 stub, ⬜ planned.

pub mod http;
pub use http::NvmsHttpAdapter;
// pub mod process;  // Phase 3C
