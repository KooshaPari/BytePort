# BytePort ontology v1.0 candidate — source-to-realized-product contract

Status: **V1.2 CANDIDATE / direct-user thesis correction incorporated; generalized infrastructure graph expanded**  
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

### Component
A desired service/resource node in the manifest graph. Components may be runnable services, jobs, managed databases or other provider resources.

### ArtifactResolution
Per-component resolution stating whether the component uses a verified prebuilt BuildArtifact, a BuildOperation-produced artifact, or no runnable artifact because the component is provider-managed/non-artifact.

### BuildPlan
Normalized build request independent of builder brand.

### BuildOperation
Durable build transition identity including request fingerprint and BuildEngine adapter/version.

### BuildArtifact
Immutable runnable output identity plus provenance, media kind and supported platform set. OCI digest is canonical for OCI targets where available; mutable tag alone is invalid evidence. Runtime evidence records the exact selected platform artifact where relevant.

### Target
Desired environment/provider/runtime/account/region/platform identity.

### DeploymentIntent
Immutable desired binding of Project + SourceSnapshot + ManifestRevision + BuildArtifact + Target + runtime configuration.

### RuntimeOperation
Durable retry/reconciliation identity established before uncertain externally visible mutation.

### ExternalOperationRef
Exact provider/engine-owned operation identity when the adapter supplies durable state. BytePort references rather than duplicates that state while retaining its cross-stage product Operation.

### ProviderResource
Exact externally assigned/observed resource identity. Never ProjectID by convenience.

### Observation
Fresh independently obtained provider/runtime/application fact with subject and timestamp/provenance.

### DeploymentGeneration
One realized/attempted rollout generation with predecessor/supersedes/rollback-of relations and roles such as desired, current, candidate or retiring. Multiple generations may coexist during blue/green or rolling update.

### RealizedDeployment
Accepted binding between one DeploymentIntent/DeploymentGeneration and the exact ProviderResources whose required component/application identities/readiness have been independently observed. It may include non-artifact managed resources.

### PortfolioProjection
Evidence-derived representation separating verified facts, generated content, manual content and freshness/provenance. Projection mode is explicit, e.g. draft/planned/realized; live verified facts require corresponding observations.

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
21. DeploymentIntent represents a component/resource graph, not one mandatory artifact.
22. Each component has explicit ArtifactResolution; provider-managed resources may have no BuildArtifact.
23. Rollout generations preserve predecessor/supersedes/rollback relations and may coexist.
24. Engine-owned durable operations are represented by ExternalOperationRef rather than copied as BytePort-owned truth.
25. Multi-platform artifacts record media/platform identity and runtime-selected subject.
26. Draft/planned portfolio projections cannot present unobserved live facts as verified.

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


## V1.2 generalized infrastructure extension

Direct user authority establishes heterogeneous infrastructure/deployment lifecycle as the mature thesis.

### DesiredResource
A node in declarative desired state. Kinds may include service, job, host/machine, network, storage, managed service, secret binding or adapter-defined infrastructure resource.

### DesiredResourceGraph
Versioned graph of DesiredResources and dependency/connectivity/placement relationships derived from ManifestRevision.

### Target
A deployment/infrastructure destination or scheduling domain: cloud account/region, bare-metal host/group, local host, VM/MicroVM runtime or other accepted adapter target.

### PlacementConstraint
Desired/policy constraints governing where/how a DesiredResource may be realized.

### ResourcePlan
Adapter-neutral planned mapping from desired graph to build actions and target resource operations. It is not realized state.

### RealizedResource
Generalization of ProviderResource: exact target/provider-owned identity realizing one DesiredResource/generation. Runnable ProviderResource remains a specialization/compatibility projection.

### InfrastructureObservation
Observed state/configuration/health/identity for a RealizedResource with freshness/provenance.

### ReconciliationPlan
Versioned diff/actions required to move observed realized graph toward desired graph, including create/update/replace/delete/no-op/unknown.

### DestructionIntent
Explicit authorized desired-state transition for resource destruction. Local uninstall is not DestructionIntent.

## V1.2 invariants

27. Application components are one DesiredResource family, not the whole infrastructure ontology.
28. Bare-metal/cloud/local targets share lifecycle semantics but may expose different capabilities.
29. Unsupported target/resource capability is explicit; adapters do not silently approximate destructive semantics.
30. ResourcePlan is not realized infrastructure.
31. Every externally visible realized resource has exact target/provider identity where the target supplies one.
32. Reconciliation compares desired and observed graphs and preserves UNKNOWN when observation is insufficient.
33. Destruction requires explicit authorized intent and exact selected resources.
34. BuildArtifact is required only for DesiredResources that actually consume runnable artifacts.
35. Target/provider choice is adapter/policy data, not BytePort product identity.
36. The selected-app vertical slice proves a narrow graph projection; it cannot be used as the mature-scope denominator.
