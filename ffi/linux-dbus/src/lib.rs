//! BytePort Linux D-Bus Bridge — FFI scaffold.
//!
//! Wraps the zbus 4 service skeleton for `org.byteport.LinuxBridge1` behind
//! a minimal Rust API. The full byte-routing pipeline (D-Bus method call →
//! BytePort IPC → content-length token) is **not** implemented in this
//! scaffold.
//!
//! Activation path: a host process (Tauri desktop shell, or any BytePort
//! daemon) registers a service under the well-known name
//! `org.byteport.LinuxBridge1`. Companion apps (GNOME Shell extensions,
//! KDE Connect share plugins, Nautilus script bindings, etc.) call into
//! `org.byteport.LinuxBridge1` to hand bytes off.
//!
//! See `docs/ffi/MOBILE.md` for the cross-FFI trifecta rationale and the
//! activation matrix.
//!
//! # Platform gating
//!
//! On `target_os = "linux"` this crate links against `zbus`, `zvariant`,
//! and `tokio`. On every other target the public API returns
//! [`LinuxBridgeError::NoBus`] so the workspace still compiles on macOS
//! and Windows without pulling Linux-only deps.
//!
//! # Layout
//!
//! - [`dbus`] — `#[zbus::dbus_interface]` skeleton for `org.byteport.LinuxBridge1`.
//! - [`ipc`] — client-side method calls + signal subscriptions.
//! - [`codegen`] — placeholder for the generated introspection XML + trait
//!   expansion produced by `zbus::dbus_interface` once we move off the
//!   scaffold.

#![cfg_attr(all(target_os = "linux", feature = "_no_real_impl"), allow(dead_code))]
// `_no_real_impl` is a docs-only/CI toggle, not a real Cargo feature. Suppress
// the unexpected-cfgs lint that -D warnings would otherwise turn into an error.
#![allow(unexpected_cfgs)]

use thiserror::Error;

pub mod codegen;
pub mod dbus;
pub mod ipc;

#[derive(Debug, Error)]
pub enum LinuxBridgeError {
    /// Returned on every non-Linux target. Keeps the workspace compiling
    /// on macOS / Windows CI without pulling zbus into the dep graph.
    #[error("D-Bus session bus unavailable (off-Linux build?)")]
    NoBus,

    /// Returned when the well-known name `org.byteport.LinuxBridge1`
    /// cannot be claimed (already taken, session bus not running).
    #[error("failed to acquire org.byteport.LinuxBridge1: {0}")]
    NameTaken(String),

    /// Returned when an inbound D-Bus call carries an invalid payload.
    #[error("invalid byte hand-off: {0}")]
    InvalidHandoff(String),

    /// Returned when the zbus method reply itself fails (channel closed,
    /// signature mismatch, etc.).
    #[error("D-Bus call failed: {0}")]
    CallFailed(String),
}

/// Opaque handle representing an in-flight share session on the desktop.
pub struct LinuxBridgeHandle {
    pub session_id: u64,
}

/// Well-known bus name reserved for the BytePort Linux bridge.
///
/// Registered via `zbus::connection::Builder::name(...)` once we move
/// off the scaffold.
pub const BUS_NAME: &str = "org.byteport.LinuxBridge1";

/// Object path under `BUS_NAME` that exposes the bridge interface.
pub const OBJECT_PATH: &str = "/org/byteport/LinuxBridge1";

/// Stub entry-point invoked from the D-Bus dispatch loop (or from a
/// companion app's proxy call).
///
/// On Linux, this is the future home of the real session registration.
/// On other targets it returns [`LinuxBridgeError::NoBus`].
pub fn open_linux_bridge_session(item_count: usize) -> Result<LinuxBridgeHandle, LinuxBridgeError> {
    if item_count == 0 {
        return Err(LinuxBridgeError::InvalidHandoff("item_count must be > 0".to_string()));
    }

    #[cfg(target_os = "linux")]
    {
        // Real implementation will:
        //   1. Acquire a `zbus::Connection` to the session bus.
        //   2. Build an `org.byteport.LinuxBridge1` interface impl
        //      (see `dbus::LinuxBridge1`).
        //   3. Serve the interface at `OBJECT_PATH` under `BUS_NAME`.
        //
        // Stubbed here pending B6 wiring.
        Ok(LinuxBridgeHandle {
            session_id: next_session_id(),
        })
    }

    #[cfg(not(target_os = "linux"))]
    {
        Err(LinuxBridgeError::NoBus)
    }
}

#[cfg(target_os = "linux")]
fn next_session_id() -> u64 {
    use std::sync::atomic::{AtomicU64, Ordering};
    static COUNTER: AtomicU64 = AtomicU64::new(1);
    COUNTER.fetch_add(1, Ordering::Relaxed)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_zero_item_count() {
        assert!(matches!(
            open_linux_bridge_session(0),
            Err(LinuxBridgeError::InvalidHandoff(_))
        ));
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn opens_session_on_linux() {
        let handle = open_linux_bridge_session(3).expect("ok");
        assert!(handle.session_id > 0);
    }

    #[cfg(not(target_os = "linux"))]
    #[test]
    fn refused_off_linux() {
        assert!(matches!(open_linux_bridge_session(1), Err(LinuxBridgeError::NoBus)));
    }

    #[test]
    fn bus_name_constant_is_canonical() {
        assert_eq!(BUS_NAME, "org.byteport.LinuxBridge1");
        assert!(OBJECT_PATH.starts_with('/'));
    }
}
