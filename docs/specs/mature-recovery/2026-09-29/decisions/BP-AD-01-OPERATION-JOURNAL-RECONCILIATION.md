# Architecture decision candidate BP-AD-01 — durable operation journal and reconciliation spine

Status: **PROVISIONAL ACCEPT — pending provider prototype and independent review**  
Date: 2026-09-30.

## Evidence

BP-F03 natively proves ProjectID != ProviderResourceID in the current lifecycle.

BP-F04 natively reproduces a second boundary: provider deployment can succeed while local Project persistence fails and no compensating stop occurs. The external runtime and local database are not one atomic transaction.

The mature source/artifact contract also distinguishes SourceSnapshot, ManifestRevision, BuildArtifact, DeploymentIntent, Operation and ProviderResource.

## Decision candidate

BytePort's deployment spine SHOULD persist a durable `Operation` before initiating externally visible provider side effects.

Minimum operation identity/state:

`OperationID, Principal, ProjectID, DeploymentIntentID, TargetID, AdapterID/version, request fingerprint, state, attempt generation, provider idempotency identity where supported, ProviderResourceID(s), timestamps, last observation, error/uncertainty provenance`.

Candidate state machine:

`PLANNED → AUTHORIZED → APPLYING → {REALIZED | FAILED_BEFORE_SIDE_EFFECT | UNKNOWN} → RECONCILING → {REALIZED | COMPENSATED | FAILED_REQUIRES_INTERVENTION}`

Stopping/removal is a separate operation referencing the exact RealizedDeployment/ProviderResource identity.

## Why reconciliation is primary

Compensation alone is insufficient:
- a timeout/lost response may not reveal whether the provider created a resource;
- provider delete may fail;
- deletion may be undesirable after a successful long build;
- retry semantics differ by provider scope;
- BytePort must preserve evidence/history even if compensation succeeds.

Therefore the durable operation journal is product truth; compensation is one adapter-specific recovery action.

## Adapter contract

Every provider/runtime adapter must declare:
- create/apply request identity;
- provider idempotency semantics and scope;
- lookup/reconciliation mechanism;
- returned ProviderResource identity;
- stop/delete semantics;
- observation/readiness semantics;
- whether compensation is safe;
- uncertainty cases it cannot resolve automatically.

An adapter that cannot reconcile uncertain creation must surface explicit UNKNOWN/needs-intervention state rather than return green.

## Build boundary

The leading architecture remains delegated build + verified immutable BuildArtifact. Build operations can use the same durable Operation model but are distinct from runtime deployment operations.

NanoVMS currently fits the runtime/sandbox adapter boundary, not proven source-build ownership.

## Falsification criteria

Revise this decision if:
- the selected provider engine exposes a stronger durable transaction/state model BytePort can adopt directly without duplicating it;
- all accepted targets provide a trustworthy engine-owned operation journal and BytePort can reference rather than own it;
- or mature product intent is narrowed so BytePort never initiates remote side effects itself.

## Required experiments before freeze

1. provider success + lost response;
2. provider success + local journal/update failure;
3. retry same OperationID/same fingerprint;
4. same OperationID/different fingerprint must reject;
5. provider lookup after restart;
6. compensation success/failure;
7. multiple provider resources for one deployment;
8. stop exact resource while unrelated project/resource survives.
