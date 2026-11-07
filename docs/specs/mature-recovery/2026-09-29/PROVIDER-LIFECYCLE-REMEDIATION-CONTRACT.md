# BytePort remediation contract — provider resource lifecycle v0.1

Status: DRAFT / derived from BP-F03 native evidence. This specifies what a future implementation change must satisfy; it is not itself a fix.

## Problem

The current terminate API accepts a BytePort Project UUID. The current implementation then sends that same UUID to NanoVMS. Native evidence proves the realized provider resource may have a different ID.

A naive patch such as “use the first deployment map entry's UUID” is not accepted because the mature product supports or intends multi-service/multi-resource deployments and replacement deployments.

## Required model

### Project
Long-lived user product identity.

### RealizedDeployment
One accepted realization of a DeploymentIntent. Has:
- stable BytePort deployment identity;
- project identity;
- source/manifest/artifact provenance;
- target;
- lifecycle state;
- zero or more provider resources;
- created/observed timestamps.

### ProviderResource
One external runtime/provider object:
- provider/adapter identity;
- provider resource ID;
- service/component role;
- generation;
- lifecycle state;
- last observation/evidence.

## Termination semantics

The public operation must state what is being terminated:

1. **terminate project current deployment** — resolve the project's currently active RealizedDeployment, then stop every owned provider resource required by that deployment according to dependency/order policy;
2. **terminate specific deployment** — stop resources bound to that deployment;
3. **terminate specific provider resource** — lower-level/internal/operator action, explicitly identified and authorized.

A Project UUID is therefore a lookup key into deployment state, not a provider stop ID.

## Mandatory negative controls

- ProjectID != ProviderResourceID.
- One project has two historical deployments; terminate-current must not stop the historical already-stopped resource.
- One realized deployment has two provider resources; both required resources are targeted exactly once.
- Foreign user's project/resource cannot be stopped.
- Stale provider resource ID is surfaced/reconciled, not silently treated as success.
- Provider stop succeeds but local state update fails.
- Provider stop times out after actually succeeding.
- Repeated terminate request is idempotent at the product level.
- Resource generation changes/replacement occurs between lookup and stop.
- Missing deployment mapping cannot fall back to ProjectID.

## Evidence required for remediation acceptance

For each stop request record:
- principal;
- product operation ID;
- project/deployment/resource subject;
- provider adapter;
- exact provider resource ID sent;
- provider response;
- fresh follow-up observation where supported;
- resulting durable state;
- candidate/verifier/configuration identity.

## Transition handling

Existing Project rows may contain:
- no deployment map;
- legacy map keyed by `default`;
- provider IDs created by placeholder alpine deployments;
- ambiguous or stale IDs.

Migration must classify these as legacy/unknown rather than manufacture verified RealizedDeployment history.

## Implementation gate

No production BP-F03 fix is accepted until the implementation can answer:

> Which realized deployment and which exact provider resource(s) does this Project-level termination request authorize us to stop?

If the answer is “whatever ID is convenient,” the ontology defect remains.
