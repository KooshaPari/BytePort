//! Canonical OCI instance record used across the compute/infra agents.
//!
//! Replaces the formerly duplicated `AcquiredInstance` (oci-lottery) and
//! `InstanceFile` (oci-post-acquire) types. Both downstream crates should
//! migrate to `phenotype_types::OciInstance` and remove their local structs.

use serde::{Deserialize, Serialize};

/// A record of a successfully-acquired OCI Always-Free instance.
///
/// # Compatibility
///
/// This type is designed to be serialised as the canonical
/// `oci-instance.json` written by `oci-lottery` and consumed by
/// `oci-post-acquire`.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OciInstance {
    /// OCID of the launched compute instance.
    pub instance_ocid: String,

    /// Region the instance was launched in (e.g. `"ap-tokyo-1"`).
    pub region: String,

    /// Availability domain name (e.g. `"AD-1"`).
    pub ad: String,

    /// Public IPv4 address, if known.
    #[serde(default)]
    pub public_ip: Option<String>,

    /// ISO-8601 timestamp of when the instance was acquired.
    pub acquired_at: String,
}

impl OciInstance {
    /// Create a new instance record.
    pub fn new(
        instance_ocid: impl Into<String>,
        region: impl Into<String>,
        ad: impl Into<String>,
        public_ip: Option<String>,
        acquired_at: impl Into<String>,
    ) -> Self {
        Self {
            instance_ocid: instance_ocid.into(),
            region: region.into(),
            ad: ad.into(),
            public_ip,
            acquired_at: acquired_at.into(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn roundtrip_json() {
        let inst = OciInstance::new(
            "ocid1.instance.oc1..example",
            "us-ashburn-1",
            "AD-1",
            Some("10.0.0.1".into()),
            "2026-06-25T00:00:00Z",
        );
        let json = serde_json::to_string(&inst).unwrap();
        let back: OciInstance = serde_json::from_str(&json).unwrap();
        assert_eq!(back.instance_ocid, inst.instance_ocid);
        assert_eq!(back.public_ip, Some("10.0.0.1".into()));
    }

    #[test]
    fn deserialise_no_public_ip() {
        let json = r#"{
            "instance_ocid": "ocid1.instance.oc1..example",
            "region": "ap-tokyo-1",
            "ad": "AD-1",
            "acquired_at": "2026-06-25T00:00:00Z"
        }"#;
        let inst: OciInstance = serde_json::from_str(json).unwrap();
        assert_eq!(inst.public_ip, None);
    }
}
