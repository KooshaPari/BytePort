//! YAML-based DagUnit executor.
//!
//! The executor reads a [`DagUnitManifest`] from YAML, builds a DAG from the
//! `depends_on` relationships, topologically sorts the units, and then
//! executes them in order (respecting the DAG).
//!
//! # Example
//!
//! ```rust,ignore
//! use pheno_dag::executor::{ExecResult, Executor};
//!
//! let manifest_yaml = r#"
//! version: "1.0.0"
//! units:
//!   - id: F1
//!     title: Build
//!     epic: epic_F
//!     repo: KooshaPari/BytePort
//!   - id: F2
//!     title: Test
//!     epic: epic_F
//!     repo: KooshaPari/BytePort
//!     depends_on: [F1]
//! "#;
//!
//! let mut exec = Executor::from_yaml(manifest_yaml).unwrap();
//! let results = exec.run().unwrap();
//! assert_eq!(results.len(), 2);
//! ```

use std::collections::HashMap;

use byteport_dag::dag::Dag;
use byteport_dag::topo;
use thiserror::Error;

use crate::unit::{DagUnit, DagUnitManifest, DagUnitStatus};

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

/// Errors that can occur during DAG execution.
#[derive(Debug, Error)]
pub enum ExecError {
    /// YAML parse error.
    #[error("YAML error: {0}")]
    Yaml(#[from] serde_yaml::Error),

    /// DAG cycle detected.
    #[error("DAG cycle detected: {0}")]
    Cycle(String),

    /// A unit referenced in depends_on was not found.
    #[error("dependency `{dependency}` required by unit `{unit}` not found in manifest")]
    MissingDependency {
        /// The unit that declared the dependency.
        unit: String,
        /// The dependency that was not found.
        dependency: String,
    },

    /// A prerequisite check failed.
    #[error("prerequisite failed for unit `{unit}`: {detail}")]
    PrerequisiteFailed {
        /// The unit whose prerequisite failed.
        unit: String,
        /// Details of the failure.
        detail: String,
    },

    /// A gate check failed.
    #[error("gate failed for unit `{unit}`: {detail}")]
    GateFailed {
        /// The unit whose gate failed.
        unit: String,
        /// Details of the failure.
        detail: String,
    },
}

// ---------------------------------------------------------------------------
// Execution result
// ---------------------------------------------------------------------------

/// The result of executing a single DagUnit.
#[derive(Debug, Clone)]
pub struct ExecResult {
    /// The unit that was executed.
    pub unit: DagUnit,
    /// Whether the unit succeeded.
    pub success: bool,
    /// Optional error message on failure.
    pub error: Option<String>,
}

// ---------------------------------------------------------------------------
// Executor
// ---------------------------------------------------------------------------

/// A DAG-aware executor that runs DagUnits in dependency order.
pub struct Executor {
    /// The in-memory manifest.
    manifest: DagUnitManifest,
    /// Quick lookup: id -> unit index.
    index: HashMap<String, usize>,
    /// The resolved execution order (topologically sorted IDs).
    order: Vec<String>,
}

impl Executor {
    /// Create an executor from a YAML string.
    ///
    /// Parses the manifest, resolves the dependency DAG, and computes a
    /// topological execution order.
    pub fn from_yaml(yaml: &str) -> Result<Self, ExecError> {
        let manifest = DagUnitManifest::from_yaml(yaml)?;
        Self::from_manifest(manifest)
    }

    /// Create an executor from an already-parsed manifest.
    pub fn from_manifest(manifest: DagUnitManifest) -> Result<Self, ExecError> {
        // Build an index.
        let index: HashMap<String, usize> = manifest
            .units
            .iter()
            .enumerate()
            .map(|(i, u)| (u.id.clone(), i))
            .collect();

        // Build a DAG from depends_on relationships.
        let mut dag: Dag<String> = Dag::new();

        // Register all nodes.
        for unit in &manifest.units {
            dag.add_node(unit.id.clone())
                .map_err(|e| ExecError::Cycle(e.to_string()))?;
        }

        // Register edges.
        for unit in &manifest.units {
            for dep in &unit.depends_on {
                // Verify the dependency exists.
                if !index.contains_key(dep) {
                    return Err(ExecError::MissingDependency {
                        unit: unit.id.clone(),
                        dependency: dep.clone(),
                    });
                }
                dag.add_edge(dep.clone(), unit.id.clone())
                    .map_err(|e| ExecError::Cycle(e.to_string()))?;
            }
        }

        // Topological sort.
        let order: Vec<String> = match topo::kahn_sort(&dag) {
            Ok(list) => list.into_iter().cloned().collect(),
            Err(e) => return Err(ExecError::Cycle(e.to_string())),
        };

        Ok(Self {
            manifest,
            index,
            order,
        })
    }

    /// Return the resolved execution order (unit IDs).
    pub fn execution_order(&self) -> &[String] {
        &self.order
    }

    /// Return a reference to the manifest.
    pub fn manifest(&self) -> &DagUnitManifest {
        &self.manifest
    }

    /// Run all units in topological order.
    ///
    /// Each unit is executed by:
    /// 1. Checking its prerequisites
    /// 2. Running the unit
    /// 3. Checking its gates
    ///
    /// Currently the "run" step is a placeholder that marks the unit as
    /// completed. Subclasses or callers can override the run behavior by
    /// providing a callback via [`run_with`].
    pub fn run(&mut self) -> Result<Vec<ExecResult>, ExecError> {
        self.run_with(|unit| {
            // Default: mark as completed (placeholder).
            unit.complete();
            Ok(())
        })
    }

    /// Run all units with a custom execution callback.
    ///
    /// The callback receives a mutable reference to each [`DagUnit`] in
    /// topological order. It should perform the unit's work and update
    /// the unit's status accordingly.
    pub fn run_with<F>(&mut self, mut execute: F) -> Result<Vec<ExecResult>, ExecError>
    where
        F: FnMut(&mut DagUnit) -> Result<(), String>,
    {
        let mut results = Vec::with_capacity(self.order.len());

        for unit_id in &self.order {
            let idx = *self.index.get(unit_id).expect("unit must be in index");
            let unit = &mut self.manifest.units[idx];

            // 1. Check prerequisites.
            for prereq in &unit.pre_reqs {
                if !check_prereq(prereq) {
                    unit.status = DagUnitStatus::Skipped;
                    return Err(ExecError::PrerequisiteFailed {
                        unit: unit.id.clone(),
                        detail: format!("prerequisite {:?} not satisfied", prereq),
                    });
                }
            }

            unit.status = DagUnitStatus::InProgress;

            // 2. Execute the unit.
            match execute(unit) {
                Ok(()) => {
                    // 3. Check gates.
                    for gate in &unit.gates {
                        if !check_gate(gate) {
                            unit.status = DagUnitStatus::Failed;
                            return Err(ExecError::GateFailed {
                                unit: unit.id.clone(),
                                detail: format!("gate {:?} not satisfied", gate),
                            });
                        }
                    }
                    unit.status = DagUnitStatus::Completed;
                    results.push(ExecResult {
                        unit: unit.clone(),
                        success: true,
                        error: None,
                    });
                }
                Err(e) => {
                    unit.status = DagUnitStatus::Failed;
                    results.push(ExecResult {
                        unit: unit.clone(),
                        success: false,
                        error: Some(e),
                    });
                    // Stop execution on first failure.
                    return Ok(results);
                }
            }
        }

        Ok(results)
    }
}

// ---------------------------------------------------------------------------
// Prerequisite / gate check helpers (stub implementations)
// ---------------------------------------------------------------------------

/// Check whether a prerequisite is satisfied.
///
/// This is a stub implementation. Real checks would probe the environment,
/// call APIs, etc.
fn check_prereq(prereq: &crate::unit::PreReq) -> bool {
    match prereq {
        crate::unit::PreReq::UnitCompleted { .. } => {
            // In a real executor, look up the unit's status.
            true
        }
        crate::unit::PreReq::Command { command } => {
            // Stub: commands are "satisfied" for now.
            tracing::warn!(
                "stub: prerequisite command `{}` — treating as satisfied",
                command
            );
            true
        }
        crate::unit::PreReq::EnvVar { variable } => {
            std::env::var(variable).is_ok()
        }
        crate::unit::PreReq::ApiHealthy { url } => {
            // Stub: APIs are "healthy" for now.
            tracing::warn!("stub: API health check for `{}` — treating as healthy", url);
            true
        }
    }
}

/// Check whether a gate passes.
///
/// This is a stub implementation. Real checks would verify outputs, statuses,
/// etc.
fn check_gate(gate: &crate::unit::Gate) -> bool {
    match gate {
        crate::unit::Gate::ExitCode { code } => {
            // Stub: assume the process exited with expected code.
            *code == 0
        }
        crate::unit::Gate::OutputContains { pattern } => {
            // Stub: assume output contains the pattern.
            tracing::warn!(
                "stub: output-contains gate for pattern `{}` — treating as passed",
                pattern
            );
            true
        }
        crate::unit::Gate::HttpOk { url } => {
            // Stub: assume the endpoint is OK.
            tracing::warn!("stub: HTTP OK gate for `{}` — treating as passed", url);
            true
        }
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn sample_manifest() -> &'static str {
        r#"
version: "1.0.0"
units:
  - id: F1
    title: Complete bp-dag-foundation work on BytePort
    epic: epic_F
    repo: KooshaPari/BytePort
    gates:
      - type: exit_code
        code: 0
  - id: F2
    title: Extract reusable DAG logic
    epic: epic_F
    repo: KooshaPari/BytePort
    depends_on: [F1]
    pre_reqs:
      - type: unit_completed
        unit_id: F1
    gates:
      - type: exit_code
        code: 0
  - id: F3
    title: Test the executor
    epic: epic_F
    repo: KooshaPari/BytePort
    depends_on: [F2]
"#
    }

    #[test]
    fn executor_from_yaml() {
        let exec = Executor::from_yaml(sample_manifest()).unwrap();
        assert_eq!(exec.execution_order().len(), 3);
    }

    #[test]
    fn execution_order_is_topological() {
        let exec = Executor::from_yaml(sample_manifest()).unwrap();
        let order = exec.execution_order();
        // F1 must come before F2, F2 before F3
        let pos = |id: &str| order.iter().position(|x| x == id).unwrap();
        assert!(pos("F1") < pos("F2"));
        assert!(pos("F2") < pos("F3"));
    }

    #[test]
    fn run_all_units() {
        let mut exec = Executor::from_yaml(sample_manifest()).unwrap();
        let results = exec.run().unwrap();
        assert_eq!(results.len(), 3);
        for r in &results {
            assert!(r.success, "unit {} should succeed", r.unit.id);
        }
    }

    #[test]
    fn missing_dependency_errors() {
        let yaml = r#"
version: "1.0.0"
units:
  - id: F2
    title: Broken
    epic: epic_F
    repo: KooshaPari/BytePort
    depends_on: [F1]
"#;
        let err = Executor::from_yaml(yaml).unwrap_err();
        assert!(matches!(err, ExecError::MissingDependency { .. }));
    }

    #[test]
    fn cycle_detection() {
        let yaml = r#"
version: "1.0.0"
units:
  - id: A
    title: A
    epic: epic_F
    repo: test
    depends_on: [B]
  - id: B
    title: B
    epic: epic_F
    repo: test
    depends_on: [A]
"#;
        let err = Executor::from_yaml(yaml).unwrap_err();
        assert!(matches!(err, ExecError::Cycle(_)));
    }

    #[test]
    fn custom_execution_callback() {
        let mut exec = Executor::from_yaml(sample_manifest()).unwrap();

        let results = exec
            .run_with(|unit| {
                // Custom logic: fail F2.
                if unit.id == "F2" {
                    unit.fail();
                    return Err("F2 intentionally failed".into());
                }
                unit.complete();
                Ok(())
            })
            .unwrap();

        assert_eq!(results.len(), 2); // Stops at F2
        assert!(results[0].success); // F1
        assert!(!results[1].success); // F2
        assert_eq!(results[1].error.as_deref(), Some("F2 intentionally failed"));
    }

    #[test]
    fn env_var_prerequisite() {
        // Set an env var for this test.
        unsafe { std::env::set_var("PHENO_TEST_VAR", "present") };

        let yaml = r#"
version: "1.0.0"
units:
  - id: T1
    title: Env test
    epic: epic_F
    repo: test
    pre_reqs:
      - type: env_var
        variable: PHENO_TEST_VAR
"#;
        let mut exec = Executor::from_yaml(yaml).unwrap();
        let results = exec.run().unwrap();
        assert!(results[0].success);

        unsafe { std::env::remove_var("PHENO_TEST_VAR") };

        // Now it should fail.
        let mut exec = Executor::from_yaml(yaml).unwrap();
        let err = exec.run().unwrap_err();
        assert!(matches!(err, ExecError::PrerequisiteFailed { .. }));
    }
}
