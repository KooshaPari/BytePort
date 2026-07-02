//! # BytePort → Phenotype Registry Adapter
//!
//! This crate provides an adapter that invokes the
//! [phenotype-registry](https://github.com/KooshaPari/phenotype-registry)
//! `grade.sh` script via subprocess, parses the generated JSON grade
//! report, and exposes the results as typed Rust structs.
//!
//! ## Usage
//!
//! ```rust,no_run
//! use std::path::Path;
//! use byteport_registry_adapter::{run_grade, print_summary};
//!
//! let grade_sh = Path::new("../phenotype-registry/grade.sh");
//! let project  = Path::new(".");
//!
//! match run_grade(grade_sh, project) {
//!     Ok(report) => {
//!         println!("Grade: {} ({}%)", report.grade, report.percentage);
//!         print_summary(&report);
//!     }
//!     Err(e) => eprintln!("Grading failed: {e}"),
//! }
//! ```

use std::path::Path;
use std::process::Command;

use serde::Deserialize;

// ---------------------------------------------------------------------------
// Data types
// ---------------------------------------------------------------------------

/// A single check result from the grading pipeline.
#[derive(Debug, Clone, Deserialize)]
pub struct CheckResult {
    /// Check name (e.g. "build", "test-unit", "clippy").
    pub name: String,
    /// Outcome: `"pass"`, `"fail"`, or `"skipped"`.
    pub status: String,
    /// Points earned for this check.
    pub score: u32,
    /// Maximum points available.
    pub max: u32,
    /// Human-readable detail (usually empty on pass).
    pub detail: String,
}

/// The full grade report produced by `grade.sh --json`.
#[derive(Debug, Clone, Deserialize)]
pub struct GradeReport {
    /// Repository / project name.
    pub project: String,
    /// Detected technology stack (e.g. `"rust"`, `"node"`, `"python"`).
    pub stack: String,
    /// Grading mode: `"fast"` or `"full"`.
    pub mode: String,
    /// Total points earned.
    pub score: u32,
    /// Maximum possible points.
    pub max: u32,
    /// Score as a percentage (0–100).
    pub percentage: u32,
    /// Letter grade (`A+` through `F`).
    pub grade: String,
    /// Per-check breakdown.
    pub checks: Vec<CheckResult>,
    /// ISO-8601 timestamp of when the grade was computed.
    pub timestamp: String,
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

/// Errors that can occur while invoking `grade.sh` or parsing its output.
#[derive(Debug)]
pub enum AdapterError {
    /// Wrapped I/O error.
    Io(std::io::Error),
    /// JSON deserialization error.
    Json(serde_json::Error),
    /// The subprocess itself failed to start or returned a non-zero exit.
    Subprocess(String),
    /// The expected JSON report file was not found.
    ReportNotFound(String),
}

impl std::fmt::Display for AdapterError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            AdapterError::Io(e) => write!(f, "I/O error: {e}"),
            AdapterError::Json(e) => write!(f, "JSON parse error: {e}"),
            AdapterError::Subprocess(s) => write!(f, "subprocess error: {s}"),
            AdapterError::ReportNotFound(p) => write!(f, "report not found: {p}"),
        }
    }
}

impl std::error::Error for AdapterError {}

impl From<std::io::Error> for AdapterError {
    fn from(e: std::io::Error) -> Self {
        AdapterError::Io(e)
    }
}

impl From<serde_json::Error> for AdapterError {
    fn from(e: serde_json::Error) -> Self {
        AdapterError::Json(e)
    }
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Run `grade.sh --json` in the given project directory and return the parsed
/// [`GradeReport`].
///
/// # Arguments
///
/// * `grade_sh_path` — Absolute or relative filesystem path to the
///   `grade.sh` script (typically
///   `"../phenotype-registry/grade.sh"` when the repos are siblings).
/// * `project_dir`   — The working directory to run `grade.sh` in
///   (the root of the project being graded, e.g. the BytePort workspace root).
///
/// # Errors
///
/// Returns [`AdapterError::Subprocess`] if the script cannot be started or
/// crashes, [`AdapterError::ReportNotFound`] if the JSON output file is
/// missing, or [`AdapterError::Json`] if the JSON cannot be deserialized.
pub fn run_grade(grade_sh_path: &Path, project_dir: &Path) -> Result<GradeReport, AdapterError> {
    // Invoke grade.sh with --json flag.
    let output = Command::new(grade_sh_path)
        .arg("--json")
        .current_dir(project_dir)
        .output()
        .map_err(|e| AdapterError::Subprocess(format!("failed to execute grade.sh: {e}")))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        let stdout = String::from_utf8_lossy(&output.stdout);
        // grade.sh may still have written .grade-reports/grade.json before
        // exiting with failure (e.g. because the percentage was < 85 %).
        eprintln!(
            "grade.sh exited with status {}: stdout={} stderr={}",
            output.status, stdout, stderr
        );
    }

    // Read the JSON report from .grade-reports/grade.json.
    let report_path = project_dir.join(".grade-reports").join("grade.json");
    if !report_path.exists() {
        return Err(AdapterError::ReportNotFound(report_path.to_string_lossy().to_string()));
    }

    let json_content = std::fs::read_to_string(&report_path)?;
    let report: GradeReport = serde_json::from_str(&json_content)?;

    Ok(report)
}

/// Print a human-readable grade summary to stdout.
pub fn print_summary(report: &GradeReport) {
    println!("========================================");
    println!("  Project:   {}", report.project);
    println!("  Stack:     {}", report.stack);
    println!("  Mode:      {}", report.mode);
    println!(
        "  Score:     {} / {} ({}%)",
        report.score, report.max, report.percentage
    );
    println!("  Grade:     {}", report.grade);
    println!("  Timestamp: {}", report.timestamp);
    println!("----------------------------------------");
    println!("  Checks:");
    for check in &report.checks {
        let status_symbol = match check.status.as_str() {
            "pass" => "PASS",
            "fail" => "FAIL",
            "skipped" => "SKIP",
            _ => "?",
        };
        println!(
            "    [{status_symbol}] {name:30} {score:>3}/{max:<3}",
            name = check.name,
            score = check.score,
            max = check.max,
        );
        if !check.detail.is_empty() {
            println!("           detail: {}", check.detail);
        }
    }
    println!("========================================");
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn test_deserialize_grade_report() {
        let json_data = json!({
            "project": "BytePort",
            "stack": "rust",
            "mode": "full",
            "score": 9,
            "max": 10,
            "percentage": 90,
            "grade": "A",
            "checks": [
                {"name": "build", "status": "pass", "score": 2, "max": 2, "detail": ""},
                {"name": "test-unit", "status": "fail", "score": 0, "max": 3, "detail": "test assertion failed"}
            ],
            "timestamp": "2026-06-29T12:00:00Z"
        });

        let report: GradeReport = serde_json::from_value(json_data).unwrap();
        assert_eq!(report.project, "BytePort");
        assert_eq!(report.stack, "rust");
        assert_eq!(report.grade, "A");
        assert_eq!(report.percentage, 90);
        assert_eq!(report.checks.len(), 2);
        assert_eq!(report.checks[0].status, "pass");
        assert_eq!(report.checks[0].score, 2);
        assert_eq!(report.checks[1].status, "fail");
        assert_eq!(report.checks[1].detail, "test assertion failed");
    }

    #[test]
    fn test_print_summary_does_not_panic() {
        let report = GradeReport {
            project: "test-project".into(),
            stack: "rust".into(),
            mode: "fast".into(),
            score: 5,
            max: 10,
            percentage: 50,
            grade: "F".into(),
            checks: vec![
                CheckResult {
                    name: "build".into(),
                    status: "pass".into(),
                    score: 2,
                    max: 2,
                    detail: "".into(),
                },
                CheckResult {
                    name: "test-unit".into(),
                    status: "fail".into(),
                    score: 0,
                    max: 3,
                    detail: "timeout".into(),
                },
                CheckResult {
                    name: "coverage".into(),
                    status: "skipped".into(),
                    score: 0,
                    max: 2,
                    detail: "skipped in fast mode".into(),
                },
            ],
            timestamp: "2026-06-29T12:00:00Z".into(),
        };
        // Smoke test: just verify it runs without panicking.
        print_summary(&report);
    }
}
