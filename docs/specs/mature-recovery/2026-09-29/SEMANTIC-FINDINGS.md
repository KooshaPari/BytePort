# BytePort — semantic findings and source-ledger supplement

Source `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`, observed2026-09-29. These are evidence-backed source findings and explicit unresolved interpretations, not fixes or native acceptance runs.

## Additional source receipts

- **BP-S11:** `backend/byteport/models/projects.go`, full, blob `2383312e2102b428840f990361cf2f25f5cd1457`. CURRENT_IMPLEMENTATION. BeforeSave serializes the deployment map and generates a UUID only when absent; it does not substitute the returned provider ID for an existing project UUID. AfterFind reconstructs the stored deployment map. BP-J01/02/05 affected; schema/migration and native persistence tests remain open.
- **BP-S12:** `backend/byteport/routes/pm.go`, full, blob `2ebbc29bebe0ec408ccdd620e34530527f7cda66`. CURRENT_IMPLEMENTATION. addNewProject delegates to DB.Create; removeProject to DB.Delete. No compensating runtime operation or identity rewrite here. BP-J02/05 require crash/target-effect oracles.
- **BP-S13:** targeted conversation retrieval for an accepted local-only pivot around September2026. Found portfolio/governance and desktop packaging discussions, not a user acceptance deleting cloud/portfolio scope. Classify as bounded search result, not proof of exhaustive absence.

## Findings

| ID | Observation and authority | Product/architecture consequence | Falsification/closure required |
|---|---|---|---|
| BP-F01 Mature scope/provenance conflict | Current registry local-only wording cites absent curated prompt corpus; SPEC opening and user-attributed historical/2026 retrieval preserve cloud+portfolio | Cannot freeze ontology, final alternatives or mature contract by copying newer prose | Recover accepted pivot or reconcile narrower stage with mature horizon; inspect original sources |
| BP-F02 Wrong deployed subject | Mounted /deploy constructs alpine:latest/native request; selected repository fields do not select upstream build/image in this handler | Successful sandbox allocation is not successful deployment of selected application | Independent selected-source marker/digest and upstream payload assertion; wrong app returning200 must fail |
| BP-F03 Provider identity mismatch | /deploy stores sandboxResp.ID in deployment map but /terminate sends project.UUID; BP-S11/12 do not normalize them | A valid distinct provider ID is not the ID targeted for stop | Fixture returns different project and provider IDs; verify actual upstream target and surviving foreign resource |
| BP-F04 Uncertain side-effect window | Remote deploy occurs before local project persistence; error paths/pm.go show no compensation here | Lost response or failed DB write can leave runtime state not durably represented | Crash after remote success/before DB save, then restart/retry/reconcile under provider-specific semantics; do not claim global absence of recovery until other modules reviewed |
| BP-F05 Local trust boundary unresolved | main binds 0.0.0.0 with protected and public routes | Local data sovereignty, desktop UI and network reachability are distinct properties | Packet-level supported-mode test; permissions/auth/CORS/credential threat model; not an asserted unauthenticated exploit |
| BP-F06 Documentation generations disagree | backend retirement note vs root/module and registry presign/healthz claims | Source denominator and public compatibility remain unresolved | Complete history/call/mount map; preserve historical sources and authorize supersession |

Trace example: recovered deployment intent BP-S10 ↔ candidate selected-source obligation ↔ BP-J01 ↔ main.go /deploy ↔ DeployProject upstream payload ↔ planned external marker/digest oracle. Current implementation edge is observed source; oracle result is MISSING. That missing edge cannot be filled by a smoke-test name or HTTP200 screenshot.

No test runner, native Go/Tauri/Rust application, real NanoVMS instance or cloud resource was executed in this pass. These findings are not a quantified usability score, complete security assessment, or evidence of remediation.
