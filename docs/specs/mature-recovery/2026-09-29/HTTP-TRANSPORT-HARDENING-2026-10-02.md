# HTTP transport safety hardening — 2026-10-02

Scope: BP-WP-B08 transport follow-on only. This does not reopen production routing or authorize infrastructure mutation.

## Source-bound native counterexamples

Reviewed parent: `7335befa55ddf643c47efa0c2fa63fcb9f5cdb84`.
Original transport Git blob: `397e1e2bed951209212d15036bf9fb7642a67fe5`.
Candidate transport Git blob: `ac5f61305bb862276e28a3e0b71dc03c7a966040`.
New regression-test Git blob: `721f7820e2b37d24488ab781c68ba46d56c840b4`.
Updated existing-test Git blob: `fcf3ed16dac25250d000fd7a685a6a3601cd675f`.

The new 16 top-level safety tests produced 12 failures and 4 passes against the original transport. Failures included POST replay through HTTP redirects, redirected observation, no finite client timeout, preflight failures misclassified as post-send ambiguity, unsafe resource path identities, HTTP status mistaken for proof of non-mutation, raw provider-response disclosure, unbounded response reads, nil-body panic, invented observed configuration and stop success without a valid acknowledgement.

The repaired transport passed those 16 tests and the four existing HTTP tests. Ten shuffled repetitions under the Go race detector produced 200 successful top-level test executions. `go vet` passed. The existing positive deploy test now returns its configuration label from the test server; the new missing-label adversary requires that absent provider configuration remain absent rather than be copied from the request.

Native verifier: Go 1.23.2, Linux amd64, `GOTOOLCHAIN=local GOPROXY=off`. Commands: `go test -race -json -count=10 -shuffle=on -timeout=40s .`; `go vet .`.

This was an isolated native compilation of the exact transport source plus exact domain declarations extracted from `nanovms_infrastructure_adapter.go`; providers were local `httptest` servers and injected round-trippers. It was NOT a full-backend test, actual NanoVMS deployment, durable restart demonstration or production qualification. Raw before/after JSONL, hashes and reproducible source-slice files are retained in the conversation's forward-verification bundle.

## Contract changes

The caller-owned HTTP client is copied, redirects are blocked and request duration is bounded. Invalid local requests are distinguished from outcomes unknown after send. Until provider-specific rejection semantics are verified, non-success HTTP responses to mutations remain UNKNOWN; a status code alone is not permission to retry. Response reads have a 1 MiB safety ceiling and provider response bodies are not included in errors. Stop requires a valid acknowledgement. Requested configuration is never fabricated as observed configuration.

The transport's 30-second request and 1 MiB response limits are operational safety ceilings, not product performance acceptance thresholds. They apply only to this experimental transport; `/deploy` wiring is unchanged.

## Remaining gates

The existing `^TestNanoVMSHTTPTransport` CI selector includes all 20 tests by name. Exact remote full-backend CI is still required. B08's earlier reference-model results do not automatically qualify this transport. Adapter propagation of ambiguous stop, real-provider identity/configuration semantics, durable reconciliation, network authority and production destructive intent remain separate gates.
