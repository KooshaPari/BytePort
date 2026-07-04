//! BytePort macOS Share Extension — FFI scaffold.
//!
//! Wraps the Swift `ShareViewController` (compiled via swift-rs) behind a
//! minimal Rust API. The full byte-routing pipeline (NSItemProvider ->
//! BytePort IPC -> content-length token) is NOT implemented in this scaffold.
//!
//! Activation path: when a user shares bytes (file, image, pasteboard item)
//! from any macOS app, the OS instantiates `ShareViewController`. That
//! controller calls back into this Rust shim, which is responsible for
//! hand-off to the BytePort host process.
//!
//! See `docs/ffi/MOBILE.md` for the full architecture.

#![cfg_attr(all(target_os = "macos", feature = "_no_real_impl"), allow(dead_code))]

use thiserror::Error;

#[derive(Debug, Error)]
pub enum ShareError {
    #[error("share extension not yet wired to a host process")]
    NotWired,
    #[error("invalid item provider: {0}")]
    InvalidProvider(String),
}

/// Opaque handle to a share session, as seen from the host process.
pub struct ShareHandle {
    pub session_id: u64,
}

/// Stub entry-point invoked from `ShareViewController.swift`.
///
/// On macOS only; on other targets this returns `ShareError::NotWired`.
pub fn open_share_session(item_count: usize) -> Result<ShareHandle, ShareError> {
    if item_count == 0 {
        return Err(ShareError::InvalidProvider(
            "item_count must be > 0".to_string(),
        ));
    }

    #[cfg(target_os = "macos")]
    {
        // Real implementation will call into the Swift side via the
        // swift-rs generated header. Stubbed here pending B6 wiring.
        Ok(ShareHandle {
            session_id: next_session_id(),
        })
    }

    #[cfg(not(target_os = "macos"))]
    {
        Err(ShareError::NotWired)
    }
}

#[cfg(target_os = "macos")]
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
            open_share_session(0),
            Err(ShareError::InvalidProvider(_))
        ));
    }

    #[cfg(target_os = "macos")]
    #[test]
    fn opens_session_on_macos() {
        let handle = open_share_session(3).expect("ok");
        assert!(handle.session_id > 0);
    }

    #[cfg(not(target_os = "macos"))]
    #[test]
    fn refused_on_non_macos() {
        assert!(matches!(
            open_share_session(1),
            Err(ShareError::NotWired)
        ));
    }
}