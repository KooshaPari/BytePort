# BytePort BuildEngine adapter contract — v0.1

Status: **PROVISIONAL ARCHITECTURE CONTRACT / engine choice remains replaceable**  
Date: 2026-09-30.

This refines BP-AD-02 after current BuildKit and Cloud Native Buildpacks research. It prevents the disposable Docker prototype from accidentally becoming product architecture.

## Product-owned request

`BuildRequest` is a product/domain object:

- `BuildOperationID`
- `SourceSnapshot`
- `ManifestRevision`
- normalized `BuildPlan`
- selected target platform/architecture constraints
- authorized secret-reference set, never raw secret values in evidence
- requested BuildEngine adapter identity/version
- request fingerprint

The fingerprint must be stable enough to reject one OperationID reused for semantically different build input.

## Adapter capabilities

Every BuildEngine adapter declares:

- source forms supported;
- explicit-build-definition support, e.g. Dockerfile;
- convention/detection build support;
- supported target OS/architectures;
- immutable artifact identity mechanism;
- OCI registry/daemon/export modes where applicable;
- provenance/attestation support;
- SBOM support;
- builder/toolchain identity;
- cache identity/scope;
- credential requirements;
- cancellation semantics;
- retry/idempotency semantics;
- observable run/build ID;
- failure/uncertainty states;
- offline/hermetic capabilities and limits.

BytePort uses capability negotiation, not engine-name conditionals throughout product logic.

## Accepted result shape

A BuildEngine success is not merely exit code 0.

`BuildResult` must contain enough evidence to construct:

`BuildArtifact {
  media/runtime kind,
  immutable identity,
  platform,
  source snapshot,
  manifest revision,
  build operation,
  engine + version,
  builder/toolchain identity where available,
  provenance reference/digest,
  SBOM reference/digest where available,
  produced_at,
  raw engine receipt
}`

For OCI output, a registry/manifest digest is preferred over a local mutable tag. A local engine-specific image ID can be architecture-prototype evidence but is insufficient as the final portable runtime artifact identity unless the selected runtime contract explicitly makes it sufficient.

## Initial dispatch policy

1. User supplies immutable prebuilt artifact:
   - verify accepted identity/provenance;
   - skip source build;
   - preserve a no-build BuildOperation receipt.

2. Explicit build definition exists and is authorized:
   - select an explicit-build adapter such as BuildKit-family tooling.

3. No explicit build definition but source matches an accepted convention adapter:
   - select a Cloud Native Buildpacks-family adapter where the target/runtime contract is compatible.

4. Existing external CI/build integration selected:
   - create/observe external BuildOperation;
   - require immutable output identity/provenance before continuation.

5. No adapter can establish the required artifact/evidence contract:
   - refuse/require configuration;
   - never guess a build method and return green.

## Provenance expectations

Current BuildKit documentation supports SLSA-style provenance containing source/VCS, build parameters/environment and materials, and can attach attestations to OCI image metadata.

Current Cloud Native Buildpacks lifecycle exports runnable OCI images and writes export reporting that includes image digest/manifest size for registry output; builder/buildpack/lifecycle identity can be versioned.

These external receipts remain imported assertions until BytePort's verifier independently checks the selected artifact and runtime subject.

## Negative controls

- build exits zero but no immutable artifact identity is returned;
- artifact tag mutates between build and deploy;
- BuildOperationID reused with changed SourceSnapshot/ManifestRevision;
- builder/version changes but evidence reuses old receipt;
- source build uses undeclared secret material and provenance leaks it;
- external build says success while artifact registry lookup fails;
- Buildpacks detection selects an unintended runtime;
- Dockerfile path changed after SourceSnapshot resolution;
- cached build result belongs to a different request fingerprint/platform;
- prebuilt artifact path silently rebuilds source.

## Prototype ladder

### P0 — local viability
Current disposable Docker fixture:
source SHA + manifest digest → local build → local immutable image ID → live probe.

### P1 — portable OCI identity
Push/export to an ephemeral/local registry or OCI layout and verify the manifest digest independently.

### P2 — provenance
Generate BuildKit provenance or equivalent and verify SourceSnapshot/BuildPlan references.

### P3 — second engine
Run the same fixture through a CNB/pack adapter, capturing exported image digest/report and builder/buildpack identities.

### P4 — engine interchangeability
Run the same downstream RuntimeAdapter acceptance against P1/P3 artifacts. Runtime acceptance must not depend on which BuildEngine produced the artifact.

## Decision rule

A BuildEngine earns first-party support because it satisfies the contract with acceptable operating burden for an accepted source family—not because it is fashionable, already present, or used by the prototype.

The mature product may support multiple adapters without becoming a general-purpose build system.
