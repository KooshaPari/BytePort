//! # pheno-dag
//!
//! Reusable DAG unit abstraction for Phenotype compute/infra automation
//! (epic F — DAG foundation + automation).
//!
//! ## Modules
//!
//! | Module         | Description                                        |
//! |----------------|----------------------------------------------------|
//! | [`unit`]       | `DagUnit` struct — id, title, epic, repo, status   |
//! | [`executor`]   | YAML-based executor — read, resolve deps, run      |
//!
//! ## DagUnit concepts
//!
//! A **DagUnit** represents a single unit of automation work within a larger
//! epic. Each unit carries:
//!
//! - An identifier and title
//! - The epic and repo it belongs to
//! - A lifecycle status
//! - Prerequisites (pre-conditions that must be satisfied before running)
//! - Gates (post-conditions that must pass after running)
//! - Dependencies on other DagUnits (expressed as a DAG)
//!
//! The [`executor`] module reads a YAML file describing a collection of
//! DagUnits, resolves their dependency graph using `byteport-dag`, topologically
//! sorts them, and executes them in order.

pub mod unit;
pub mod executor;
