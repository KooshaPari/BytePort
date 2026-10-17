# BytePort ontology v0.1 — source, operation, deployment, and provider identity

Status: **DRAFT / evidence-backed slice, not full ontology freeze**.  
Source snapshot: `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.  
Date: 2026-09-29.

## Purpose

BP-F03 is now natively reproduced: the current terminate handler targeted the BytePort project UUID rather than the persisted provider runtime ID. This proves the need for distinct deployment identities before a remediation patch.

## First-class identities

| Entity | Meaning | Must not be conflated with |
|---|---|---|
| `Principal` | authenticated human/service authority | request-body owner |
| `Project` | durable product/project identity | one deployment/runtime |
| `SourceSnapshot` | immutable selected source identity, e.g. repository + commit/ref resolution | repository display name |
| `ManifestRevision` | exact accepted deployment/productization configuration | source snapshot |
| `BuildArtifact` | immutable produced/pulled runnable artifact identity, preferably digest-backed | mutable image tag |
| `Target` | intended provider/runtime/account/region/environment destination | provider resource |
| `DeploymentIntent` | desired project+source+manifest+target state | execution attempt |
| `Operation` | retry/reconciliation identity for one state transition | project UUID |
| `ProviderResource` | actual provider/runtime resource identity returned/observed externally | project/deployment intent |
| `RealizedDeployment` | accepted binding of intent/artifact to one or more provider resources | provider's HTTP success alone |
| `Observation` | fresh provider/runtime/app state evidence | desired state |
| `PortfolioProjection` | presentation derived from authorized project/deployment evidence | product truth source |
| `EvidenceReceipt` | verifier result bound to exact subject/candidate/configuration | route/test name |

## Core identity chain

`Project → SourceSnapshot → ManifestRevision → BuildArtifact → DeploymentIntent → Operation → ProviderResource → Observation`

A portfolio projection references this chain; it does not replace it.

## Required invariants

1. **Project identity persists across deployments.**
2. **Provider resource identity is externally assigned/observed and must be stored explicitly.**
3. **Stop/delete/reconcile actions target provider resources, not project IDs by convenience.**
4. **Selected-source acceptance requires artifact/application identity, not a successful sandbox allocation.**
5. **Mutable tags cannot serve as final evidence identity when immutable digests are available.**
6. **Desired state and observed state are separate.**
7. **Remote side effects and local persistence are not atomic.** Unknown outcomes require reconciliation.
8. **Operation retry semantics are provider/target scoped.**
9. **Authorization binds principal to project/target/resource, not request-body claims.**
10. **Portfolio facts must carry deployment/source provenance and freshness.**
11. **Local-first control-plane placement does not imply local-only target semantics.**
12. **CLI and desktop are projections over the same application contract unless an authorized product decision says otherwise.**

## Current implementation mapping — first slice

| Ontology relation | Current surface | Current evidence |
|---|---|---|
| principal → project authorization | mounted Gin auth + owner-scoped terminate lookup | source-confirmed |
| project identity creation | `deployment.go` UUID generation | source-confirmed |
| repository selection | `deployRequest.repository_id/repository` persisted to Project | source-confirmed, but not propagated to provider artifact |
| source → artifact | no accepted implementation located in mounted handler | **missing / blocking** |
| manifest → deployment | governance says required; mounted handler does not parse selected manifest | **missing/contradictory** |
| deployment → provider request | hard-coded `alpine:latest` SandboxConfig | source-confirmed placeholder behavior |
| provider resource identity | `sandboxResp.ID` stored in deployments map | source-confirmed |
| provider resource → stop | current handler uses `project.UUID` | **native counterexample reproduced** |
| operation journal | no durable pre-side-effect operation identified in inspected path | open/missing in this slice |
| observation/reconcile | monitoring surfaces exist; exact identity binding incomplete | open |
| portfolio projection | mature intent recovered; named journey implementation paths absent | mature obligation candidate, current implementation missing |

## Native evidence

Candidate `934b0456de5509b5e89294c3adaa046e632a2767`, job `109572633061`:

- fixture ProjectID = `project-123`
- persisted ProviderResourceID = `provider-sandbox-777`
- actual stop target = `project-123`
- expected stop target = `provider-sandbox-777`

Classification: `BP-F03 = NATIVE_COUNTEREXAMPLE_REPRODUCED`.

## Immediate oracle consequences

- remediation must define how one project with zero/one/multiple realized deployments selects a stop target;
- selected-source fixture must prove SourceSnapshot/ManifestRevision → BuildArtifact → live application;
- crash-window experiment must persist Operation before remote side effects or establish equivalent recoverability;
- retry tests must deliberately vary operation ID, parameters, target and provider scope;
- portfolio tests must reject stale/wrong/unverified facts.

## Open decisions before v1.0 ontology freeze

- Does BytePort own build execution or delegate it to provider/CI?
- Is `odin.nvms` the canonical user contract, an import format, or one adapter?
- What engine owns desired-state reconciliation?
- How are multiple resources/services represented inside one realized deployment?
- Is public portfolio rendering owned by BytePort or emitted to a separate product?
- CLI-first vs desktop-first surface priority remains unresolved, but both consume the same domain model.
