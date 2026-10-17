# Architecture decision candidate BP-AD-02 — delegated build engine with verified immutable artifact

Status: **PROVISIONAL ACCEPT / prototype required**  
Date: 2026-09-30.

## Context

The mature BytePort horizon requires selected source/manifest to become a deployed application. The mounted implementation instead sends `alpine:latest`.

NanoVMS boundary evidence supports runtime/sandbox ownership, not source-build orchestration. Existing deployment systems already support Git builds or immutable-image deployment.

OCI descriptors standardize digest-based content identity; an artifact digest identifies bytes independently of a mutable tag. Current Coolify documentation demonstrates both source-build deployment and deployment of an already-built image by immutable SHA256 digest. OpenTofu demonstrates a mature pattern for binding desired resource instances to remote object identities and protecting state writes with locking.

## Decision candidate

BytePort owns:
- Project;
- SourceReference resolution to SourceSnapshot;
- ManifestRevision and validation;
- DeploymentIntent;
- BuildOperation identity/provenance;
- immutable BuildArtifact reference;
- RuntimeDeployment Operation;
- ProviderResource binding;
- reconciliation/evidence;
- portfolio projection.

BytePort SHOULD NOT initially own a general-purpose build engine.

Instead it calls a `BuildEngine` adapter:

`Build(SourceSnapshot, ManifestRevision, BuildPlan) -> BuildArtifact + provenance`

The result is unacceptable unless it provides immutable artifact identity appropriate to the runtime, e.g. OCI digest for container-image targets.

## Candidate engines

The adapter may target:
- existing CI that publishes OCI artifacts;
- a deployment/build platform capable of returning/locating immutable image identity;
- a local established builder for CVP;
- another engine demonstrated superior in the prototype.

Coolify is both an alternative product stack and a useful falsification baseline, not automatically a BytePort dependency.

## Runtime boundary

A separate `RuntimeAdapter` consumes BuildArtifact + runtime configuration and returns ProviderResource identities.

NanoVMS is currently modeled here.

BuildEngine and RuntimeAdapter may be implemented by the same external product, but BytePort's ontology does not collapse their identities.

## Acceptance

A deployment cannot become REALIZED merely because:
- build command exited zero;
- mutable image tag exists;
- runtime API returned 200/201;
- health endpoint of the control plane is green.

Acceptance requires an independently verifiable chain:
`SourceSnapshot + ManifestRevision → BuildArtifact digest → ProviderResource → live application observation`.

## Prototype

Use one tiny repository exposing source commit + manifest digest + build/artifact marker.

Run at least:
- delegated build → immutable OCI digest → runtime;
- strongest realistic alternative end-to-end stack.

Measure integration LOC/configuration, elapsed build/deploy time, identity/provenance completeness, retry/recovery behavior, credentials, manual intervention and operating burden.

No architecture winner is declared from feature checklists alone.
