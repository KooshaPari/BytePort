//! D-Bus service skeleton for `org.byteport.LinuxBridge1`.
//!
//! This module holds the `#[zbus::interface]`-annotated impl block that
//! will eventually expose the bridge over the session bus. Because we
//! only ship the scaffold in this PR, the interface body is intentionally
//! a no-op (returning `Ok(())`); the real methods will be wired in once
//! the v8.2 rollout lands.
//!
//! Linux-only — guarded by `#[cfg(target_os = "linux")]` on every item
//! so the rest of the workspace compiles unchanged on macOS / Windows.

/// Inert implementation of the `org.byteport.LinuxBridge1` interface
/// used by the scaffold on every target.
///
/// On Linux this struct is wired up by [`zbus::interface`] so the
/// service can be served at [`crate::OBJECT_PATH`] under
/// [`crate::BUS_NAME`]. On other targets it is a plain value type that
/// keeps `cfg`-agnostic code compiling.
pub struct LinuxBridgeSkeleton;

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
#[zbus::interface(name = "org.byteport.LinuxBridge1")]
impl LinuxBridgeSkeleton {
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
    #[zbus(signal)]
    async fn bytes_received(
        ctxt: &zbus::object_server::SignalContext<'_>,
        session_id: u64,
        byte_count: u64,
    ) -> zbus::Result<()>;
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
        let v = s.ping().expect("ping");
        assert_eq!(v, "pong");
    }

    #[cfg(not(target_os = "linux"))]
    #[test]
    fn skeleton_ping_errors_off_linux() {
        // On non-Linux targets we don't have the real `ping` impl, so we
        // assert the struct surface + constants are intact instead.
        let _s = LinuxBridgeSkeleton;
        assert_eq!(crate::BUS_NAME, "org.byteport.LinuxBridge1");
        assert_eq!(crate::OBJECT_PATH, "/org/byteport/LinuxBridge1");
    }
}
