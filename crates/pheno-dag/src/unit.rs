//! DagUnit — the fundamental automation unit.
//!
//! Each DagUnit describes a single unit of work within an epic. Units can
//! depend on other units, forming a DAG that the executor resolves and runs
//! in topological order.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

// ---------------------------------------------------------------------------
// Status
// ---------------------------------------------------------------------------

/// Lifecycle status of a DagUnit.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum DagUnitStatus {
    /// Unit has been defined but not started.
    Pending,
    /// Unit is currently executing.
    InProgress,
    /// Unit completed successfully.
    Completed,
    /// Unit failed during execution.
    Failed,
    /// Unit was skipped (e.g. a prerequisite was not met).
    Skipped,
    /// Unit has been cancelled.
    Cancelled,
}

impl Default for DagUnitStatus {
    fn default() -> Self {
        Self::Pending
    }
}

// ---------------------------------------------------------------------------
// Prerequisite
// ---------------------------------------------------------------------------

/// A condition that must be satisfied **before** a DagUnit may execute.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum PreReq {
    /// Another DagUnit must have completed.
    UnitCompleted {
        /// ID of the unit that must have completed.
        unit_id: String,
    },
    /// A shell command must exit with code 0.
    Command {
        /// The command to run.
        command: String,
    },
    /// An environment variable must be set.
    EnvVar {
        /// Variable name.
        variable: String,
    },
    /// An HTTP endpoint must be healthy.
    ApiHealthy {
        /// URL to probe.
        url: String,
    },
}

// ---------------------------------------------------------------------------
// Gate
// ---------------------------------------------------------------------------

/// A condition that must pass **after** a DagUnit executes for the unit
/// to be considered successful.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum Gate {
    /// The process must exit with this code.
    ExitCode {
        /// Expected exit code (default 0).
        #[serde(default)]
        code: i32,
    },
    /// Output must contain this pattern.
    OutputContains {
        /// Substring or regex pattern.
        pattern: String,
    },
    /// An HTTP endpoint must respond 2xx.
    HttpOk {
        /// URL to check.
        url: String,
    },
}

// ---------------------------------------------------------------------------
// DagUnit
// ---------------------------------------------------------------------------

/// A single automation unit within an epic.
///
/// Units can declare prerequisites (pre-conditions) and gates (post-conditions).
/// Dependencies between units are expressed via [`PreReq::UnitCompleted`] or
/// via the `depends_on` field which the executor converts to DAG edges.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DagUnit {
    /// Unique identifier for this unit (e.g. `"F1"`, `"F2"`).
    pub id: String,

    /// Human-readable title.
    pub title: String,

    /// Parent epic identifier (e.g. `"epic_F"`).
    pub epic: String,

    /// GitHub repository slug (e.g. `"KooshaPari/BytePort"`).
    pub repo: String,

    /// Current lifecycle status.
    #[serde(default)]
    pub status: DagUnitStatus,

    /// Prerequisites that must be satisfied before this unit runs.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub pre_reqs: Vec<PreReq>,

    /// Gates that must pass after this unit completes.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub gates: Vec<Gate>,

    /// Other unit IDs that this unit depends on (creates DAG edges).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub depends_on: Vec<String>,

    /// Arbitrary key-value metadata.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub metadata: Option<std::collections::HashMap<String, String>>,

    /// When this unit was created.
    #[serde(default = "Utc::now")]
    pub created_at: DateTime<Utc>,

    /// Internal UUID.
    #[serde(default = "new_uuid")]
    pub uuid: Uuid,
}

fn new_uuid() -> Uuid {
    Uuid::new_v4()
}

impl DagUnit {
    /// Create a new DagUnit with the minimal required fields.
    pub fn new(
        id: impl Into<String>,
        title: impl Into<String>,
        epic: impl Into<String>,
        repo: impl Into<String>,
    ) -> Self {
        Self {
            id: id.into(),
            title: title.into(),
            epic: epic.into(),
            repo: repo.into(),
            status: DagUnitStatus::Pending,
            pre_reqs: Vec::new(),
            gates: Vec::new(),
            depends_on: Vec::new(),
            metadata: None,
            created_at: Utc::now(),
            uuid: Uuid::new_v4(),
        }
    }

    /// Add a prerequisite.
    pub fn with_prereq(mut self, prereq: PreReq) -> Self {
        self.pre_reqs.push(prereq);
        self
    }

    /// Add a gate.
    pub fn with_gate(mut self, gate: Gate) -> Self {
        self.gates.push(gate);
        self
    }

    /// Add a dependency.
    pub fn depends_on(mut self, unit_id: impl Into<String>) -> Self {
        self.depends_on.push(unit_id.into());
        self
    }

    /// Set metadata.
    pub fn with_metadata(mut self, key: impl Into<String>, value: impl Into<String>) -> Self {
        self.metadata
            .get_or_insert_with(std::collections::HashMap::new)
            .insert(key.into(), value.into());
        self
    }

    /// Mark the unit as completed.
    pub fn complete(&mut self) {
        self.status = DagUnitStatus::Completed;
    }

    /// Mark the unit as failed.
    pub fn fail(&mut self) {
        self.status = DagUnitStatus::Failed;
    }
}

// ---------------------------------------------------------------------------
// DagUnitManifest — a collection of units that can be serialized to YAML
// ---------------------------------------------------------------------------

/// A manifest file containing a set of DagUnits, typically loaded from YAML.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DagUnitManifest {
    /// Manifest schema version.
    pub version: String,
    /// The collection of units.
    pub units: Vec<DagUnit>,
}

impl DagUnitManifest {
    /// Parse a manifest from a YAML string.
    pub fn from_yaml(yaml: &str) -> Result<Self, serde_yaml::Error> {
        serde_yaml::from_str(yaml)
    }

    /// Serialize to a YAML string.
    pub fn to_yaml(&self) -> Result<String, serde_yaml::Error> {
        serde_yaml::to_string(self)
    }

    /// Find a unit by its ID.
    pub fn find(&self, id: &str) -> Option<&DagUnit> {
        self.units.iter().find(|u| u.id == id)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn create_minimal_unit() {
        let unit = DagUnit::new("F1", "Build dag crate", "epic_F", "KooshaPari/BytePort");
        assert_eq!(unit.id, "F1");
        assert_eq!(unit.status, DagUnitStatus::Pending);
    }

    #[test]
    fn unit_with_prereqs_and_gates() {
        let unit = DagUnit::new("F2", "Test", "epic_F", "KooshaPari/BytePort")
            .with_prereq(PreReq::UnitCompleted { unit_id: "F1".into() })
            .with_gate(Gate::ExitCode { code: 0 });

        assert_eq!(unit.pre_reqs.len(), 1);
        assert_eq!(unit.gates.len(), 1);
    }

    #[test]
    fn unit_with_dependencies() {
        let unit = DagUnit::new("F3", "Deploy", "epic_F", "KooshaPari/BytePort")
            .depends_on("F1")
            .depends_on("F2");

        assert_eq!(unit.depends_on.len(), 2);
    }

    #[test]
    fn manifest_yaml_round_trip() {
        let manifest = DagUnitManifest {
            version: "1.0.0".into(),
            units: vec![
                DagUnit::new("F1", "Build", "epic_F", "KooshaPari/BytePort"),
                DagUnit::new("F2", "Test", "epic_F", "KooshaPari/BytePort").depends_on("F1"),
            ],
        };

        let yaml = manifest.to_yaml().expect("serialization");
        let restored = DagUnitManifest::from_yaml(&yaml).expect("deserialization");
        assert_eq!(restored.version, "1.0.0");
        assert_eq!(restored.units.len(), 2);
        assert_eq!(restored.find("F1").unwrap().title, "Build");
        assert_eq!(restored.find("F2").unwrap().depends_on, vec!["F1"]);
    }

    #[test]
    fn unit_lifecycle() {
        let mut unit = DagUnit::new("F1", "Build", "epic_F", "KooshaPari/BytePort");
        assert_eq!(unit.status, DagUnitStatus::Pending);
        unit.complete();
        assert_eq!(unit.status, DagUnitStatus::Completed);
    }
}
