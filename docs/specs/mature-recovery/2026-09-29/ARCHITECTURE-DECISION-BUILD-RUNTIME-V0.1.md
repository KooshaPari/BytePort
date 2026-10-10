# Architecture decision candidate — BytePort build/runtime split v0.1

Status: **PROVISIONAL DECISION / prototype required**  
Date: 2026-09-30.

## Decision under test

BytePort owns:
- Project/Principal;
- SourceReference resolution to immutable SourceSnapshot;
- ManifestRevision validation;
- DeploymentIntent and durable Operation identity;
- evidence/provenance;
- target/provider adapter selection;
- reconciliation and portfolio projection.

BytePort should **delegate build execution to an established build/CI engine where practical**, require immutable BuildArtifact identity/provenance back, and treat NanoVMS as one runtime/sandbox provider rather than as the source/build orchestrator.

## Why this is currently favored

Current mounted BytePort:
- receives repository metadata but does not resolve immutable source;
- does not parse the accepted manifest in the mounted deploy path;
- submits a hard-coded mutable image to NanoVMS;
- receives/stores provider runtime identity;
- natively demonstrates a ProjectID/ProviderResourceID stop defect.

Registry NanoVMS boundary evidence places high-level orchestration/cloud provisioning outside NanoVMS and describes sandbox/VMM/runtime responsibilities. That is consistent with the observed API shape.

Established build/deployment systems already solve large portions of source→artifact transformation. Reimplementing a generic build engine would expand BytePort into commodity infrastructure without current differentiation evidence.

## Alternatives

### A — BytePort-owned build engine
Keep only if prototype shows established engines cannot satisfy manifest/provenance/local-first requirements without greater complexity.

### B — delegated build + verified artifact
**Current leading candidate.** BytePort resolves source/manifest, delegates build, receives immutable artifact digest/provenance, then deploys via runtime adapter.

### C — provider-owned source build
Currently unsupported by inspected NanoVMS boundary/API evidence. Re-open only if direct provider source proves an immutable source+manifest build contract.

## Mandatory prototype

A fixture repository must produce an app exposing:
- exact source commit;
- manifest digest;
- artifact/build identity.

The prototype must:
1. resolve mutable ref once to immutable commit;
2. validate exact manifest bytes/schema;
3. invoke chosen build engine;
4. capture immutable artifact identity;
5. deploy through one runtime adapter;
6. independently query the app;
7. bind provider resource identity;
8. stop that exact resource;
9. preserve operation/evidence history.

## Recovery requirement

BP-F04 tests the current remote-success/local-persistence-failure window. The final architecture must choose explicit durable semantics—journal/reconcile/adopt/compensate—not merely reorder calls and hope external/local state is atomic.

## Remaining gate

Compare candidate B against one realistic integrated alternative on:
- implementation/integration surface;
- provenance;
- recovery;
- target portability;
- local-first UX;
- operating burden;
- maintenance/licensing;
- time-to-live application.

No final technology selection is made by this document.
