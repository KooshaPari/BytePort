#![no_main]

use libfuzzer_sys::fuzz_target;

use byteport_dag::serialize::DagSchema;

fuzz_target!(|data: &[u8]| {
    // Interpret the fuzz input as a UTF-8 YAML string (lossy conversion
    // handles arbitrary bytes gracefully).
    if let Ok(yaml_str) = std::str::from_utf8(data) {
        // Attempt to deserialize a DagSchema from the (potentially
        // malformed) YAML.  serde_yaml has its own internal recursion
        // limits and error handling, but we still exercise the full
        // deserialization codepath to catch panics, stack overflows,
        // or logic errors in the schema types.
        if let Ok(schema) = DagSchema::from_yaml(yaml_str) {
            // Round-trip back through YAML to catch serialization panics.
            let _round = schema.to_yaml();

            // Round-trip through JSON for cross-format consistency coverage.
            if let Ok(json) = schema.to_json() {
                let _from_json = DagSchema::from_json(&json);
            }
        }
    }
});
