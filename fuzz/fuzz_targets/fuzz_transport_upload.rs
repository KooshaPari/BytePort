#![no_main]

use libfuzzer_sys::fuzz_target;

use byteport_transport::{UploadRequest, UploadTransport};

fuzz_target!(|data: &[u8]| {
    // Feed arbitrary bytes as a potential JSON-serialized UploadRequest.
    // If deserialization succeeds, we exercise the entire UploadTransport
    // pipeline — this catches panics, logical errors, and edge cases in
    // the S3UploadTransport::create_upload() codepath.
    if let Ok(request) = serde_json::from_slice::<UploadRequest>(data) {
        // Construct a minimal S3 transport with arbitrary configuration.
        let transport = byteport_transport::S3UploadTransport::new(
            "https://fuzz.example.com",
            "fuzz-bucket",
            Some("/fuzz"),
        );

        // Exercise the upload creation codepath with the deserialized input.
        let _result = transport.create_upload(&request);
    }
});
