//! # byteport-dag
//!
//! DAG (Directed Acyclic Graph) foundation for BytePort compute/infra orchestration.
//!
//! This crate provides the core types for modelling deployment/infrastructure
//! workflows as an acyclic graph of **nodes** (tasks/steps) and **edges**
//! (dependencies). The DAG engine validates ordering, detects cycles at
//! insertion time, and produces a topological execution order.

use std::collections::{BTreeMap, BTreeSet};

use serde::{Deserialize, Serialize};
use thiserror::Error;

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

/// Unique identifier for a DAG node.
pub type NodeId = String;

/// Opaque payload attached to a node. Consumers may attach any serializable
/// value (e.g. a deployment step, a build command, a config template).
pub type Payload = serde_json::Value;

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

/// Errors returned by DAG construction and validation.
#[derive(Debug, Error, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum DagError {
    /// A node with the same ID already exists.
    #[error("node '{0}' already exists")]
    DuplicateNode(NodeId),

    /// An edge references a source or target that does not exist.
    #[error("node '{0}' not found")]
    NodeNotFound(NodeId),

    /// Adding this edge would introduce a cycle.
    #[error("adding edge {0} -> {1} would create a cycle")]
    CycleDetected(NodeId, NodeId),

    /// The DAG is empty (no nodes).
    #[error("dag has no nodes")]
    EmptyDag,
}

// ---------------------------------------------------------------------------
// Core types
// ---------------------------------------------------------------------------

/// A single node in the DAG.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    /// Unique identifier.
    pub id: NodeId,
    /// Human-readable label (optional).
    pub label: Option<String>,
    /// Arbitrary payload carried by this node.
    pub payload: Option<Payload>,
    /// Set of node IDs that must execute **before** this node.
    pub dependencies: BTreeSet<NodeId>,
}

/// The DAG itself.
///
/// # Invariants
///
/// - No duplicate node IDs.
/// - Every dependency reference (edge) points to an existing node.
/// - The graph is acyclic (enforced at edge insertion time).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Dag {
    nodes: BTreeMap<NodeId, Node>,
}

impl Dag {
    /// Create an empty DAG.
    pub fn new() -> Self {
        Self {
            nodes: BTreeMap::new(),
        }
    }

    /// Create a DAG with a pre-built node map.  **Does not validate** –
    /// use [`Self::validate`] if constructing from an external source.
    pub fn from_raw(nodes: BTreeMap<NodeId, Node>) -> Self {
        Self { nodes }
    }

    /// Add a node with no dependencies.
    ///
    /// Returns `DuplicateNode` if the ID is already registered.
    pub fn add_node(
        &mut self,
        id: impl Into<NodeId>,
        label: Option<String>,
        payload: Option<Payload>,
    ) -> Result<(), DagError> {
        let id = id.into();
        if self.nodes.contains_key(&id) {
            return Err(DagError::DuplicateNode(id));
        }
        self.nodes.insert(
            id,
            Node {
                id: id.clone(),
                label,
                payload,
                dependencies: BTreeSet::new(),
            },
        );
        Ok(())
    }

    /// Add a dependency edge (`from -> to`), meaning "`to` depends on `from`".
    ///
    /// Both endpoints must exist.  Insertion fails if the edge would form a
    /// cycle (detected via DFS).
    pub fn add_edge(&mut self, from: &NodeId, to: &NodeId) -> Result<(), DagError> {
        if !self.nodes.contains_key(from) {
            return Err(DagError::NodeNotFound(from.clone()));
        }
        if !self.nodes.contains_key(to) {
            return Err(DagError::NodeNotFound(to.clone()));
        }

        // Provisional insert + cycle check.
        self.nodes
            .get_mut(to)
            .expect("verified above")
            .dependencies
            .insert(from.clone());

        if self.detect_cycle(to) {
            // Revert
            self.nodes
                .get_mut(to)
                .expect("verified above")
                .dependencies
                .remove(from);
            return Err(DagError::CycleDetected(from.clone(), to.clone()));
        }

        Ok(())
    }

    /// Return a topological ordering of all nodes, or `EmptyDag` / error if a
    /// cycle somehow snuck in.
    pub fn topological_sort(&self) -> Result<Vec<&Node>, DagError> {
        if self.nodes.is_empty() {
            return Err(DagError::EmptyDag);
        }

        // Kahn's algorithm.
        let mut in_degree: BTreeMap<&NodeId, usize> =
            self.nodes.keys().map(|id| (id, 0)).collect();
        for node in self.nodes.values() {
            for dep in &node.dependencies {
                *in_degree.entry(dep).or_insert(0) += 0; // just ensure entry
                // increment in-degree of the dependant
            }
        }
        // Recalculate properly
        let mut in_degree = BTreeMap::new();
        for id in self.nodes.keys() {
            in_degree.insert(id.as_str(), 0usize);
        }
        for node in self.nodes.values() {
            for dep in &node.dependencies {
                // "dep" is a dependency of "node", so node depends on dep.
                // In topological ordering, dep comes before node.
                // We need to track how many dependencies each node has.
            }
        }
        // Correct approach: for each node, count how many nodes depend on it
        // Actually for Kahn's: in_degree = number of incoming edges = number of dependencies
        for node in self.nodes.values() {
            in_degree.insert(node.id.as_str(), node.dependencies.len());
        }

        let mut queue: Vec<&NodeId> = in_degree
            .iter()
            .filter(|(_, &deg)| deg == 0)
            .map(|(id, _)| *id)
            .collect();

        let mut sorted: Vec<&Node> = Vec::with_capacity(self.nodes.len());

        while let Some(id) = queue.pop() {
            // Find all nodes that depend on `id`
            for node in self.nodes.values() {
                if node.dependencies.contains(id) {
                    let entry = in_degree.get_mut(node.id.as_str()).expect("exists");
                    *entry = entry.saturating_sub(1);
                    if *entry == 0 {
                        queue.push(&node.id);
                    }
                }
            }
            sorted.push(self.nodes.get(id).expect("exists"));
        }

        if sorted.len() != self.nodes.len() {
            // Cycle detected (shouldn't happen if edges were validated)
            // Fallback: detect cycle and return error
            return Err(DagError::CycleDetected("unknown".into(), "unknown".into()));
        }

        Ok(sorted)
    }

    /// Validate graph invariants.
    pub fn validate(&self) -> Result<(), DagError> {
        if self.nodes.is_empty() {
            return Err(DagError::EmptyDag);
        }

        // All dependency references must point to existing nodes.
        for node in self.nodes.values() {
            for dep in &node.dependencies {
                if !self.nodes.contains_key(dep) {
                    return Err(DagError::NodeNotFound(dep.clone()));
                }
            }
        }

        // Cycle check via DFS (iterative).
        let mut visited: BTreeSet<&NodeId> = BTreeSet::new();
        let mut stack: BTreeSet<&NodeId> = BTreeSet::new();

        for id in self.nodes.keys() {
            if self._detect_cycle_from(id, &mut visited, &mut stack) {
                return Err(DagError::CycleDetected("unknown".into(), "unknown".into()));
            }
        }

        Ok(())
    }

    /// Number of nodes in the DAG.
    pub fn len(&self) -> usize {
        self.nodes.len()
    }

    /// Returns true if the DAG has no nodes.
    pub fn is_empty(&self) -> bool {
        self.nodes.is_empty()
    }

    /// Iterate over all nodes.
    pub fn iter(&self) -> impl Iterator<Item = &Node> {
        self.nodes.values()
    }

    // -----------------------------------------------------------------------
    // Private helpers
    // -----------------------------------------------------------------------

    /// DFS cycle detection starting from a single node.
    fn detect_cycle(&self, start: &NodeId) -> bool {
        let mut visited: BTreeSet<&NodeId> = BTreeSet::new();
        let mut stack: BTreeSet<&NodeId> = BTreeSet::new();
        self._detect_cycle_from(start, &mut visited, &mut stack)
    }

    fn _detect_cycle_from<'a>(
        &'a self,
        id: &'a NodeId,
        visited: &mut BTreeSet<&'a NodeId>,
        stack: &mut BTreeSet<&'a NodeId>,
    ) -> bool {
        if stack.contains(id) {
            return true;
        }
        if visited.contains(id) {
            return false;
        }
        visited.insert(id);
        stack.insert(id);

        if let Some(node) = self.nodes.get(id) {
            for dep in &node.dependencies {
                if self._detect_cycle_from(dep, visited, stack) {
                    return true;
                }
            }
        }

        stack.remove(id);
        false
    }
}

impl Default for Dag {
    fn default() -> Self {
        Self::new()
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn empty_dag_validation_fails() {
        let dag = Dag::new();
        assert_eq!(dag.validate(), Err(DagError::EmptyDag));
    }

    #[test]
    fn add_single_node() {
        let mut dag = Dag::new();
        dag.add_node("build", Some("Build step".into()), None)
            .unwrap();
        assert_eq!(dag.len(), 1);
    }

    #[test]
    fn duplicate_node_is_rejected() {
        let mut dag = Dag::new();
        dag.add_node("build", None, None).unwrap();
        assert_eq!(
            dag.add_node("build", None, None),
            Err(DagError::DuplicateNode("build".into()))
        );
    }

    #[test]
    fn edge_to_nonexistent_node_fails() {
        let mut dag = Dag::new();
        dag.add_node("build", None, None).unwrap();
        assert_eq!(
            dag.add_edge("phantom", "build"),
            Err(DagError::NodeNotFound("phantom".into()))
        );
    }

    #[test]
    fn cycle_detection_on_simple_cycle() {
        let mut dag = Dag::new();
        dag.add_node("a", None, None).unwrap();
        dag.add_node("b", None, None).unwrap();
        dag.add_edge("a", "b").unwrap();
        assert_eq!(
            dag.add_edge("b", "a"),
            Err(DagError::CycleDetected("b".into(), "a".into()))
        );
    }

    #[test]
    fn topological_sort_linear() {
        let mut dag = Dag::new();
        dag.add_node("build", None, None).unwrap();
        dag.add_node("test", None, None).unwrap();
        dag.add_node("deploy", None, None).unwrap();
        dag.add_edge("build", "test").unwrap();
        dag.add_edge("test", "deploy").unwrap();

        let order = dag.topological_sort().unwrap();
        let ids: Vec<&str> = order.iter().map(|n| n.id.as_str()).collect();
        // build must come before test, test before deploy
        let pos_build = ids.iter().position(|&i| i == "build").unwrap();
        let pos_test = ids.iter().position(|&i| i == "test").unwrap();
        let pos_deploy = ids.iter().position(|&i| i == "deploy").unwrap();
        assert!(pos_build < pos_test);
        assert!(pos_test < pos_deploy);
    }

    #[test]
    fn topological_sort_diamond() {
        let mut dag = Dag::new();
        dag.add_node("root", None, None).unwrap();
        dag.add_node("left", None, None).unwrap();
        dag.add_node("right", None, None).unwrap();
        dag.add_node("merge", None, None).unwrap();

        dag.add_edge("root", "left").unwrap();
        dag.add_edge("root", "right").unwrap();
        dag.add_edge("left", "merge").unwrap();
        dag.add_edge("right", "merge").unwrap();

        let order = dag.topological_sort().unwrap();
        let ids: Vec<&str> = order.iter().map(|n| n.id.as_str()).collect();
        // root must be first
        assert_eq!(ids[0], "root");
        // merge must be last
        assert_eq!(ids[ids.len() - 1], "merge");
    }

    #[test]
    fn validate_passes_for_acyclic_graph() {
        let mut dag = Dag::new();
        dag.add_node("a", None, None).unwrap();
        dag.add_node("b", None, None).unwrap();
        dag.add_edge("a", "b").unwrap();
        assert!(dag.validate().is_ok());
    }

    #[test]
    fn validate_fails_for_cycle() {
        let mut nodes = BTreeMap::new();
        let a_id = "a".to_string();
        let b_id = "b".to_string();
        let mut a_deps = BTreeSet::new();
        a_deps.insert("b".to_string());
        let mut b_deps = BTreeSet::new();
        b_deps.insert("a".to_string());

        nodes.insert(
            a_id.clone(),
            Node {
                id: a_id,
                label: None,
                payload: None,
                dependencies: a_deps,
            },
        );
        nodes.insert(
            b_id.clone(),
            Node {
                id: b_id,
                label: None,
                payload: None,
                dependencies: b_deps,
            },
        );

        let dag = Dag::from_raw(nodes);
        assert_eq!(
            dag.validate(),
            Err(DagError::CycleDetected("unknown".into(), "unknown".into()))
        );
    }
}
