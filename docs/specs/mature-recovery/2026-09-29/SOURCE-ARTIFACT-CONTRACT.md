# BytePort source → manifest → artifact contract — draft 0.1

Status: DRAFT / architecture decision input, not accepted final schema. Source snapshot `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

## Recovered normative pressure

The mature governance/spec corpus requires:
- a BytePort/NVMS manifest;
- schema validation with loud failure;
- multiple services;
- service fields including runtime/port/source reference;
- selected Git repository plus branch/ref deployment;
- successful deployment exposing live endpoint/status;
- portfolio/productization derived from the deployed project.

The mounted handler currently persists repository metadata but does not resolve repository/ref into an immutable source snapshot, does not parse/validate the manifest, and sends a hard-coded `alpine:latest` SandboxConfig to NanoVMS.

Therefore a correct fix cannot simply replace `alpine:latest` with a guessed image name.

## Required identity sequence

### SourceReference
User-facing mutable request, for example:
- repository provider/owner/name or repository ID;
- requested branch/tag/ref/commit.

### SourceSnapshot
Resolved immutable source:
- provider;
- repository stable identity;
- exact commit SHA/content identity;
- acquisition timestamp;
- authorization/account identity;
- optional archive/content digest.

A branch name alone is not a SourceSnapshot.

### ManifestRevision
Exact manifest bytes/schema version bound to the SourceSnapshot, including:
- manifest path;
- content digest;
- schema version;
- parsed semantic representation;
- validation result.

### BuildPlan
Deterministic or explicitly parameterized interpretation of SourceSnapshot + ManifestRevision for a Target:
- services;
- build method;
- runtime;
- ports;
- build args/environment identities excluding secret values;
- provider/runtime adapter.

### BuildArtifact
Produced or selected immutable runnable subject:
- OCI digest or equivalent immutable artifact ID;
- service identity;
- build provenance;
- source snapshot;
- manifest revision;
- builder/version.

### DeploymentIntent
Desired binding:
`Project + SourceSnapshot + ManifestRevision + BuildArtifact(s) + Target`.

### Operation
Durable state transition identity used for retry/reconciliation.

### ProviderResource
Runtime/provider-assigned IDs returned or later observed.

## Decision fork: who builds?

Three legitimate architectures remain:

### A — BytePort-owned build
BytePort resolves source, validates manifest, invokes build engine, records artifact digest, then asks NanoVMS/provider to run it.

Pros: strongest end-to-end evidence/control.  
Cost: substantial custom build-system burden.

### B — delegated build with verified artifact result
BytePort resolves source/manifest and delegates build to CI/runtime, but requires immutable artifact identity and provenance back before deployment is accepted.

Pros: thinner product, composes established systems.  
Risk: provider API must expose enough identity/provenance.

### C — provider-owned source build
BytePort sends immutable source+manifest identity to a provider capable of source builds; provider returns artifact/resource identity.

Pros: thinnest BytePort engine.  
Risk: current NanoVMS SandboxConfig inspected so far is image-centric and does not establish this contract.

No architecture is selected yet.

## First vertical fixture

Create a tiny fixture repository whose application returns:

`source_commit, manifest_digest, build_nonce/artifact_digest`

Acceptance sequence:
1. request named repo/ref;
2. resolve exact commit;
3. read and validate manifest at that commit;
4. build/select immutable artifact;
5. capture artifact digest;
6. deploy to disposable adapter/runtime;
7. independently query running app;
8. compare running marker to requested SourceSnapshot/ManifestRevision/Artifact;
9. record ProviderResourceID distinct from ProjectID;
10. stop exact provider resource;
11. verify it is gone while unrelated fixture survives.

Negative controls:
- branch moves after request resolution;
- manifest changes without source identity update;
- wrong image returns HTTP 200;
- mutable image tag points elsewhere;
- provider returns resource ID different from project ID;
- build returns no immutable digest;
- provider accepts then response is lost;
- local persistence fails after provider success;
- wrong actor/target;
- stale portfolio projection.

## Implementation mapping

| Contract stage | Frozen implementation |
|---|---|
| SourceReference | repository ID/object accepted by deploy request |
| SourceSnapshot | not found in mounted deploy path |
| ManifestRevision | not found in mounted deploy path |
| BuildPlan | not found in mounted deploy path |
| BuildArtifact | hard-coded mutable `alpine:latest`; not selected project artifact |
| DeploymentIntent | implicit Project + placeholder SandboxConfig |
| Operation | no durable pre-side-effect operation found |
| ProviderResource | returned sandbox ID stored in project deployments map |
| Observation | partial runtime/status surfaces; exact subject binding open |
| Stop | wrong identity natively reproduced |

## Decision gate

Before production remediation of selected-source deployment, choose A/B/C by prototype against:
- NanoVMS actual API capability;
- one established alternative deployment engine;
- integration burden;
- provenance/evidence quality;
- recovery semantics;
- user-visible latency;
- portability;
- maintenance cost.

The winning architecture may make BytePort substantially thinner. That is acceptable.
