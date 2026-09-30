# BytePort forward WBS — mature recovery after handoff gate

Status: ACTIVE  
Date: 2026-09-30.

## Critical path

`W4 identity/recovery evidence → W5 build/runtime prototypes → W6 domain implementation → W7 exact vertical slice → W8 portfolio/productization → W9 lifecycle/security/platform → W10 independent review → W11 spec/design freeze`

## W4 — finish current architecture-risk evidence

### W4.1 BP-F03 exact provider identity
Preserve existing native counterexample; remediation later must pass distinct IDs.

### W4.2 BP-F04 persistence window
Preserve native remote-success/local-failure evidence.

### W4.3 Duplicate deploy
Execute repeated indistinguishable request probe and classify whether current API creates multiple provider operations/projects.

### W4.4 Operation reference model
Execute same-ID/same-fingerprint, changed-fingerprint rejection, restart UNKNOWN, provider-ID preservation.

## W5 — build/runtime architecture experiments

### W5.1 P0 delegated build
Current selected-app fixture: exact source + manifest → local immutable image ID → live app marker.

### W5.2 P1 portable OCI
Export/push artifact; independently verify manifest digest.

### W5.3 P2 provenance
BuildKit provenance/SBOM; verify source/manifest/build-plan binding.

### W5.4 P3 second BuildEngine
CNB/pack or strongest alternative on same source fixture.

### W5.5 RuntimeAdapter fixture
Consume immutable artifact and return provider resources; independently query app marker.

### W5.6 reconciliation
lost response, restart, duplicate retry, provider lookup, compensation, multi-resource.

## W6 — domain implementation

Dependency order:
1. durable IDs/schema/migrations;
2. SourceResolver;
3. manifest validation/content identity;
4. BuildEngine interface;
5. BuildOperation;
6. RuntimeAdapter;
7. RuntimeOperation journal;
8. reconciliation worker;
9. Observation/RealizedDeployment;
10. common application service;
11. CLI/API/desktop projections.

Parallelizable:
- 2/3 after schema;
- 4/6 interfaces;
- UI projections after application service stabilizes.

## W7 — first complete vertical slice

Fixture repository with deterministic marker.

Required chain:
`SourceReference → SourceSnapshot → ManifestRevision → BuildOperation → BuildArtifact digest → RuntimeOperation → ProviderResource → app Observation → RealizedDeployment`.

Negative controls:
- branch moves mid-operation;
- manifest changes;
- wrong artifact;
- mutable tag substitution;
- provider response lost;
- DB write failure;
- duplicate retry;
- same OperationID/different fingerprint;
- restart while APPLYING;
- wrong provider resource stop;
- unrelated resource survives;
- app control plane healthy but selected marker absent.

Exit: exact public-path receipts through CLI and one desktop/API projection.

## W8 — portfolio/productization

### W8.1 Projection schema
Fields classified:
- verified;
- generated;
- manual;
- stale/unknown.

### W8.2 generic API publisher
Template fetch → structured publish → receipt.

### W8.3 retry/reconcile
No duplicate project entries after lost response/retry.

### W8.4 Git fallback
Authorized target repo/path; provenance; no source-repo assumption.

### W8.5 secret boundary
No credentials/private env in projection, prompt, evidence or publication.

### W8.6 deployment independence
Publication failure cannot invalidate healthy deployment.

## W9 — lifecycle/security/platform

- auth actor/project/target/resource;
- local bind/access model;
- OAuth/keyring/secret refs;
- schema migration/recovery;
- install/clean host;
- CLI parity;
- desktop/Tauri packaging;
- platform matrix;
- release/update/rollback;
- uninstall/data retention;
- observability/SLOs;
- accessibility/support.

## W10 — independent adversarial review

Attack:
- source ambiguity;
- TOCTOU source/manifest;
- artifact substitution;
- provider duplicate side effects;
- state loss;
- cross-project stop;
- stale observations;
- secret leakage;
- generated fact hallucination;
- publication duplication;
- interface divergence;
- candidate-owned grader.

## W11 — freeze gate

Require:
- ontology accepted;
- BuildEngine/RuntimeAdapter/operation architecture survived prototypes;
- source→realized-app vertical closed;
- productization contract closed;
- lifecycle/security families resolved;
- bidirectional trace complete;
- no blocking authority conflict;
- every critical obligation has independent negative control;
- exact dependency/provider identities recorded;
- independent review complete.

No requirement-count target and no completion from CI/doc volume alone.
