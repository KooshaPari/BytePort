# BytePort developer work packages — selected-app spine v0.1

Status: **TIER-A READY / TIER-B MERGE-GATED**  
Date: 2026-09-30.

## BP-WP-A01 — additive domain identity package

Introduce domain types:
- SourceReference / SourceSnapshot;
- ManifestRevision;
- Component / ArtifactResolution;
- BuildOperation / BuildArtifact;
- DeploymentIntent;
- RuntimeOperation / ExternalOperationRef;
- DeploymentGeneration;
- ProviderResource;
- Observation;
- RealizedDeployment;
- PortfolioProjection / PublicationOperation;
- EvidenceReceipt.

No existing route behavior changes.

Acceptance:
- serialization/value-object tests;
- ProjectID and ProviderResourceID type mismatch cannot compile at typed boundary where practical;
- same OperationID/different fingerprint represented as conflict;
- component may be prebuilt/built/non-artifact.

## BP-WP-A02 — additive persistence schema

Add versioned tables/records for A01 identities.

Migration rules:
- existing Project/User/Repository rows remain readable;
- existing DeploymentsJSON imported only as legacy provider-instance observations/references;
- source/artifact/manifest verification remains unknown;
- no legacy row becomes RealizedDeployment without evidence.

Acceptance:
- migrate representative old DB fixture;
- downgrade/rollback behavior documented;
- migration is idempotent;
- no remote provider mutation.

## BP-WP-A03 — common application operation interfaces

Define application services independent of Gin/Tauri:
- ResolveSource;
- ResolveManifest;
- ResolveArtifact;
- PlanDeployment;
- Start/Observe/Reconcile/Stop RuntimeOperation;
- ProjectPortfolio;
- PublishPortfolio.

Initial implementations may be test doubles.

CLI/Desktop/API adapters translate into these interfaces rather than own semantics.

## BP-WP-A04 — BuildEngine interface

Implement BUILD-ENGINE-ADAPTER-CONTRACT:
- capability declaration;
- BuildRequest fingerprint;
- immutable BuildResult;
- provenance/SBOM refs;
- failure/unknown states.

Initial:
- prebuilt verifier fake;
- disposable BuildKit/Dockerfile prototype adapter.

No product-level Docker assumption.

## BP-WP-A05 — RuntimeAdapter interface

Define:
- Plan/Create;
- Observe;
- FindByExternalOperation;
- Stop exact ProviderResource;
- reconcile UNKNOWN;
- provider consistency/capability metadata.

Use fake/disposable adapter first.

## BP-WP-A06 — evidence/trace validator

Require exact source/manifest/artifact/operation/generation/resource/verifier identities for accepted greens and quarantine old generated scorecards.

## BP-WP-B01 — provider stop remediation

Merge gate:
- BP-F03 exact negative retained;
- multi-resource/generation selection semantics;
- no ProjectID fallback;
- unrelated resource survival test.

## BP-WP-B02 — durable operation journal

Merge gate:
- isolated operation-model + duplicate-request receipts;
- restart/lost-response fixture;
- UNKNOWN/RECONCILING semantics.

## BP-WP-B03 — source/manifest integration

Merge gate:
- exact Git ref resolver;
- manifest schema/version decision;
- branch-move + invalid-manifest controls.

## BP-WP-B04 — selected-app BuildEngine integration

Merge gate:
- portable immutable artifact/provenance prototype;
- selected-app live probe;
- wrong-artifact-200 rejection.

## BP-WP-B05 — explicit network mode/auth

Design:
- loopback/local;
- authorized LAN/tailnet if retained;
- external mode only if explicitly supported.

Merge gate:
- session-expiry oracle;
- auth/reachability threat review;
- health/metrics exposure decision.

## BP-WP-B06 — portfolio projection/publisher

Merge gate:
- realized/draft projection rules;
- publication idempotency;
- failure independence;
- API publisher + authorized Git fallback decision.

## BP-WP-B07 — CLI/Desktop/API convergence

Move all projections onto A03 services.

Plan-only result must remain distinguishable from execution.

## Ordering

`A01 → A02/A03 → A04/A05 → A06 → B01/B02/B03/B04/B05/B06/B07`

A02 and A03 can proceed in parallel after A01. A04/A05 can be prototyped independently behind A03.
