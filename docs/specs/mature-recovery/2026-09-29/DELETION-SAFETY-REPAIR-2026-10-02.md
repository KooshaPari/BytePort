# Deletion safety correction — acceptance remains blocked

Reviewed branch: 56dc6ba28a150c8ff8fac7d58f2e7f3a7ce644b9. Adapter blob: 7c87d935bf3406ac92706a1a9a6a754dd34f0874.

The previous NanoVMS adapter advertised DELETE and returned REALIZED after Stop. The lifecycle fixtures reinforced that bug by removing their resource from the mock inventory when Stop was called. Passing those tests established neither actual resource deletion nor live-provider capability. The previous B08 deletion-qualification claim is withdrawn; unrelated bounded source/journal/transport evidence remains separate.

The repair separates NanoVMSDeletionTransport from Stop. The ordinary NanoVMSHTTPTransport does not implement Delete and therefore does not advertise it. Direct DELETE calls on Stop-only transports fail before I/O. A deletion-capable provider must receive exact action/operation/desired/realized/target/provider/sandbox identity and then establish absence through a separate observation. Retained resources, uncertain responses or observation failures remain UNKNOWN. Supplied desired state and lifecycle are validated; CREATE cannot mutate an observe-only/unknown-lifecycle resource or silently select another desired identity. Missing provider attribution and blank operation identity fail closed.

The positive HTTP removal endpoint now exists only in an explicitly named test fixture, under /fixture/delete. It is not an asserted NanoVMS API. Stop leaves a stopped resource observable. The real generic HTTP transport remains deletion-unsupported. The lifecycle test reconstructs in-memory adapter state; it does not demonstrate a durable database/process restart.

## Native local evidence

Toolchain: go1.23.2 linux/amd64. Actual fetched adapter and domain files were verified against Git blob hashes before execution in an isolated stdlib-only package. Transport implementations were test doubles. New 18 named tests: original adapter 15 FAIL / 3 PASS; repaired adapter 18 PASS. Eight retained adapter controls, with the destructive ambiguity fixture corrected to model Delete rather than Stop, also pass. Ten shuffled race-detector repetitions produced 260 passing top-level tests, no skips. go vet passed for the isolated package.

Baseline raw log SHA256: 98fd0cf4d15d6fcbd01342ab72218bcff4851385cb3898fe2fe86d91dc373abe.
Repaired repeated raw log SHA256: ff925224e3021648994abf6e050dfa346703a475697884ce37cac5fb2a8bd429.
Repaired adapter blob: e40881cc2d9d7bf5e11afbe79d714200589d8776.
New safety test blob: 476076972fd2177b994001533db1e25181948ab5.
Retained/repaired adapter test blob: 4dbba38259f6b66c792f8461429eead92e1630bf.

Full repository integration and revised full lifecycle fixtures were NOT_RUN locally. The dedicated CI oracle builds the actual model package at the exact PR head, requires every selected named test, uses the race detector, and records source hashes, command, tools, tested checkout, timestamps and raw logs. Missing/ignored/zero tests, stale checkout and collector errors are never acceptance.

No live /deploy route, real provider, production resource, release or merge authorization is changed. A genuine provider deletion contract and authoritative observation semantics, provenance verification, durable restart and the broader mature-product gates remain open.
