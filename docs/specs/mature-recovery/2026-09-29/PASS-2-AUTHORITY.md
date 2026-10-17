# BytePort — recovery pass 2: authority reconciliation and identity oracle

Date: 2026-09-29. Analyzed product source remains `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`. Native adversarial-test commit on the draft branch: `934b0456de5509b5e89294c3adaa046e632a2767`.

## Authority reconciliation materially advanced

The local repository contains multiple explicit, internally consistent mature-scope records at the frozen source revision:

- `CHARTER.md` (blob `2b169190fbe952370971e5be1fee22e6add45a01`), last updated 2026-06-12, says BytePort turns one `odin.nvms` into a deployed, portfolio-worthy project, provisions through the separate NVMS runtime, registers endpoints with a portfolio, and treats portfolio as first-class. It says mission/charter changes require Level-4 Executive Authority.
- `ADR.md` (blob `fac766911cbf18fc1102e62ef39add75e7fb73b1`) marks custom NVMS manifest, AWS v1 target, Go+web architecture, LLM-assisted portfolio generation and **CLI-first** interface Accepted.
- `PRD.md` (blob `31b94626815c1d738bbb5c162403ff87a3c14d26`) defines IaC deployment + portfolio UX generation, Git source/ref deployment and CLI deploy/status.
- `FUNCTIONAL_REQUIREMENTS.md` (blob `049086597a3bff2b88f763f8b3dfad1b2c353d12`) contains stable manifest, AWS/Git, portfolio and CLI obligations.
- `USER_JOURNEYS.md` (blob `bba7b2ea1777e5470b44cb03c66dcb11d7410cc2`) defines founder manifest deploy, deployment status and public portfolio render journeys, and explicitly admits the listed smoke tests are not journey tests.

These are stronger than the previous pass's SPEC-only contradiction because they include accepted-marked ADRs and a charter with an explicit authority rule. The September registry local-first/desktop-first interpretation is therefore **not authorized to supersede this mature horizon merely by being newer**. It may represent a CVP/placement evolution or later proposal, but an accepted Level-4 mission change or equivalent user authority has not been recovered.

### Provisional authority resolution

Until contrary authority is found:
1. **Mature product identity:** self-hosted deployment + productization control plane: source/manifest → owned deployment → observable endpoints/state → portfolio projection.
2. **Placement:** BytePort control plane may run locally/on one VM; “local-first” is compatible with a remote deployment target.
3. **Surface conflict:** ADR-005 says CLI-first while current registry says desktop-first. Neither surface is allowed to erase the other until a later accepted decision is recovered. Treat CLI and desktop as projections over the same application contract.
4. **AWS v1 vs future targets:** AWS is accepted v1 target, not necessarily eternal product identity. Multi-cloud remains out of v1.
5. **Portfolio:** first-class mature obligation with explicit opt-out in the charter, not a decorative post-build extra.
6. **NanoVMS:** separate runtime/provider adapter, not BytePort identity.

This resolution is still marked provisional because the original user conversations and any post-June executive mission decision have not been exhaustively recovered.

## Existing journey documents contain implementation-fiction risk

`USER_JOURNEYS.md` names routes and code paths such as `POST /api/deploy`, `backend/nvms`, manifest validation and portfolio route/LLM modules. Current mounted router/source already contradict parts of this. Therefore journey documents are normative/supporting intent where authorized, **not proof those code paths exist or are mounted**. Each step must be traced to actual callers.

Its explicit statement that smoke tests only prove the harness is retained as a useful honesty constraint.

## BP-S14 — native provider-identity adversarial control

A new handler-level Go test `backend/byteport/routes/recovery_provider_identity_test.go` deliberately seeds:
- product project ID = `project-123`;
- persisted provider runtime ID = `provider-sandbox-777`.

It invokes the real `TerminateInstance` handler against an HTTP stub and requires the upstream stop request to target the persisted provider ID. The current inspected handler instead constructs `/v1/stop?id=<project.UUID>`.

This test is an **oracle-only expected failure**; no production fix is included. Equal-ID fixtures are specifically rejected because they can manufacture a false green. GitHub Actions evidence for the exact commit must be tied to a job that actually runs this Go package/test; unrelated green workflows do not qualify.

## Selected-artifact oracle remains deliberately unimplemented

The accepted FR/PRD require Git repository/ref deployment, but the current mounted request contract does not expose enough accepted source/manifest identity to write a correct fix without deciding architecture first. The existing handler's hard-coded `alpine:latest` is sufficient to falsify “selected application deployed,” but the recovery program will not invent an expected image naming convention merely to make a test compile.

The vertical slice must first choose how SourceSnapshot + ManifestRevision resolve to BuildArtifact digest and Target adapter. The oracle then compares requested source/manifest, produced artifact digest, provider request and live application marker.

## SOTA pass-2 attack

Current Coolify documentation materially strengthens the existence-gate alternative: it supports Git repository source, selected branch/commit, multiple build methods, Docker Compose, existing immutable image digests, queued deployment history, logs, domains, health checks, previews/rolling-update paths and CLI deployment. This makes “Git source → running application control plane” commodity/contested, not differentiation.

OCI descriptors give a standard content-digest identity primitive. AWS documents provider-specific idempotency tokens and scope. OpenTofu documents state locking with lock identity. Tauri capabilities provide frontend/webview permission scoping but do not secure a separately bound Go HTTP listener.

BytePort's surviving candidate differentiation is therefore narrower: one coherent developer-owned source→deployment→portfolio experience, potentially across local and cloud targets, with unusually strong evidence/provenance and low operational burden. The custom NVMS DSL/runtime must beat adapting established formats/engines; its existence is not self-justifying.

## Gate delta

- Mature-scope authority conflict: **narrowed substantially**, not fully closed.
- CLI-first vs desktop-first surface authority: OPEN.
- Provider identity bug: native adversarial TEST COMMITTED / CI evidence pending.
- Selected-artifact contract: BLOCKED on architecture/identity decision.
- Journey implementation mapping: OPEN; existing journey docs cannot self-prove reachability.
- Existence gate: harder; Coolify/Compose/IaC alternatives cover more of the core than pass 1 established.
- Completion percentage: null.
