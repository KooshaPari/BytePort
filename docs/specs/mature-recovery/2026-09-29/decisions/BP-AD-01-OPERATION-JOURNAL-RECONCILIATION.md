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

## Engine-owned state versus BytePort operation truth

BytePort MUST NOT duplicate a selected IaC/runtime engine's authoritative resource state.

OpenTofu-style engines already own resource-state snapshots, backend locking, state lineage/serial protections and provider-specific resource bindings. Pulumi-style engines similarly own checkpoints/state plus explicit refresh against provider reality. Those mechanisms are stronger and more specialized than a second BytePort resource-state database.

The BytePort journal instead spans **product stages** that no one external engine owns end-to-end:

`SourceSnapshot → ManifestRevision → BuildOperation/BuildArtifact → RuntimeOperation/ProviderResource → Observation → PortfolioProjection/Publication`.

For an engine-backed stage, the BytePort operation stores references such as:
- engine/adapter identity + version;
- engine workspace/stack/state identity where applicable;
- engine operation/run ID;
- immutable input/request fingerprint;
- returned resource/artifact identities;
- last independently observed outcome;
- evidence/provenance pointer.

It does not copy the complete provider state graph unless BytePort itself is the selected engine for that stage.

This keeps recovery composable: BytePort can reconcile its product workflow while the specialized engine reconciles its own resource state.

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
- a selected orchestration engine can durably own the **entire accepted BytePort product operation**, including source/build/runtime/publication identities, so a BytePort cross-stage journal would be pure duplication;
- or mature product intent is narrowed so BytePort never coordinates multi-stage side effects itself.

Do **not** reject this decision merely because one stage has strong engine-owned state; referencing that state is the intended composition.

## Required experiments before freeze

1. provider success + lost response;
2. provider success + local journal/update failure;
3. retry same OperationID/same fingerprint;
4. same OperationID/different fingerprint must reject;
5. provider lookup after restart;
6. compensation success/failure;
7. multiple provider resources for one deployment;
8. stop exact resource while unrelated project/resource survives.
