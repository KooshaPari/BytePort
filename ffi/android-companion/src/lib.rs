//! BytePort Android Companion — FFI scaffold.
//!
//! Provides a JNI entry point that the Kotlin share-sheet receiver calls
//! when a user shares bytes (file, image, text) from another Android app.
//!
//! SCAFFOLD ONLY. The actual byte hand-off to the BytePort host process
//! is not implemented. See `docs/ffi/MOBILE.md` for the design.
//!
//! Build target: `aarch64-linux-android`, `x86_64-linux-android`.

#![cfg_attr(all(target_os = "android", feature = "_no_real_impl"), allow(dead_code))]

use thiserror::Error;

#[derive(Debug, Error)]
pub enum CompanionError {
    #[error("JNI environment not available (off-Android build?)")]
    NoJvm,
    #[error("invalid share intent: {0}")]
    InvalidIntent(String),
}

/// Opaque handle representing an in-flight share session on the device.
pub struct CompanionHandle {
    pub session_id: u64,
}

/// Stub entry-point invoked from the Kotlin side.
///
/// On Android, this would be reached via JNI; on other targets it
/// returns `CompanionError::NoJvm` so the workspace still compiles
/// during cross-compile matrix setup.
pub fn open_companion_session(item_count: usize) -> Result<CompanionHandle, CompanionError> {
    if item_count == 0 {
        return Err(CompanionError::InvalidIntent("item_count must be > 0".to_string()));
    }

    #[cfg(target_os = "android")]
    {
        Ok(CompanionHandle {
            session_id: next_session_id(),
        })
    }

    #[cfg(not(target_os = "android"))]
    {
        Err(CompanionError::NoJvm)
    }
}

#[cfg(target_os = "android")]
fn next_session_id() -> u64 {
    use std::sync::atomic::{AtomicU64, Ordering};
    static COUNTER: AtomicU64 = AtomicU64::new(1);
    COUNTER.fetch_add(1, Ordering::Relaxed)
}

/// JNI bridge — `Java_io_byteport_android_BytePortCompanion_openSession`.
///
/// Real wiring (deferred to B6 of the v8.1 rollout): the Kotlin receiver
/// receives `Intent.EXTRA_STREAM`, reads the URI via
/// `ContentResolver`, and calls into this entry point to hand bytes off.
#[cfg(target_os = "android")]
#[allow(non_snake_case)]
pub mod jni_bridge {
    use super::{open_companion_session, CompanionError};

    #[no_mangle]
    pub extern "system" fn Java_io_byteport_android_BytePortCompanion_openSession(
        _env: jni::JNIEnv<'_>,
        _class: jni::objects::JClass<'_>,
        item_count: jni::sys::jint,
    ) -> jni::sys::jlong {
        match open_companion_session(item_count as usize) {
            Ok(handle) => handle.session_id as jni::sys::jlong,
            Err(CompanionError::InvalidIntent(reason)) => {
                // Real impl: throw IllegalArgumentException with `reason`.
                let _ = reason;
                -1
            }
            Err(_) => -1,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_zero_item_count() {
        assert!(matches!(
            open_companion_session(0),
            Err(CompanionError::InvalidIntent(_))
        ));
    }

    #[cfg(target_os = "android")]
    #[test]
    fn opens_session_on_android() {
        let handle = open_companion_session(3).expect("ok");
        assert!(handle.session_id > 0);
    }

    #[cfg(not(target_os = "android"))]
    #[test]
    fn refused_off_android() {
        assert!(matches!(open_companion_session(1), Err(CompanionError::NoJvm)));
    }
}
