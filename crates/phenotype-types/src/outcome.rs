//! Outcome type for OCI launch attempts.
//!
//! Extracted from `oci-lottery/src/oci.rs` to make the launch-result contract
//! available to any consumer (including future CI-grade smoke-test agents).

use serde::{Deserialize, Serialize};

/// The result of a single `oci compute instance launch` attempt.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LaunchOutcome {
    /// `true` when the launch succeeded (instance OCID was returned).
    pub success: bool,

    /// The instance OCID, if the launch succeeded.
    pub instance_ocid: Option<String>,

    /// Raw stdout from the `oci` CLI invocation.
    pub raw_stdout: String,

    /// Raw stderr from the `oci` CLI invocation.
    pub raw_stderr: String,

    /// `true` when the CLI returned a recognised "out of capacity" error so
    /// the caller can retry silently rather than treat it as a hard failure.
    pub out_of_capacity: bool,
}

impl LaunchOutcome {
    /// Shorthand for a successful launch.
    pub fn success(instance_ocid: impl Into<String>, stdout: impl Into<String>) -> Self {
        Self {
            success: true,
            instance_ocid: Some(instance_ocid.into()),
            raw_stdout: stdout.into(),
            raw_stderr: String::new(),
            out_of_capacity: false,
        }
    }

    /// Shorthand for a failure, optionally tagged as out-of-capacity.
    pub fn failure(stderr: impl Into<String>, out_of_capacity: bool) -> Self {
        Self {
            success: false,
            instance_ocid: None,
            raw_stdout: String::new(),
            raw_stderr: stderr.into(),
            out_of_capacity,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn success_outcome() {
        let o = LaunchOutcome::success("ocid1.instance..1", "{}");
        assert!(o.success);
        assert_eq!(o.instance_ocid.as_deref(), Some("ocid1.instance..1"));
        assert!(!o.out_of_capacity);
    }

    #[test]
    fn failure_outcome() {
        let o = LaunchOutcome::failure("Out of host capacity", true);
        assert!(!o.success);
        assert!(o.instance_ocid.is_none());
        assert!(o.out_of_capacity);
    }
}
