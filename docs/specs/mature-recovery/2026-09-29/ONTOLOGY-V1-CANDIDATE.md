# BytePort ontology v1.0 candidate — source-to-realized-product contract

Status: **CANDIDATE FREEZE / architecture-risk experiments still gate final acceptance**  
Date: 2026-09-30.  
Frozen implementation source: `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

This supersedes ontology v0.1 as the working mature-product vocabulary. It does not declare implementation completion.

## Product thesis

BytePort is a local/self-hostable deployment control plane that turns an explicitly selected application source/configuration into a verifiably realized deployment and then, where requested, an evidence-derived portfolio/productization projection.

BytePort owns cross-stage product identity, authorization, operations, reconciliation and evidence. It delegates commodity build/runtime/provider work behind adapters where stronger engines already exist.

CLI, desktop and machine/API surfaces project the same application/domain operations.

## Identity chain

`Principal`
→ authorizes `Project`

`Project`
+ `SourceReference`
→ `SourceSnapshot`

`SourceSnapshot`
+ exact config
→ `ManifestRevision`

`SourceSnapshot + ManifestRevision + BuildPlan`
→ `BuildOperation`
→ `BuildArtifact`

`BuildArtifact + Target + RuntimeConfig`
→ `DeploymentIntent`
→ `RuntimeOperation`
→ `ProviderResource[]`
→ `Observation[]`
→ `RealizedDeployment`

Optional productization:
`Project + SourceSnapshot + RealizedDeployment + Observation`
→ `PortfolioProjection`
→ `PublicationOperation`
→ `PublicationReceipt`

## First-class entities

### Principal
Authenticated human/service authority. Body-supplied owner fields are not authority.

### Project
Durable user product identity spanning multiple source/build/deployment/publication generations.

### SourceReference
User-selected mutable or immutable source selector: repository + branch/tag/ref/commit/path as applicable.

### SourceSnapshot
Immutable resolved source identity. A mutable ref is resolved once per intent.

### ManifestRevision
Exact validated deployment/productization configuration bytes + schema/version/content digest.

### BuildPlan
Normalized build request independent of builder brand.

### BuildOperation
Durable build transition identity including request fingerprint and BuildEngine adapter/version.

### BuildArtifact
Immutable runnable output identity plus provenance. OCI digest is canonical for OCI targets where available; mutable tag alone is invalid evidence.

### Target
Desired environment/provider/runtime/account/region/platform identity.

### DeploymentIntent
Immutable desired binding of Project + SourceSnapshot + ManifestRevision + BuildArtifact + Target + runtime configuration.

### RuntimeOperation
Durable retry/reconciliation identity established before uncertain externally visible mutation.

### ProviderResource
Exact externally assigned/observed resource identity. Never ProjectID by convenience.

### Observation
Fresh independently obtained provider/runtime/application fact with subject and timestamp/provenance.

### RealizedDeployment
Accepted binding between one DeploymentIntent and the exact ProviderResources whose selected application identity/readiness has been independently observed.

### PortfolioProjection
Evidence-derived representation separating verified facts, generated content, manual content and freshness/provenance.

### PublicationOperation
Independent durable operation for API/Git/other PortfolioPublisher output.

### EvidenceReceipt
Verifier result bound to exact source/artifact/operation/resource/candidate/configuration.

## Adapter boundaries

### SourceResolver
Resolves SourceReference → SourceSnapshot.

### BuildEngine
Builds/verifies immutable BuildArtifact. Capability-based; candidate families include BuildKit, Cloud Native Buildpacks, existing external CI, or prebuilt-artifact verification.

### RuntimeAdapter
Consumes immutable BuildArtifact + Target/runtime config and returns/observes ProviderResources. NanoVMS is one historical/current candidate, not product identity.

### ReconciliationAdapter
May be part of RuntimeAdapter; declares idempotency, lookup, uncertainty and compensation semantics.

### PortfolioPublisher
Publishes PortfolioProjection through configured API/template integration or authorized Git/server-side fallback.

## Mature invariants

1. Mutable source resolves once to immutable SourceSnapshot.
2. Exact ManifestRevision is content-identified and validated.
3. BuildArtifact is immutable and provenance-linked.
4. Build success without artifact identity is not acceptance.
5. ProjectID, operation IDs, realized-deployment IDs and provider IDs are distinct.
6. Durable RuntimeOperation exists before uncertain remote side effects, or references an equivalent engine-owned durable operation.
7. Same operation ID + different fingerprint rejects.
8. Lost response/restart can become UNKNOWN/RECONCILING.
9. Desired state and Observation are separate.
10. REALIZED requires independent selected-application observation.
11. Stop/delete resolves exact authorized ProviderResources.
12. Engine-owned resource state is referenced rather than blindly duplicated.
13. BytePort journal spans product stages engines do not own end-to-end.
14. CLI/Desktop/API create/observe the same domain operations.
15. Deployment success is independent from publication success.
16. PortfolioProjection distinguishes verified/generated/manual/stale fields.
17. Secrets are references across evidence/publication boundaries unless explicit use is authorized.
18. Publication retry is idempotent/reconcilable.
19. Local-first control plane does not imply local-only deployment target.
20. Evidence binds exact source, dependency, adapter/provider and candidate identity.

## Current mechanism disposition

| Current/historical surface | Mature disposition |
|---|---|
| mounted Gin backend | candidate application/API projection; must converge on domain operations |
| Tauri deploy planner | retain planning UX only if it projects same Build/Runtime operations |
| hard-coded `alpine:latest` | invalid placeholder; remove from accepted deployment path |
| Project deployments map | insufficient ontology; migrate to explicit RealizedDeployment/ProviderResource model |
| direct NanoVMS HTTP mutation | retain only behind RuntimeAdapter/reconciliation contract |
| historical Demonstrator | recover productization semantics, not implementation ownership |
| portfolio endpoint/API key | retain as PortfolioPublisher configuration, with secret-reference handling |
| Git fallback | mature fallback obligation unless superseded by direct authority |
| generic transport/DAG/codec work | must trace to a mature operation/journey or remain infrastructure/supporting scope |

## Architecture freeze blockers

Final v1.0 acceptance still requires:
- delegated build artifact-chain prototype through portable immutable identity;
- second BuildEngine comparison;
- isolated duplicate-operation/journal native receipts;
- provider success/lost-response/restart reconciliation prototype;
- selected application observation through a real RuntimeAdapter fixture;
- multi-resource stop/reconcile model;
- generic PortfolioPublisher API + Git-fallback experiment;
- independent adversarial review.

Until then this candidate ontology is the canonical vocabulary for further contract decomposition.
