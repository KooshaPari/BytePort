# BytePort developer-agent handoff gate — v0.1

Status: **PARTIAL HANDOFF READY**  
Date: 2026-09-30.

Developer agents MAY implement the bounded domain spine and disposable prototypes below. They MUST NOT claim the mature product architecture/specification is complete.

## Green-to-implement work packages

### BP-DEV-01 — explicit identity/domain records
Introduce domain/storage identities without yet replacing every public surface:
- SourceReference;
- SourceSnapshot;
- ManifestRevision;
- BuildOperation;
- BuildArtifact;
- DeploymentIntent;
- RuntimeOperation;
- ProviderResource;
- RealizedDeployment;
- Observation.

Migration must preserve existing Project/User data.

### BP-DEV-02 — provider lifecycle correctness
Fix wrong-resource termination:
- resolve authorized current RealizedDeployment/ProviderResource;
- never use ProjectID as provider ID;
- handle zero/one/multiple resources explicitly;
- retain exact provider IDs returned/observed.

### BP-DEV-03 — durable operation journal
Implement BP-AD-01 semantics:
- operation + request fingerprint before remote mutation;
- same ID/same fingerprint retry;
- same ID/different fingerprint reject;
- APPLYING across uncertain restart → UNKNOWN/RECONCILING;
- provider lookup/reconcile;
- compensation is adapter-specific, not truth erasure.

### BP-DEV-04 — BuildEngine interface
Implement the capability-based interface, not a Docker-specific product layer.
Initial adapters may include:
- prebuilt immutable artifact verifier;
- BuildKit/Dockerfile prototype.

BuildResult must return immutable identity/provenance fields required by BUILD-ENGINE-ADAPTER-CONTRACT.md.

### BP-DEV-05 — selected source/manifest path
Resolve mutable refs once, content-identify exact manifest, and feed SourceSnapshot/ManifestRevision into BuildOperation.

### BP-DEV-06 — common application operation surface
Create one application/domain deploy operation consumed by:
- CLI;
- desktop/web;
- machine API.

Existing Tauri plan-only UX may remain, but actual execution must converge on the same operation model.

## Prototype-only work packages

### BP-EXP-01 — portable OCI artifact
Upgrade disposable selected-app build from local image ID to independently verified OCI/registry digest + provenance.

### BP-EXP-02 — second BuildEngine
Run the same selected-app fixture through Cloud Native Buildpacks or another serious source-build adapter and compare receipts/operating burden.

### BP-EXP-03 — runtime reconciliation
Provider fixture for:
- success + lost response;
- success + local update failure;
- lookup after restart;
- duplicate retry;
- compensation success/failure;
- multi-resource deployment.

### BP-EXP-04 — PortfolioPublisher
Generic fixture:
- template/API fetch;
- structured publication;
- idempotent retry;
- API failure independent from deployment;
- authorized Git fallback;
- secret-redaction and verified/generated authority separation.

## Not authorized yet

- rebuilding NanoVMS inside BytePort;
- making Docker the sole product build engine;
- restoring historical Demonstrator wholesale;
- treating desktop as a separate deployment implementation;
- deleting CLI-first contract without direct superseding authority;
- adding provider/cloud breadth before one exact vertical slice closes;
- product completion based on route 200, CI green or document count.

## First vertical slice target

One fixture repository must prove:

`SourceReference → SourceSnapshot → ManifestRevision → BuildOperation → immutable BuildArtifact → RuntimeOperation → ProviderResource → application Observation → RealizedDeployment`

Then exercise:
- retry;
- restart;
- stop exact resource;
- one wrong-artifact negative;
- one publication projection.

## Mandatory receipts

Every architecture-sensitive PR identifies:
- ontology/obligation/decision IDs;
- migration/state effects;
- exact candidate and dependency revisions;
- provider/build adapter identities;
- positive + negative oracle runs;
- raw evidence identity;
- unresolved uncertainty.

## Handoff verdict

**READY FOR DEVELOPER AGENTS: YES, bounded BP-DEV-01..06 and BP-EXP-01..04.**

**READY FOR “implement the entire mature BytePort”: NO.**

The next promotion gate is the first complete source→artifact→runtime→observation vertical slice plus reconciliation evidence.
