# BytePort mature obligations — semantic slice 01

Status: **PROVISIONAL SEMANTIC SLICE / NOT FULL CONTRACT**  
Date: 2026-09-30.  
Source snapshot: `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

This slice covers source/artifact/deployment/recovery/portfolio semantics now supported by direct user intent, native evidence and current architecture research. It does not replace existing FR IDs yet.

## BP-MO-SRC-001 — mutable source reference resolves once to immutable snapshot

**Statement.** A deployment request using branch/tag/ref MUST resolve to an immutable SourceSnapshot before build/deployment acceptance.

**Positive acceptance.** requested branch resolves to exact commit; later branch movement does not alter the in-flight intent.
**Counterexample.** build stage re-resolves `main` after it moved and deploys different source.

## BP-MO-MAN-001 — exact manifest revision is validated and bound

**Statement.** The exact manifest bytes/schema used to plan deployment MUST be content-identified, validated and bound to the SourceSnapshot/DeploymentIntent.

**Authority.** Direct user requirement for NVMS manifest parsing/schema validation.
**Counterexample.** manifest changes after source resolution but deployment/evidence still claims old configuration.

## BP-MO-ART-001 — runnable artifact has immutable identity and provenance

**Statement.** A BuildArtifact used for deployment MUST have immutable identity appropriate to its runtime and provenance linking it to SourceSnapshot + ManifestRevision + BuildOperation.

**Counterexample.** `latest` is treated as artifact identity.
**Current gap.** mounted deploy path uses `alpine:latest`.

## BP-MO-BUILD-001 — build execution may be delegated but evidence may not

**Statement.** BytePort MAY delegate build execution to an established BuildEngine, but MUST receive/derive sufficient immutable artifact identity and build provenance to establish the source→artifact edge.

**Counterexample.** external builder returns success without identifying the produced artifact.

## BP-MO-OP-001 — durable operation identity precedes uncertain remote side effects

**Statement.** Before initiating a provider mutation whose result may become uncertain, BytePort MUST establish durable operation identity/state or adopt an equivalent engine-owned durable operation record that BytePort can reference.

**Authority/rationale.** BP-F04 natively reproduces provider success followed by failed local persistence with no compensation.

## BP-MO-OP-002 — uncertainty is first-class

**Statement.** Lost response, timeout, local persistence failure, restart or conflicting provider observation MUST be capable of producing explicit UNKNOWN/RECONCILING state rather than false success or false absence.

## BP-MO-ID-001 — project/deployment/provider identities are distinct

**Statement.** ProjectID, RealizedDeploymentID and ProviderResourceID MUST remain distinct and traceable.

**Native evidence.** BP-F03: project-123 was incorrectly used instead of provider-sandbox-777.

## BP-MO-STOP-001 — stop targets exact realized provider resources

**Statement.** A project-level terminate request MUST resolve an authorized current RealizedDeployment and its exact provider resources; it MUST NOT fall back to ProjectID as provider identity.

**Negative cases.** multiple historical deployments, multi-resource deployment, stale/replaced generation, wrong principal.

## BP-MO-OBS-001 — desired and observed runtime state are separate

**Statement.** REALIZED/healthy claims MUST rely on fresh observation of the intended provider resource/application, not solely on a successful control-plane HTTP response.

## BP-MO-APP-001 — selected application identity is independently observable

**Statement.** Deployment acceptance MUST independently establish that the running application corresponds to the selected SourceSnapshot/ManifestRevision/BuildArtifact.

**Vertical fixture.** `/__byteport_probe` source/manifest/nonce markers.

**Counterexample.** unrelated alpine/nginx-like placeholder responds HTTP 200.

## BP-MO-AUTH-001 — actor/project/target/resource authorization binds before side effect

**Statement.** Provider mutations MUST be authorized from trusted session/domain state for the exact project/target/resource and MUST NOT trust body-supplied ownership.

## BP-MO-PORT-001 — portfolio projection is evidence-derived

**Statement.** BytePort MUST derive portfolio project/list representations from authorized Project + source + realized deployment + fresh observation, separating verified facts from generated copy.

**Authority.** direct November 2024 user intent.

## BP-MO-PUB-001 — portfolio publication is adapter-driven

**Statement.** BytePort MUST support the accepted template/API publication model through a PortfolioPublisher boundary and preserve an authorized Git/server-side fallback path unless later authority supersedes it.

**Counterexample.** unavailable historical Slickport blocks all portfolio export.

## BP-MO-PUB-002 — publication failure is not deployment failure

**Statement.** Deployment, portfolio generation and publication outcomes MUST be independently represented.

**Counterexample.** portfolio API outage causes a healthy deployment to be recorded as not deployed.

## BP-MO-SECRET-001 — projectization never leaks protected configuration

**Statement.** Projection/build/deployment evidence and portfolio payloads MUST exclude secrets/private environment values except through explicitly authorized secret-reference mechanisms.

## BP-MO-PROV-001 — generated content carries authority/provenance

**Statement.** LLM/generated descriptions, mockups and inferred metadata MUST remain distinguishable from verified deployment/source facts and carry generation/review provenance.

## Slice exit criteria

Promote this slice only after:
- delegated-build artifact-chain experiment executes;
- immutable artifact evidence is upgraded from local image ID to accepted runtime/distribution identity for the chosen adapter;
- operation-journal/reconciliation prototype handles lost response/restart;
- selected-app and stop-resource fixtures pass against remediation candidate;
- portfolio publisher/Git-fallback contract is mapped to current/historical surfaces;
- independent review attacks source/artifact/operation/publication interpretations.

No requirement-count target applies.
