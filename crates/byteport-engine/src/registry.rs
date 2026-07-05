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
}
