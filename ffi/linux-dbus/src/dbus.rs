//! D-Bus service skeleton for `org.byteport.LinuxBridge1`.
//!
//! This module holds the `#[zbus::dbus_interface]`-annotated struct that
//! will eventually expose the bridge over the session bus. Because we
//! only ship the scaffold in this PR, the interface body is intentionally
//! a no-op (returning `Ok(())`); the real methods will be wired in once
//! the v8.2 rollout lands.
//!
//! Linux-only — guarded by `#[cfg(target_os = "linux")]` on every item
//! so the rest of the workspace compiles unchanged on macOS / Windows.

#[cfg(target_os = "linux")]
use zbus::dbus_interface;

/// D-Bus interface served at [`crate::OBJECT_PATH`] under
/// [`crate::BUS_NAME`].
///
/// # Activation (deferred)
///
/// Real impl will:
/// 1. Acquire a `zbus::Connection` to the session bus.
/// 2. Register the well-known name.
/// 3. Serve this interface at `OBJECT_PATH`.
/// 4. Forward inbound `HandOff` calls to the BytePort host process.
#[cfg(target_os = "linux")]
#[dbus_interface(name = "org.byteport.LinuxBridge1")]
pub trait LinuxBridge1 {
    /// Acknowledge a pending hand-off from a companion app.
    ///
    /// Returns the assigned `session_id` as a `u64`. Companion apps are
    /// expected to keep this id and match it against the
    /// `BytesReceived` signal emitted from
    /// [`crate::ipc::subscribe_bridge_signals`].
    fn handoff(&mut self, item_count: u64) -> zbus::fdo::Result<u64> {
        let _ = item_count;
        // Scaffold: real impl allocates a session id from an atomic
        // counter and persists the in-flight hand-off state.
        Ok(0)
    }

    /// Ping method used by companion apps to verify the bridge is alive.
    fn ping(&self) -> zbus::fdo::Result<String> {
        Ok("pong".to_string())
    }

    /// Emitted after the host process has consumed the bytes handed off
    /// via `handoff(session_id)`.
    #[dbus_interface(signal)]
    fn bytes_received(session_id: u64, byte_count: u64) -> zbus::Result<()>;
}

/// Stub interface kept around on non-Linux targets so callers can still
/// reference the type in `cfg`-agnostic code.
#[cfg(not(target_os = "linux"))]
#[allow(dead_code)]
pub trait LinuxBridge1 {
    fn handoff(&mut self, item_count: u64) -> Result<u64, String>;
    fn ping(&self) -> Result<String, String>;
}

/// Inert implementation of [`LinuxBridge1`] used by the scaffold on
/// non-Linux targets and by Linux-side tests before the real wiring
/// lands.
pub struct LinuxBridgeSkeleton;

#[cfg(target_os = "linux")]
impl LinuxBridge1 for LinuxBridgeSkeleton {
    fn handoff(&mut self, _item_count: u64) -> zbus::fdo::Result<u64> {
        // Scaffold: real impl allocates a session id from an atomic
        // counter and persists the in-flight hand-off state.
        Ok(0)
    }

    fn ping(&self) -> zbus::fdo::Result<String> {
        Ok("pong".to_string())
    }
}

#[cfg(not(target_os = "linux"))]
impl LinuxBridge1 for LinuxBridgeSkeleton {
    fn handoff(&mut self, _item_count: u64) -> Result<u64, String> {
        Err("D-Bus bridge unavailable off-Linux".to_string())
    }

    fn ping(&self) -> Result<String, String> {
        Err("D-Bus bridge unavailable off-Linux".to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[cfg(target_os = "linux")]
    #[test]
    fn skeleton_ping_returns_pong() {
        let s = LinuxBridgeSkeleton;
        // `.ping()` on the impl returns `Result<String, _>`; we unwrap
        // here because the scaffold impl is infallible.
        let v = LinuxBridge1::ping(&s).expect("ping");
        assert_eq!(v, "pong");
    }

    #[cfg(not(target_os = "linux"))]
    #[test]
    fn skeleton_ping_errors_off_linux() {
        let s = LinuxBridgeSkeleton;
        assert!(LinuxBridge1::ping(&s).is_err());
    }
}
