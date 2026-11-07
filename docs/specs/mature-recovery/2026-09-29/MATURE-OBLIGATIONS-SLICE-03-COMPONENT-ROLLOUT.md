# BytePort mature obligations — semantic slice 03: component graph, rollout, external operations

Status: **PROVISIONAL SEMANTIC SLICE / NOT FULL CONTRACT**  
Date: 2026-09-30.

## BP-MO-COMP-001 — deployment intent is a component graph
A ManifestRevision resolves to explicit Components and dependency/relationship edges. One project MAY contain multiple services/jobs/managed resources.

## BP-MO-COMP-002 — artifact resolution is per component
Each Component has ArtifactResolution:
- verified prebuilt BuildArtifact;
- BuildOperation-produced BuildArtifact;
- provider-managed/non-artifact.

No global assumption says every component must build an image.

## BP-MO-COMP-003 — non-artifact resources remain first-class
Managed database/network/storage/provider resources can participate in DeploymentIntent/RealizedDeployment without fake BuildArtifact identities.

## BP-MO-BUILD-006 — multi-platform artifact identity is explicit
BuildArtifact records media kind/platform set. Runtime evidence identifies the exact selected platform manifest/artifact where relevant.

## BP-MO-EXTOP-001 — external durable operation identity is preserved
If BuildEngine/RuntimeAdapter owns a durable operation, BytePort records ExternalOperationRef with engine/provider identity and lookup semantics rather than duplicating its internal state as product truth.

## BP-MO-ROLL-001 — deployment generations are explicit
Every rollout/update/rollback creates or references a DeploymentGeneration with predecessor/supersedes/rollback-of relationships.

## BP-MO-ROLL-002 — multiple generations may coexist
Blue/green/rolling transitions may have current/candidate/retiring ProviderResources simultaneously. Stop/reconcile cannot assume one Project→one runtime.

## BP-MO-ROLL-003 — acceptance selects generation explicitly
RealizedDeployment/current-generation selection requires accepted observation/evidence. New resource creation alone does not make it current.

## BP-MO-REC-004 — eventual provider visibility does not equal absence
After uncertain mutation, provider lookup returning “not found yet” cannot immediately prove the mutation did not happen when the provider's consistency contract permits delayed visibility.

## BP-MO-PORT-011 — projection mode is explicit
PortfolioProjection mode is draft, planned or realized. Draft/planned projections MAY exist before deployment, but unobserved live facts cannot be labeled verified.

## BP-MO-PORT-012 — publication policy controls non-realized projections
Whether draft/planned projections may be public is an explicit publication policy, not an accidental renderer behavior.

## BP-MO-SRC-006 — source adapters may widen without redefining identity
Git is the recovered primary source family. Future local/archive/source-provider adapters may map into SourceSnapshot only through accepted adapters; ontology generality does not claim implementation support.

## Verification backlog
- app + managed database fixture;
- one prebuilt service + one source-built service;
- multi-platform OCI index selection;
- external provider operation lookup after restart;
- blue/green exact-resource stop;
- delayed provider visibility;
- rollback to earlier source/artifact;
- draft portfolio before deployment without fake live claims.
