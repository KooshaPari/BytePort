//! Integration test: `byteport-transport` against a real S3-compatible store.
//!
//! Gate: this test only runs when `LOCALSTACK_URL` is set in the
//! environment. In CI (or in any environment without an S3-compatible
//! fixture) the test prints a skip message and returns. Operators opt in
//! by running:
//!
//! ```bash
//! # Start LocalStack. The dummy credentials map to the LocalStack
//! # in-process `test`/`test` pair, which is accepted by every S3 op.
//! docker run -d --rm --name byteport-localstack \
//!     -p 4566:4566 \
//!     -e SERVICES=s3 \
//!     -e DEFAULT_REGION=us-east-1 \
//!     localstack/localstack:3.8
//!
//! # Create the test bucket once.
//! docker exec byteport-localstack \
//!     awslocal s3 mb s3://byteport-test
//!
//! LOCALSTACK_URL=http://localhost:4566 \
//!     cargo test -p byteport-transport --test presign_smoke -- --nocapture
//! ```
//!
//! What the test proves:
//!
//! 1. `S3UploadTransport::with_credentials` emits a presigned PUT URL
//!    that carries every mandatory SigV4 query parameter, including
//!    `X-Amz-Signature=`.
//! 2. The URL is accepted by a real S3-compatible store: a 1-byte PUT
//!    over `ureq` returns `200` or `204`.
//! 3. The audit's defect (unsigned URLs being silently rejected by real
//!    S3 endpoints) is closed end to end — not only at the URL-shape
//!    level.

use byteport_transport::{S3UploadTransport, UploadRequest, UploadTransport};

/// End-to-end SigV4 presigned PUT against a live S3-compatible store.
///
/// The transport is built with the LocalStack dummy credentials
/// (`test`/`test`); LocalStack accepts them for the lifetime of the
/// container. A 1-byte `PUT` to the presigned URL must return `200` or
/// `204`.
#[test]
fn s3_presign_full_roundtrip_against_s3_compatible() {
    let Some(endpoint) = std::env::var("LOCALSTACK_URL").ok() else {
        eprintln!("skipping presign smoke: LOCALSTACK_URL is not set");
        return;
    };

    let bucket = "byteport-test";
    let object_key = "demo/screenshot.png";

    // 1. Build a signed transport pointing at LocalStack.
    let transport = S3UploadTransport::with_credentials(&endpoint, bucket, Some(""), "test", "test", "us-east-1", None);

    // 2. Presign a single PUT for a 1-byte `image/png` body.
    let instruction = transport
        .create_upload(&UploadRequest {
            object_key: format!("/{object_key}"),
            content_type: "image/png".into(),
            content_length: 1,
        })
        .expect("presign should succeed against LOCALSTACK_URL endpoint");

    // 3. Check the URL shape.
    assert_eq!(instruction.method, "PUT", "presigned instructions must be PUT");

    let url = &instruction.url;
    assert!(
        url.contains("X-Amz-Signature="),
        "presigned URL must carry SigV4 signature; got: {url}"
    );

    // All mandatory SigV4 query parameters must be present. An unsigned
    // URL (the pre-fix defect) would fail at least one of these.
    for param in [
        "X-Amz-Algorithm=AWS4-HMAC-SHA256",
        "X-Amz-Credential=test",
        "X-Amz-Date=",
        "X-Amz-Expires=",
        "X-Amz-SignedHeaders=",
        "X-Amz-Signature=",
    ] {
        assert!(url.contains(param), "missing presign query param {param:?} in {url}");
    }

    // 4. PUT 1 byte to the signed URL with `ureq`. LocalStack returns
    //    `200` on a successful object create. AWS S3 returns `200` too.
    let body: &[u8] = b"x";
    let response = ureq::put(url)
        .set("content-type", "image/png")
        .set("content-length", "1")
        .send_bytes(body);

    let status = match response {
        Ok(r) => r.status(),
        Err(ureq::Error::Status(code, response)) => {
            // Pull the body into a `String` for diagnostics so the failure
            // mode is obvious in CI logs.
            let mut buf = Vec::new();
            let _ = std::io::Read::read_to_end(&mut response.into_reader(), &mut buf);
            panic!(
                "S3-compatible store returned HTTP {code} for presigned PUT; body: {}",
                String::from_utf8_lossy(&buf)
            );
        }
        Err(other) => panic!("presigned PUT transport error: {other:?}"),
    };

    assert!(
        status == 200 || status == 204,
        "expected 200 or 204 from S3-compatible store; got {status}"
    );

    // 5. Diagnostic line so the test output shows the round-trip succeeded
    //    even when `--quiet` is used.
    eprintln!(
        "presign smoke ok: PUT {n}-byte body to {endpoint}/{bucket}/{key} returned {status}",
        n = body.len(),
        endpoint = endpoint.trim_end_matches('/'),
        bucket = bucket,
        key = object_key
    );
}
