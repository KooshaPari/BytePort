//! ## phenotype-types
//!
//! Shared type definitions across the Phenotype ecosystem.
//!
//! ### OCI helpers extracted from duplicated sources
//!
//! | Type / fn      | Origin (oci-lottery)              | Origin (oci-post-acquire)        |
//! |----------------|-----------------------------------|----------------------------------|
//! | `OciInstance`  | `state::AcquiredInstance`         | `InstanceFile`                   |
//! | `LaunchOutcome`| `oci.rs::LaunchOutcome`           | —                                |
//! | `expand`       | `config.rs::dirs_home` (inline)   | `expand` / `dirs_home`           |
//! | `dirs_home`    | `config.rs::dirs_home`            | `dirs_home`                      |
//!
//! Downstream crates should depend on `phenotype-types` and delete their local
//! copies once the migration is complete.

pub mod instance;
pub mod outcome;
pub mod path;

pub use instance::OciInstance;
pub use outcome::LaunchOutcome;
pub use path::{dirs_home, expand};
