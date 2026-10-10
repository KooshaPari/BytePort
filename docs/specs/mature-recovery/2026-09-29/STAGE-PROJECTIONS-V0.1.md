# BytePort stage projections — mature-contract projections v0.1

Status: **PROVISIONAL / no disposable mini-products**  
Date: 2026-09-30.

## CVP — one exact selected application

Close the selected-app vertical spine for:
- one Git source adapter;
- one manifest schema/version;
- one runnable component;
- one BuildEngine adapter;
- one disposable RuntimeAdapter;
- CLI-primary operation plus at least one machine interface;
- exact observation/stop/restart evidence.

Must use mature identities:
SourceSnapshot, ManifestRevision, Component/ArtifactResolution, BuildArtifact, RuntimeOperation, DeploymentGeneration, ProviderResource, Observation.

Portfolio may be a stub projection, but verified/generated/manual distinction must already exist.

## MVP — useful multi-component local control plane

Adds:
- multiple components;
- prebuilt + source-built artifact paths;
- one provider-managed/non-artifact resource;
- durable reconciliation;
- CLI/Desktop/API common operation;
- realized PortfolioProjection + one Publisher;
- explicit loopback/local network mode.

## Beta — provider/build breadth

Adds:
- second BuildEngine family;
- second RuntimeAdapter/target;
- rollout/update/rollback generations;
- publication retry/idempotency;
- installed desktop lifecycle;
- stronger provenance/SBOM;
- authorized non-loopback mode if justified.

## GA

Requires:
- supported source/build/runtime matrix;
- auth/network threat closure;
- schema migration/update/rollback;
- exact legacy-state handling;
- installed CLI/Desktop/API convergence;
- platform journey evidence;
- portfolio publication security;
- no critical UNKNOWN/false-green lifecycle path.

## Mature

Adds:
- more source/build/runtime/publication adapters;
- richer portfolio/productization;
- provider breadth based on measured value;
- automation/policy breadth without changing identity spine.

## Transition debt rules

- CVP may not use ProjectID as runtime identity.
- CVP may not use mutable branch/tag/image tag as final evidence identity.
- CVP operation journal/schema must be additive and extensible to multi-component generations.
- portfolio stub may omit rich rendering but cannot fabricate verified live facts.
- provider/build adapters are replaceable from the first stage.


## Thesis correction overlay — 2026-09-30

The original projection used the selected-app slice too strongly as product framing.

### Revised CVP
The one-app/one-target slice remains acceptable, but its manifest/domain model MUST already express a DesiredResourceGraph and Target abstraction capable of widening to non-application infrastructure without identity replacement.

### Revised MVP
MVP must include heterogeneous desired resources, minimally:
- one runnable service;
- one provider-managed/non-artifact resource;
- one explicit target;
- reconciliation/update/destruction semantics.

Bare-metal support may be Beta if implementation cost demands it, but the target/resource abstraction must not be cloud-specific.

### Revised Beta
Must prove at least two materially different target families, with bare metal strongly preferred as one because direct user intent specifically expanded BytePort beyond AWS/GCP-style deployment.

### Mature
Mature BytePort covers infrastructure/deployment lifecycle generally through adapters and declarative desired state. Portfolio/productization is an integrated downstream projection, not the mature product boundary.
