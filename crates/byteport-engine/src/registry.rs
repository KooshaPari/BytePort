//! [`EngineRegistry`] — dispatches calls to a named engine.
//!
//! Engines are registered at startup by name (e.g. `"docker"`,
//! `"firecracker"`, `"mock"`). Callers obtain a reference via
//! [`EngineRegistry::get`] and then call the [`Engine`] trait methods
//! directly.
//!
//! # Example
//!
//! ```
//! use byteport_engine::registry::EngineRegistry;
//! use byteport_engine::adapters::mock::MockEngine;
//!
//! let mut reg = EngineRegistry::new();
//! reg.register("mock", Box::new(MockEngine::new()));
//!
//! let engine = reg.get("mock").expect("mock engine registered");
//! assert_eq!(engine.name(), "mock");
//! ```

use std::collections::HashMap;

use crate::adapters::mock::MockEngine;
use crate::adapters::nvms::NvmsHttpAdapter;
use crate::engine::Engine;

/// Thread-safe registry of named [`Engine`] implementations.
///
/// # Errors
///
/// - [`RegistryError::NotFound`] when the engine name is not registered.
/// - [`RegistryError::AlreadyRegistered`] on duplicate registration.
#[derive(Debug)]
pub struct EngineRegistry {
    engines: HashMap<&'static str, Box<dyn Engine>>,
}

/// Registry-specific errors.
#[derive(Debug, thiserror::Error)]
pub enum RegistryError {
    #[error("engine not registered: {0}")]
    NotFound(String),
    #[error("engine already registered: {0}")]
    AlreadyRegistered(&'static str),
}

impl EngineRegistry {
    /// Create an empty registry.
    pub fn new() -> Self {
        Self {
            engines: HashMap::new(),
        }
    }

    /// Register an engine by name. Returns an error if the name is taken.
    pub fn register(
        &mut self,
        name: &'static str,
        engine: Box<dyn Engine>,
    ) -> Result<(), RegistryError> {
        if self.engines.contains_key(name) {
            return Err(RegistryError::AlreadyRegistered(name));
        }
        self.engines.insert(name, engine);
        Ok(())
    }

    /// Get a reference to a registered engine.
    pub fn get(&self, name: &str) -> Result<&dyn Engine, RegistryError> {
        self.engines
            .get(name)
            .map(|b| b.as_ref() as &dyn Engine)
            .ok_or_else(|| RegistryError::NotFound(name.to_owned()))
    }

    /// List all registered engine names.
    pub fn names(&self) -> Vec<&'static str> {
        self.engines.keys().copied().collect()
    }

    /// Number of registered engines.
    pub fn len(&self) -> usize {
        self.engines.len()
    }

    /// True if no engines are registered.
    pub fn is_empty(&self) -> bool {
        self.engines.is_empty()
    }
}

impl Default for EngineRegistry {
    fn default() -> Self {
        Self::new()
    }
}

/// Register the well-known engine set against an existing registry.
///
/// Includes:
/// - `"mock"` — [`MockEngine`] (always registered, safe for tests).
/// - `"nvms"` — [`NvmsHttpAdapter`], configured from `NVMS_DAEMON_URL`
///   (default `http://127.0.0.1:9700`).
///
/// Caller-chosen engines (e.g. `"docker"`) are not registered here — the
/// operator decides whether to register them after construction.
impl EngineRegistry {
    pub fn register_defaults(&mut self) -> Result<&mut Self, RegistryError> {
        self.register("mock", Box::new(MockEngine::new()))?;
        self.register("nvms", Box::new(NvmsHttpAdapter::from_env()))?;
        Ok(self)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::adapters::mock::MockEngine;

    #[test]
    fn register_and_get() {
        let mut reg = EngineRegistry::new();
        reg.register("mock", Box::new(MockEngine::new())).unwrap();
        let engine = reg.get("mock").unwrap();
        assert_eq!(engine.name(), "mock");
    }

    #[test]
    fn duplicate_registration_fails() {
        let mut reg = EngineRegistry::new();
        reg.register("mock", Box::new(MockEngine::new())).unwrap();
        let err = reg.register("mock", Box::new(MockEngine::new())).unwrap_err();
        assert!(matches!(err, RegistryError::AlreadyRegistered("mock")));
    }

    #[test]
    fn missing_engine_returns_not_found() {
        let reg = EngineRegistry::new();
        let err = reg.get("nonexistent").unwrap_err();
        assert!(matches!(err, RegistryError::NotFound(_)));
    }

    #[test]
    fn list_names() {
        let mut reg = EngineRegistry::new();
        reg.register("mock", Box::new(MockEngine::new())).unwrap();
        reg.register("docker", Box::new(crate::adapters::docker::DockerEngine))
            .unwrap();
        let mut names = reg.names();
        names.sort();
        assert_eq!(names, vec!["docker", "mock"]);
    }

    #[test]
    fn register_defaults_includes_mock_and_nvms() {
        // Ensure no NVMS_DAEMON_URL leaks from the env into the test.
        // SAFETY: tests run on a single thread for this case; clearing an env var
        // is safe given the surrounding test harness. Use std::env::remove_var
        // for portability.
        std::env::remove_var("NVMS_DAEMON_URL");

        let mut reg = EngineRegistry::new();
        reg.register_defaults().unwrap();
        let mut names = reg.names();
        names.sort();
        assert_eq!(names, vec!["mock", "nvms"]);

        // NVMS adapter should be wired and report the expected name.
        let nvms = reg.get("nvms").unwrap();
        assert_eq!(nvms.name(), "nvms");
    }
}
