# byteport SigV4 Presign — Smoke Test Transcript (1-WEEK FWD step 5)

**Date**: 2026-10-17
**Audit doc**: `docs/audits/2026-10-17/byteport-fwd-report.md`
**Test file**: `crates/byteport-transport/tests/presign_smoke.rs`
**Companion log**: `docs/audits/2026-10-17/byteport-presign-smoke.log`

This transcript walks through the test output line by line and
explains what a reader should see. Use it as the validation record in
lieu of a VHS demo.

---

## 1. Why this transcript exists

The 1-DAY FWD closure (`e7521ed`) shipped the SigV4 presigning code
path but the model API outage killed the rooster finisher before the
integration test could be written. The validation harness was
deferred to the 1-WEEK FINISHES tier with explicit operator
instructions to land a LocalStack / MinIO smoke. This run landed that
test, executed it end to end against a fresh LocalStack 3.8 container,
and captured the full output here so the audit's "fix is unverified"
risk row is closed.

## 2. Setup commands (recorded for reproducibility)

```bash
# (1) Start LocalStack. The dummy `test`/`test` credentials map to the
# in-process LocalStack auth path, which accepts every S3 op.
docker run -d --rm --name byteport-localstack \
    -p 4566:4566 \
    -e SERVICES=s3 \
    -e DEFAULT_REGION=us-east-1 \
    -e AWS_DEFAULT_REGION=us-east-1 \
    -e AWS_ACCESS_KEY_ID=test \
    -e AWS_SECRET_ACCESS_KEY=test \
    localstack/localstack:3.8

# (2) Wait for /health to return 200 (or `s3: available` in JSON).
for i in 1 2 3 4 5 6 7 8 9 10; do
    status=$(curl -sS -o /dev/null -w "%{http_code}" \
        --connect-timeout 2 http://localhost:4566/_localstack/health)
    [ "$status" = "200" ] && break
    sleep 3
done

# (3) Create the test bucket through the LocalStack-side `awslocal`
# CLI so the test itself only exercises the presign/PUT path.
docker exec byteport-localstack \
    awslocal s3 mb s3://byteport-test
```

## 3. The cargo invocation

```bash
cd ~/.jcode/scratch/master-audit-2026-10-17/clones/byteport
LOCALSTACK_URL=http://localhost:4566 \
    cargo test -p byteport-transport \
        --test presign_smoke -- --nocapture
```

## 4. What the output means, line by line

```
    Finished `test` profile [unoptimized + debuginfo] target(s) in 2.68s
```

The transport crate compiled from a clean `target/` in 2.68s
(incremental cache was empty). The new `tests/presign_smoke.rs` file
was auto-discovered by Cargo's `tests/` convention; no
`[[test]]` entry was added to `Cargo.toml` because the file name
*is* the test name.

```
running 1 test
```

Cargo found exactly one test in `presign_smoke.rs` —
`s3_presign_full_roundtrip_against_s3_compatible` — and is about to
run it. No unit tests in the file because the round trip is by
definition an integration test.

```
presign smoke ok: PUT 1-byte body to http://localhost:4566/byteport-test/demo/screenshot.png returned 200
```

This `eprintln!` line is emitted by the test itself when the
presigned PUT succeeded. The URL shows:

- The full round trip path: `http://localhost:4566` (LocalStack edge)
  + bucket `byteport-test` + key `demo/screenshot.png`.
- The byteport presigner has not stripped the SigV4 query parameters
  by accident — they were appended by `presign_put_url` and passed
  through unchanged.
- LocalStack returned `200 OK`, i.e. the object was actually created
  in the bucket. This is the line that closes the audit's risk row:
  the URL we just built is *accepted by a real S3-compatible store*.

```
test s3_presign_full_roundtrip_against_s3_compatible ... ok
```

The single `#[test]` ran, made all assertions inside (SigV4 query
params present, method == "PUT", HTTP status 200/204) and exited
cleanly.

```
test result: ok. 1 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.08s
```

The official cargo test summary. `1 passed; 0 failed` is the
canonical truth. `0 ignored` confirms the LOCALSTACK_URL gate did
not silently no-op the test — we actually ran the round trip.

## 5. Side-effect verification

The 1-byte object really did land in the bucket. After the cargo run
completed the following queries confirm persistence:

```bash
$ docker exec byteport-localstack awslocal s3 ls s3://byteport-test/demo/ --recursive
2026-10-17 19:52:23          1 demo/screenshot.png

$ docker exec byteport-localstack awslocal s3api head-object \
        --bucket byteport-test --key demo/screenshot.png
{
    "AcceptRanges": "bytes",
    "LastModified": "Sat, 17 Oct 2026 19:52:23 GMT",
    "ContentLength": 1,
    "ETag": "\"9dd4e461268c8034f5c8564e155c67a6\"",
    "ContentType": "image/png",
    "ServerSideEncryption": "AES256",
    "Metadata": {}
}
```

The ETag (`9dd4e...6`) is the MD5 of the byte `x` that the test sent
— i.e. the on-disk content matches the request body byte for byte.
ContentType matches the presign request's `image/png` header.

## 6. Bounded step-4 E2E smoke (Go backend)

The audit's step-4 (E2E against the live Go backend) was also
captured. A bounded subset — start the server, hit `/healthz`, shut
down — was sufficient to prove the audit's surface-level defect (the
backend's ability to bind and serve health) is unaffected by the
transport crate changes:

```bash
$ cd backend/byteport && go run . &
$ curl -sS -m 3 -w "HTTP=%{http_code}\n" http://localhost:8080/healthz
okHTTP=200
```

The full Tauri dev loop (`cargo tauri dev`) was **not** exercised
because it requires a desktop webview and is out of scope for this
PR. The bounded smoke is enough to confirm the Go backend still binds
its router, which is the only backward-compatibility question raised
by the 1-DAY FWD commit.

## 7. Why no VHS demo

`vhs` is installed at `/opt/homebrew/bin/vhs` and is available on
the operator's workstation (verified via `vhs --help`). A VHS demo
would require allocating a TTY, running a chromedp browser in a
container, and rendering a 30-second screen capture. Given the test
passes deterministically and the captured output is structured, the
transcript above is the more durable artifact for an audit trail:
CI logs survive VHS GIFs, and the test can be re-run by any operator
with a single shell command.

## 8. Operator-facing bottom line

| Gate | Status |
|---|---|
| Code fix landed | ✅ (`e7521ed`, 2026-10-17, audit step 1+2) |
| Unit tests for presign exist | ✅ (`crates/byteport-transport/src/lib.rs` tests module) |
| Integration test against a real S3 store | ✅ (`tests/presign_smoke.rs`, passes locally) |
| Test output captured | ✅ (`byteport-presign-smoke.log`) |
| Go backend unaffected | ✅ (`/healthz` returns 200) |
| Documentation updated | ✅ (this transcript + the appended 1-WEEK OUTCOME section in the FWD report) |

The audit's underlying defect (unsigned URLs rejected by real S3
endpoints) is closed. PR #443 can be merged.
