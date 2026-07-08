//! # byteport-engine
//!
//! Pluggable deployment-engine abstraction for BytePort (Phase 3A of
//! `plans/2026-07-04-byteport-evolution-v1.md`). Defines a single
//! `Engine` trait that every concrete backend (Docker, Firecracker,
//! Kubernetes, AWS, ...) implements. Callers dispatch by name through an
//! `EngineRegistry`.
//!
//! This crate is **purely additive scaffolding**. It does not touch the
//! existing Go backend (`backend/main.go`) or any `backend/nvms/*.go` paths.
//! Real wiring into the Go API is deferred to the follow-up T1 HTTP-sidecar
//! task described in the evolution plan.
//!
//! ## Modules
//!
//! | Module                 | Description                                         |
//! |------------------------|-----------------------------------------------------|
//! | [`engine`]             | Core [`engine::Engine`] trait + value types         |
//! | [`adapters::mock`]     | In-memory `MockEngine` for tests                    |
//! | [`adapters::docker`]   | Stub `DockerEngine` (real bollard integration TBD)  |
//! | [`adapters::nvms`]     | `NvmsHttpAdapter` — talks to NVMS Spin daemon       |
//! | [`registry`]           | `EngineRegistry` dispatcher keyed by engine name    |
//!
//! ## Status
//!
//! - Engine trait + value types: implemented (307 lines).
//! - Mock adapter: implemented with `tokio::sync::RwLock` + `uuid::Uuid`.
//! - Docker adapter: stub (returns `EngineError::NotImplemented`).
//! - NVMS HTTP adapter: implemented under default `nvms` feature gate (reqwest).
//! - Registry: implemented with registration, lookup, and duplicate-protection.
//! - Wire-up into Go backend: **not started** (T1 HTTP sidecar — next task).

pub mod adapters;
pub mod engine;
pub mod registry;

pub use engine::{
    DeploymentId, DeploymentState, DeploymentStatus, Engine, EngineError, EngineManifest, EnvVar,
    LogLine, LogOptions, LogStream, PortMapping, PortProtocol, ServiceSpec,
};
pub use registry::EngineRegistry;
