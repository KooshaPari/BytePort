//! Pluggable deployment-engine adapters.
//!
//! Each sub-module implements [`Engine`](crate::engine::Engine) for a specific
//! backend. The [`mock`] adapter is always available and used by tests.
//! Other adapters are feature-gated or stubbed.

pub mod docker;
pub mod mock;
#[cfg(feature = "nvms")]
pub mod nvms;

// Future adapters (Phase 3B+):
// pub mod k8s;     // Kubernetes
// pub mod aws_ecs; // AWS ECS / Fargate
