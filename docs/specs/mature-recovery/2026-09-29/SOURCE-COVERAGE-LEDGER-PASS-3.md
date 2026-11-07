# BytePort source/contract coverage ledger — pass 3 delta

Date: 2026-09-30. Supplements pass 2; denominator remains open.

## Newly resolved/refined

### Ontology
Adversarial review added:
- Component graph;
- per-component ArtifactResolution;
- provider-managed/non-artifact resources;
- DeploymentGeneration with predecessor/supersedes/rollback relations;
- ExternalOperationRef;
- multi-platform artifact identity;
- draft/planned/realized PortfolioProjection modes.

Contract slice 03 derives corresponding obligations.

### BuildEngine
Official BuildKit and Cloud Native Buildpacks research now supports a capability-based adapter contract rather than Docker-specific product architecture.

Candidate dispatch:
- immutable prebuilt → verify;
- explicit Dockerfile/build definition → BuildKit-family;
- supported source without explicit build definition → CNB-family;
- existing CI → external adapter;
- unknown semantics → require configuration.

### Interfaces
Accepted CLI-first ADR remains the current normative priority absent direct supersession; desktop/API are peer projections over one domain Operation.

Current implementation drift is explicitly mapped:
- web path performs real placeholder mutation;
- Tauri deploy is plan-only.

### Dependency/boundary
PhenoInfra exact revision frozen. NanoVMS remains a material RuntimeAdapter boundary whose current canonical source/API revision is unresolved; no fake revision is supplied.

### Evidence
BP-F03/BP-F04 retain exact native receipts.
Duplicate deploy, journal model and delegated-build experiments remain pending/non-evidence because workflow runs were cancelled/queued.

## Still blocking denominator closure

- SourceReference→SourceSnapshot native resolution;
- exact ManifestRevision validation path;
- portable OCI artifact/provenance prototype;
- second BuildEngine comparison;
- operation journal + duplicate retry receipts;
- lost-response/restart/delayed-visibility reconciliation;
- multi-resource/blue-green lifecycle fixture;
- selected-application observation through RuntimeAdapter;
- PortfolioPublisher API + Git fallback;
- security/network/secret boundaries;
- additive schema migration plan and compatibility fixtures;
- installed CLI/Desktop/API convergence;
- packaging/update/uninstall/platform parity;
- fresh independent reviewer.

## Handoff correction

Developer handoff v0.2 separates additive domain/interface work from production lifecycle/build remediation gates. Pending experiments cannot be treated as architecture closure.
