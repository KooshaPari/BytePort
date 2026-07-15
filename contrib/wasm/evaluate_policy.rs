// BytePort WASM Component Model — Standalone Policy Evaluator
//
// This component compiles to wasm32-unknown-unknown and evaluates
// JSON-based deployment policies using only serde_json + std.
// No tokio / async dependencies required.
//
// Build:   cargo build --target wasm32-unknown-unknown --release
// Verify:  wasmtime wasm32-unknown-unknown/release/byteport_wasm_policy.wasm

use std::collections::HashMap;

use serde_json::Value;

// ── Simple policy engine (no byteport-engine dep needed) ──────────────

/// A PolicyRule is a single condition-action pair.
#[derive(serde::Deserialize)]
struct PolicyRule {
    /// JSONPath-like condition, e.g. "$.memory_gb > 4"
    condition: String,
    /// Effect: "allow" | "deny" | "warn"
    effect: String,
    /// Human-readable reason
    #[serde(default)]
    reason: String,
}

/// A Policy is a named collection of rules.
#[derive(serde::Deserialize)]
struct Policy {
    name: String,
    rules: Vec<PolicyRule>,
}

/// Evaluation result returned to the host.
#[derive(serde::Serialize)]
struct EvaluationOutput {
    policy: String,
    allowed: bool,
    reasons: Vec<String>,
    violations: Vec<String>,
}

/// Simple condition evaluator for expressions like `"$.memory_gb > 4"`.
///
/// Supported operators: `>`, `<`, `>=`, `<=`, `==`, `!=`.
/// Left-hand side is a JSON path prefixed with `$.` (e.g. `$.memory_gb`).
fn evaluate_condition(condition: &str, input: &Value) -> bool {
    // Split on whitespace: "$.field OP value"
    let parts: Vec<&str> = condition.split_whitespace().collect();
    if parts.len() != 3 {
        return false; // malformed
    }
    let path = parts[0].strip_prefix("$.").unwrap_or(parts[0]);
    let op = parts[1];
    let raw_rhs: &str = parts[2];

    let lhs = input.get(path);
    let lhs = match lhs {
        Some(v) => v,
        None => return false, // field not present
    };

    // Try numeric comparison first
    if let (Some(lhs_num), Some(rhs_num)) = (lhs.as_f64(), raw_rhs.parse::<f64>().ok()) {
        return match op {
            ">" => lhs_num > rhs_num,
            "<" => lhs_num < rhs_num,
            ">=" => lhs_num >= rhs_num,
            "<=" => lhs_num <= rhs_num,
            "==" => (lhs_num - rhs_num).abs() < f64::EPSILON,
            "!=" => (lhs_num - rhs_num).abs() >= f64::EPSILON,
            _ => false,
        };
    }

    // String comparison
    let rhs_str = raw_rhs.trim_matches('"');
    let lhs_str = lhs.as_str().unwrap_or("");
    match op {
        "==" => lhs_str == rhs_str,
        "!=" => lhs_str != rhs_str,
        _ => false,
    }
}

fn run_evaluation(policy_json: &str, input_json: &str) -> Result<String, String> {
    let policy: Policy =
        serde_json::from_str(policy_json).map_err(|e| format!("invalid policy: {e}"))?;
    let input: Value =
        serde_json::from_str(input_json).map_err(|e| format!("invalid input: {e}"))?;

    let mut reasons: Vec<String> = Vec::new();
    let mut violations: Vec<String> = Vec::new();

    for rule in &policy.rules {
        let matched = evaluate_condition(&rule.condition, &input);
        if matched {
            match rule.effect.as_str() {
                "deny" => violations.push(rule.reason.clone()),
                "warn" => reasons.push(format!("warn: {}", rule.reason)),
                _ => reasons.push(rule.reason.clone()),
            }
        }
    }

    let output = EvaluationOutput {
        policy: policy.name,
        allowed: violations.is_empty(),
        reasons,
        violations,
    };

    serde_json::to_string(&output).map_err(|e| format!("serialize: {e}"))
}

// ── C ABI entrypoint ─────────────────────────────────────────────────

#[no_mangle]
pub extern "C" fn evaluate_policy(
    policy_json: *const u8,
    policy_len: usize,
    input_json: *const u8,
    input_len: usize,
) -> *mut u8 {
    let policy_str = unsafe {
        let slice = std::slice::from_raw_parts(policy_json, policy_len);
        std::str::from_utf8_unchecked(slice)
    };

    let input_str = unsafe {
        let slice = std::slice::from_raw_parts(input_json, input_len);
        std::str::from_utf8_unchecked(slice)
    };

    let result = match run_evaluation(policy_str, input_str) {
        Ok(r) => r,
        Err(e) => format!(r#"{{"error":"{e}"}}"#),
    };
    string_to_ptr(result)
}

fn string_to_ptr(s: String) -> *mut u8 {
    let mut b = s.into_bytes();
    let ptr = b.as_mut_ptr();
    std::mem::forget(b);
    ptr
}

// ── Tests (native; run with `cargo test`) ────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_allow_low_memory() {
        let policy = r#"{
            "name": "mem-quota",
            "rules": [
                {"condition": "$.memory_gb > 4", "effect": "deny", "reason": "memory exceeds 4 GB quota"}
            ]
        }"#;
        let input = r#"{"memory_gb": 2}"#;
        let out = run_evaluation(policy, input).unwrap();
        assert!(out.contains(r#""allowed":true"#));
        assert!(out.contains(r#""violations":[]"#));
    }

    #[test]
    fn test_deny_high_memory() {
        let policy = r#"{
            "name": "mem-quota",
            "rules": [
                {"condition": "$.memory_gb > 4", "effect": "deny", "reason": "memory exceeds 4 GB quota"}
            ]
        }"#;
        let input = r#"{"memory_gb": 8}"#;
        let out = run_evaluation(policy, input).unwrap();
        assert!(out.contains(r#""allowed":false"#));
        assert!(out.contains(r#""memory exceeds 4 GB quota""#));
    }

    #[test]
    fn test_string_equality() {
        let policy = r#"{
            "name": "env-check",
            "rules": [
                {"condition": "$.env == \"production\"", "effect": "warn", "reason": "production deployment audited"}
            ]
        }"#;
        let input = r#"{"env": "production"}"#;
        let out = run_evaluation(policy, input).unwrap();
        assert!(out.contains(r#""allowed":true"#));
        assert!(out.contains(r#""warn: production deployment audited""#));
    }

    #[test]
    fn test_missing_field_no_match() {
        let policy = r#"{
            "name": "gpu-check",
            "rules": [
                {"condition": "$.gpu_count > 0", "effect": "deny", "reason": "GPU not allowed"}
            ]
        }"#;
        let input = r#"{"memory_gb": 4}"#;
        let out = run_evaluation(policy, input).unwrap();
        assert!(out.contains(r#""allowed":true"#)); // no violation since field missing
    }
}
